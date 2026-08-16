package ingest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/internal/domain/entities"
	"backend/internal/domain/repository"
	"backend/internal/infrastructure"
	logger "backend/internal/log"
	"backend/internal/tx"
	"backend/internal/util"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/invopop/jsonschema"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	gmailoption "google.golang.org/api/option"
	"gorm.io/gorm"
)

var jst = time.FixedZone("JST", 9*60*60)

type runner struct {
	uuidGenerator      util.UUIDGenerator
	oauthCfg           *oauth2.Config
	ingestRepository   *repository.IngestRepository
	oauthRepository    *repository.OauthRepository
	problemsRepository *repository.ProblemsRepository
	claudeClient       anthropic.Client
	txManager          tx.Manager
}

func Ingest(db *gorm.DB) error {
	logger.Info("ingest started!")

	ctx := context.Background()
	googleCfg := config.LoadGoogleConfigFromEnv()
	claudeCfg := config.LoadClaudeConfigFromEnv()

	r := &runner{
		uuidGenerator:      util.NewUUIDGenerator(),
		oauthCfg:           config.LoadOauthConfig(googleCfg),
		ingestRepository:   repository.NewIngestRepository(db),
		oauthRepository:    repository.NewOauthRepository(db),
		problemsRepository: repository.NewProblemsRepository(db),
		claudeClient:       anthropic.NewClient(option.WithAPIKey(claudeCfg.APIKey)),
		txManager:          infrastructure.NewTransactionManager(db),
	}

	now := time.Now().In(jst)
	ingestDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)

	userRepository := repository.NewUsersRepository(db)
	userIds, err := userRepository.GetAllUserIds(ctx)
	if err != nil {
		return err
	}

	for _, userId := range userIds {
		logger.Infof("starting ingest for user: %s", userId)
		if err := r.ingestForUser(ctx, userId, ingestDate); err != nil {
			logger.Warnf("ingest failed for user %s: %v", userId, err)
		}
	}

	return nil
}

func (r *runner) ingestForUser(ctx context.Context, userId string, ingestDate time.Time) error {
	ingestUuid, err := r.uuidGenerator.NewV7()
	if err != nil {
		return fmt.Errorf("creating ingest id: %w", err)
	}

	alreadyRan, err := r.hasAlreadyRun(ctx, userId, ingestDate)
	if err != nil {
		return fmt.Errorf("checking existing ingest run: %w", err)
	}
	if alreadyRan {
		logger.Infof("ingest already ran for today for user %s. skipping...", userId)
		return nil
	}

	userOauth, err := r.oauthRepository.FindUserByUserId(ctx, userId)
	if err != nil {
		return fmt.Errorf("getting oauth credentials: %w", err)
	}

	// Only a dead refresh token is a real failure — the user has to
	// reconnect Gmail themselves, so it's never papered over with a
	// fallback problem (docs/DECISIONS.md D1: reauth must stay visible,
	// never a silent failure). Every other way of not getting today's real
	// email (transient exchange error, Gmail unreachable, no email found)
	// just falls through to Claude generating one instead, with rawBody
	// left empty.
	tokenSource := r.oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: userOauth.RefreshToken})
	rawBody := ""
	if _, tokenErr := tokenSource.Token(); tokenErr != nil {
		if err := r.handleTokenExchangeFailure(ctx, tokenErr, ingestUuid, userId, ingestDate); err != nil {
			return err
		}
	} else if fetched, fetchErr := fetchTodaysEmail(ctx, tokenSource); fetchErr != nil {
		logger.Infof("could not fetch today's email for user %s (%v), asking claude for a fallback problem", userId, fetchErr)
	} else {
		rawBody = fetched
	}

	parsed, parseErr := r.parseWithClaude(ctx, rawBody)
	needsReviewFlag := parseErr != nil ||
		parsed.Title == nil ||
		parsed.ProblemText == nil ||
		parsed.FoundInEmail == nil ||
		!*parsed.FoundInEmail
	if parseErr != nil {
		logger.Warnf("claude parse failed for user %s, flagging for review: %v", userId, parseErr)
	} else if parsed.FoundInEmail != nil && !*parsed.FoundInEmail {
		logger.Infof("no problem found in today's email for user %s, using claude's fallback problem", userId)
	}

	problemId, err := r.uuidGenerator.NewV7()
	if err != nil {
		return fmt.Errorf("creating problem id: %w", err)
	}

	problem := entities.Problems{
		ProblemId:       problemId,
		UserId:          userId,
		RawProblem:      rawBody,
		Title:           parsed.Title,
		ProblemText:     parsed.ProblemText,
		AlgorithmTag:    parsed.AlgorithmTag,
		Difficulty:      parsed.Difficulty,
		NeedsReviewFlag: needsReviewFlag,
	}

	err = r.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := r.problemsRepository.CreateProblem(ctx, &problem); err != nil {
			return err
		}
		return r.ingestRepository.CreateIngestWithErr(ctx, entities.IngestRuns{
			IngestRunId: ingestUuid,
			UserId:      userId,
			ProblemId:   &problemId,
			Status:      "success",
			Retried:     false,
			IngestDate:  ingestDate,
		})
	})
	if err != nil {
		return fmt.Errorf("saving problem and ingest run: %w", err)
	}

	return nil
}

