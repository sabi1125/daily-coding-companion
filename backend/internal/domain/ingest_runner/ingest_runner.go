package ingestrunner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository"
	logger "backend/internal/log"
	"backend/internal/tx"
	"backend/internal/util"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-sql-driver/mysql"
	"github.com/invopop/jsonschema"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	gmailoption "google.golang.org/api/option"
	"gorm.io/gorm"
)

var errConcurrentRetry = errors.New("ingest run already recorded by a concurrent request")

func isDuplicateIngestRun(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

type IngestRunner struct {
	uuidGenerator      util.UUIDGenerator
	oauthCfg           *oauth2.Config
	ingestRepository   *repository.IngestRepository
	oauthRepository    *repository.OauthRepository
	problemsRepository *repository.ProblemsRepository
	claudeClient       anthropic.Client
	txManager          tx.Manager
	db                 *gorm.DB
}

func NewIngestRunner(
	uuidGenerator util.UUIDGenerator,
	oauthCfg *oauth2.Config,
	ingestRepository *repository.IngestRepository,
	oauthRepository *repository.OauthRepository,
	problemsRepository *repository.ProblemsRepository,
	claudeClient anthropic.Client,
	txManager tx.Manager,
	db *gorm.DB,
) *IngestRunner {
	return &IngestRunner{
		uuidGenerator:      uuidGenerator,
		oauthCfg:           oauthCfg,
		ingestRepository:   ingestRepository,
		oauthRepository:    oauthRepository,
		problemsRepository: problemsRepository,
		claudeClient:       claudeClient,
		txManager:          txManager,
		db:                 db,
	}
}

func (r *IngestRunner) Ingest(ctx context.Context, userIds []string, retried bool) error {
	logger.Info("ingest started!")

	for _, userId := range userIds {
		logger.Infof("starting ingest for user: %s", userId)
		if err := r.RunForUser(ctx, userId, retried); err != nil {
			logger.Warnf("ingest failed for user %s: %v", userId, err)
		}
	}

	logger.Info("ingest ended!")
	return nil
}

func (r *IngestRunner) RunForUser(ctx context.Context, userId string, retried bool) error {
	now := time.Now().In(util.JST)
	ingestDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, util.JST)
	return r.ingestForUser(ctx, userId, ingestDate, retried)
}

func (r *IngestRunner) ingestForUser(ctx context.Context, userId string, ingestDate time.Time, retried bool) error {
	ingestUuid, err := r.uuidGenerator.NewV7()
	if err != nil {
		return fmt.Errorf("creating ingest id: %w", err)
	}

	alreadyRan, err := r.hasAlreadyRun(ctx, userId, ingestDate, retried)
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
	if userOauth == nil {
		return errors.New("no oauth credentials found for user, despite being in the connected-users list")
	}

	tokenSource := r.oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: userOauth.RefreshToken})
	rawBody := ""
	_, tokenErr := tokenSource.Token()
	if tokenErr != nil {
		var receivedErr *oauth2.RetrieveError
		if errors.As(tokenErr, &receivedErr) && receivedErr.ErrorCode == "invalid_grant" {
			return r.writeFailedIngestRun(ctx, ingestUuid, userId, ingestDate, retried, "refresh token invalid")
		}
		logger.Warnf("problem exchanging refresh token for user %s, asking claude for a fallback problem: %v", userId, tokenErr)
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
		logger.Warnf("claude parse failed for user %s, recording as a failed ingest: %v", userId, parseErr)
		return r.writeFailedIngestRun(ctx, ingestUuid, userId, ingestDate, retried, "claude parse failed: "+parseErr.Error())
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

	txErr := r.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := r.problemsRepository.CreateProblem(ctx, &problem); err != nil {
			return err
		}
		if err := r.ingestRepository.CreateIngestWithErr(ctx, entities.IngestRuns{
			IngestRunId: ingestUuid,
			UserId:      userId,
			ProblemId:   &problemId,
			Status:      "success",
			Retried:     retried,
			IngestDate:  ingestDate,
		}); err != nil {
			if isDuplicateIngestRun(err) {
				return errConcurrentRetry
			}
			return err
		}
		return nil
	})
	if errors.Is(txErr, errConcurrentRetry) {
		logger.Infof("ingest run for user %s already recorded by a concurrent request", userId)
		return nil
	}
	if txErr != nil {
		return fmt.Errorf("saving problem and ingest run: %w", txErr)
	}

	return nil
}

func (r *IngestRunner) hasAlreadyRun(ctx context.Context, userId string, ingestDate time.Time, retried bool) (bool, error) {
	ingest, err := r.ingestRepository.GetIngestByUserId(ctx, userId, ingestDate, retried)
	if err != nil {
		return false, err
	}
	return len(ingest) > 0, nil
}

func (r *IngestRunner) writeFailedIngestRun(ctx context.Context, ingestUuid, userId string, ingestDate time.Time, retried bool, errMsg string) error {
	if err := r.ingestRepository.CreateIngestWithErr(ctx, entities.IngestRuns{
		IngestRunId: ingestUuid,
		UserId:      userId,
		Status:      "failed",
		Error:       &errMsg,
		Retried:     retried,
		IngestDate:  ingestDate,
	}); err != nil {
		if isDuplicateIngestRun(err) {
			return nil
		}
		return fmt.Errorf("saving failed ingest run: %w", err)
	}
	return nil
}

type parsedProblem struct {
	Title        *string `json:"title" jsonschema:"required,description=A short descriptive title you write for the problem"`
	ProblemText  *string `json:"problem_text" jsonschema:"required,description=The problem statement; verbatim from the email if one was found there — otherwise one you write yourself. Only the final problem statement itself — never your own reasoning, corrections, or verification work"`
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

func (r *IngestRunner) parseWithClaude(ctx context.Context, rawBody string) (parsedProblem, error) {
	var prompt string
	if rawBody == "" {
		prompt = "No Daily Coding Problem email was found for today. Invent an original " +
			"coding problem yourself — similar style and scope to a typical technical " +
			"interview question — so there's still something to solve today. Work out " +
			"the problem and any example cases in your own thinking first — problem_text " +
			"must contain only the final, clean problem statement, never your reasoning, " +
			"corrections, or verification narration. Call " + extractProblemTool +
			" with problem_text/title/algorithm_tag/difficulty for your invented problem, " +
			"and found_in_email: false."
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