func (r *runner) hasAlreadyRun(ctx context.Context, userId string, ingestDate time.Time) (bool, error) {
	ingest, err := r.ingestRepository.GetIngestByUserId(ctx, userId, ingestDate, false)
	if err != nil {
		return false, err
	}
	return len(ingest) > 0, nil
}

// handleTokenExchangeFailure is the one place ingestForUser can still return
// a hard failure. A dead refresh token (invalid_grant) writes a failed
// ingest_runs row and returns an error, aborting this user's run. Any other
// exchange error isn't fatal — same non-fatal treatment as Gmail being
// unreachable, so it's just logged here and the caller proceeds with an
// empty rawBody.
func (r *runner) handleTokenExchangeFailure(ctx context.Context, tokenErr error, ingestUuid, userId string, ingestDate time.Time) error {
	var receivedErr *oauth2.RetrieveError
	if errors.As(tokenErr, &receivedErr) && receivedErr.ErrorCode == "invalid_grant" {
		return r.writeFailedIngestRun(ctx, ingestUuid, userId, ingestDate, "refresh token invalid")
	}

	logger.Warnf("problem exchanging refresh token for user %s, asking claude for a fallback problem: %v", userId, tokenErr)
	return nil
}

func (r *runner) writeFailedIngestRun(ctx context.Context, ingestUuid, userId string, ingestDate time.Time, errMsg string) error {
	if err := r.ingestRepository.CreateIngestWithErr(ctx, entities.IngestRuns{
		IngestRunId: ingestUuid,
		UserId:      userId,
		Status:      "failed",
		Error:       &errMsg,
		Retried:     false,
		IngestDate:  ingestDate,
	}); err != nil {
		return fmt.Errorf("saving failed ingest run: %w", err)
	}
	return nil
}

type parsedProblem struct {
	Title        *string `json:"title" jsonschema:"required,description=A short descriptive title you write for the problem"`
	ProblemText  *string `json:"problem_text" jsonschema:"required,description=The problem statement; verbatim from the email if one was found there — otherwise one you write yourself"`
	AlgorithmTag *string `json:"algorithm_tag" jsonschema:"required,description=The primary algorithm/data-structure it exercises; your own assessment"`
	Difficulty   *string `json:"difficulty" jsonschema:"required,description=One of Easy/Medium/Hard — your own assessment of the problem's complexity"`
	FoundInEmail *bool   `json:"found_in_email" jsonschema:"required,description=True only if problem_text came from the email itself; false if you had to invent a problem because the email had none"`
}

const extractProblemTool = "extract_problem"

func toolInputSchema(v any) anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties:  false,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
	}

	raw, err := json.Marshal(reflector.Reflect(v))
	if err != nil {
		panic(fmt.Sprintf("marshaling generated schema for %T: %v", v, err))
	}
	var schema anthropic.ToolInputSchemaParam
	if err := json.Unmarshal(raw, &schema); err != nil {
		panic(fmt.Sprintf("unmarshaling generated schema for %T: %v", v, err))
	}
	return schema
}

func (r *runner) parseWithClaude(ctx context.Context, rawBody string) (parsedProblem, error) {
	var prompt string
	if rawBody == "" {
		prompt = "No Daily Coding Problem email was found for today. Invent an original " +
			"coding problem yourself — similar style and scope to a typical technical " +
			"interview question — so there's still something to solve today. Call " +
			extractProblemTool + " with problem_text/title/algorithm_tag/difficulty for " +
			"your invented problem, and found_in_email: false."
	} else {
		prompt = "This is a Daily Coding Problem newsletter email. Usually the problem " +
			"statement itself is in the email — there is no title, algorithm tag, or " +
			"difficulty printed anywhere in it, and none of the sponsor/subscribe/" +
			"forward/snooze/unsubscribe boilerplate or the \"asked by <company>\" line " +
			"is part of the problem. Call " + extractProblemTool + " with:\n" +
			"- problem_text: the problem statement, verbatim, with all of the above noise removed\n" +
			"- title: a short descriptive title you write for it (not present in the email)\n" +
			"- algorithm_tag: the primary algorithm/data-structure it exercises, your own assessment\n" +
			"- difficulty: Easy, Medium, or Hard, your own assessment of the problem's complexity\n" +
			"- found_in_email: true\n\n" +
			"If the email does not actually contain a coding problem (e.g. it's an " +
			"announcement or a skipped-day notice), invent an original coding problem " +
			"of similar style yourself instead, so there's still something to solve — " +
			"but set found_in_email: false in that case, so the fallback stays " +
			"distinguishable from real content.\n\n" + rawBody
	}

	message, err := r.claudeClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: []anthropic.ToolUnionParam{
			anthropic.ToolUnionParamOfTool(toolInputSchema(parsedProblem{}), extractProblemTool),
		},
		ToolChoice: anthropic.ToolChoiceParamOfTool(extractProblemTool),
	})
	if err != nil {
		return parsedProblem{}, fmt.Errorf("calling claude: %w", err)
	}

	for _, block := range message.Content {
		toolUse := block.AsToolUse()
		if toolUse.Name != extractProblemTool {
			continue
		}
		var result parsedProblem
		if err := json.Unmarshal(toolUse.Input, &result); err != nil {
			return parsedProblem{}, fmt.Errorf("decoding claude's tool call: %w", err)
		}
		return result, nil
	}

	return parsedProblem{}, errors.New("claude did not call " + extractProblemTool)
}

func fetchTodaysEmail(ctx context.Context, tokenSource oauth2.TokenSource) (string, error) {
	gmailClient, err := gmail.NewService(ctx, gmailoption.WithTokenSource(tokenSource))
	if err != nil {
		return "", fmt.Errorf("creating gmail client: %w", err)
	}

	res, err := gmailClient.Users.Messages.List("me").
		Q("from:founders@dailycodingproblem.com newer_than:1d").
		MaxResults(1).
		Do()
	if err != nil {
		return "", fmt.Errorf("listing messages: %w", err)
	}
	if len(res.Messages) == 0 {
		return "", errors.New("no daily coding problem email found")
	}

	message, err := gmailClient.Users.Messages.Get("me", res.Messages[0].Id).Do()
	if err != nil {
		return "", fmt.Errorf("getting message: %w", err)
	}

	return extractPlainText(message.Payload), nil
}

func extractPlainText(part *gmail.MessagePart) string {
	if part.MimeType == "text/plain" && part.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(decoded)
		}
	}
	for _, child := range part.Parts {
		if text := extractPlainText(child); text != "" {
			return text
		}
	}
	return ""
}
