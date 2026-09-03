package interactor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend/internal/domain/entities"
	ingestInputPort "backend/internal/domain/ingest_runner"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/util"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type ProblemsInteractor struct {
	problemsRepository inputport.ProblemsRepositoryInputPort
	ingestRunner       ingestInputPort.IngestRunnerInputPort
	ingestRepository   inputport.IngestRepositoryInputPort
	settingRepository  inputport.SettingsRepositoryInputPort
	claudeClient       anthropic.Client
}

func NewProblemsInteractor(
	problemsRepository inputport.ProblemsRepositoryInputPort,
	ingestRunner ingestInputPort.IngestRunnerInputPort,
	ingestRepository inputport.IngestRepositoryInputPort,
	settingRepository inputport.SettingsRepositoryInputPort,
	claudeClient anthropic.Client,
) *ProblemsInteractor {
	return &ProblemsInteractor{
		problemsRepository: problemsRepository,
		ingestRunner:       ingestRunner,
		ingestRepository:   ingestRepository,
		settingRepository:  settingRepository,
		claudeClient:       claudeClient,
	}
}

func (interactor *ProblemsInteractor) GetProblems(ctx context.Context, userId string, status entities.ProblemStatus, difficulty entities.ProblemDifficulty) (problems []entities.Problems, err error) {
	logger.Info("ProblemInteractor: GetProblems")

	problems, err = interactor.problemsRepository.GetProblems(ctx, userId, string(status), string(difficulty))
	if err != nil {
		return
	}
	return
}

func (interactor *ProblemsInteractor) GetProblemDetails(ctx context.Context, userId string, problemId string) (problem entities.Problems, err error) {
	logger.Info("ProblemsInteractor: GetProblemDetails")

	problem, err = interactor.problemsRepository.GetProblemDetails(ctx, userId, problemId)
	if err != nil {
		return
	}

	return
}

func (interactor *ProblemsInteractor) GetTodaysProblem(ctx context.Context, userId string) (problem entities.Problems, err error) {
	logger.Info("ProblemsInteractor: GetTodaysProblem")
	todaysDate := util.NewTimeProvider().TodaysDate()

	problem, err = interactor.problemsRepository.GetTodaysproblem(ctx, userId, todaysDate)
	if err != nil {
		return
	}
	if err = verifyProblemOwnership(problem, userId); err != nil {
		problem = entities.Problems{}
		return
	}
	if problem.ProblemId != "" {
		return
	}

	alreadyRetried, getErr := interactor.ingestRepository.GetIngestByUserId(ctx, userId, todaysDate, true)
	if getErr != nil {
		err = getErr
		return
	}
	if len(alreadyRetried) > 0 {
		err = response.NewNoProblemToday(errors.New("retry already used up for today"))
		return
	}

	if ingestErr := interactor.ingestRunner.RunForUser(ctx, userId, true); ingestErr != nil {
		err = ingestErr
		return
	}

	problem, err = interactor.problemsRepository.GetTodaysproblem(ctx, userId, todaysDate)
	if err != nil {
		return
	}
	if err = verifyProblemOwnership(problem, userId); err != nil {
		problem = entities.Problems{}
		return
	}
	if problem.ProblemId == "" {
		err = response.NewNoProblemToday(errors.New("ingest did not produce a problem for today"))
	}
	return
}

func verifyProblemOwnership(problem entities.Problems, userId string) error {
	if problem.ProblemId != "" && problem.UserId != userId {
		return response.NewUnauthorized(errors.New("today's problem does not belong to the caller"))
	}
	return nil
}

func aiHelpIsComplete(aiHelp entities.AIHelp) bool {
	return aiHelp.Concept != "" &&
		aiHelp.Nudge != "" &&
		aiHelp.Approach != "" &&
		aiHelp.Walkthrough != ""
}

func (interactor *ProblemsInteractor) GetAIHelp(ctx context.Context, userId string, problemId string) (aiHelp entities.AIHelp, err error) {
	logger.Info("ProblemInteractor: GetAIHelp")

	problem, err := interactor.problemsRepository.GetProblemDetails(ctx, userId, problemId)
	if err != nil {
		return
	}

	if problem.ProblemId == "" {
		err = response.NewProblemNotFound(errors.New("problem not found"))
		return
	}

	if problem.AiHelp != nil {
		err = json.Unmarshal([]byte(*problem.AiHelp), &aiHelp)
		if err != nil {
			err = response.NewInternalError(err)
		}
		return
	}

	setting, getSettingErr := interactor.settingRepository.GetUserSetting(ctx, userId)
	if getSettingErr != nil {
		err = getSettingErr
		return
	}

	aiHelp, err = interactor.getAIHelpFromClaude(ctx, problem, setting)
	if err != nil {
		return
	}

	marshalledAiHelp, marshalErr := json.Marshal(aiHelp)
	if marshalErr != nil {
		err = response.NewInternalError(marshalErr)
		return
	}

	if updateErr := interactor.problemsRepository.UpdateProblemWithAIHelp(ctx, userId, problemId, string(marshalledAiHelp)); updateErr != nil {
		err = updateErr
		return
	}

	return
}

func (interactor *ProblemsInteractor) getAIHelpFromClaude(ctx context.Context, problem entities.Problems, setting entities.Settings) (aiHelp entities.AIHelp, err error) {
	problemText := problem.RawProblem
	if problem.ProblemText != nil {
		problemText = *problem.ProblemText
	}

	prompt := fmt.Sprintf(
		`A user is stuck on the following coding problem and clicked "Get Help":

%s

Respond with structured help for this problem, filling in exactly these four parts:

- concept: explain the underlying concept/technique this problem relies on, not just the specific problem. Always include this, even if the user seems close to solving it.
- nudge: a short question or hint that points the user toward the right idea without stating it outright.
- approach: describe the strategy/algorithm to use, in plain language.
- walkthrough: a prose talk-through of how to apply that approach to this problem's inputs, step by step.

Do not include a full working solution or runnable code anywhere in the response, unless the user's saved preference below explicitly asks for code examples — the default goal is to help the user solve it themselves, not to solve it for them.

Every field is plain text, not markdown — no code fences (no triple backticks), no headers, no bullet lists. If code is included, write it as plain lines of text within the string.`,
		problemText,
	)

	if setting.GetHelpPreferences != nil && *setting.GetHelpPreferences != "" {
		prompt += fmt.Sprintf(
			"\n\nThe user has also saved this preference for how they'd like help delivered: %s\nApply it on top of everything above — it can steer tone, depth, format, or ask for code examples, but it can never remove the concept explanation or any of the four required parts, and any code must still be plain text, not markdown.",
			*setting.GetHelpPreferences,
		)
	}

	extractHelpTool := "get_ai_help"

	schema, err := toolInputSchema(entities.AIHelp{})
	if err != nil {
		return
	}

	message, err := interactor.claudeClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: []anthropic.ToolUnionParam{
			anthropic.ToolUnionParamOfTool(schema, extractHelpTool),
		},
		ToolChoice: anthropic.ToolChoiceParamOfTool(extractHelpTool),
	})
	if err != nil {
		err = response.NewInternalError(errors.New("failed to get help from claude api"))
		return
	}

	found := false
	for _, block := range message.Content {
		toolUse := block.AsToolUse()
		if toolUse.Name != extractHelpTool {
			continue
		}
		found = true

		if unmarshalErr := json.Unmarshal(toolUse.Input, &aiHelp); unmarshalErr != nil {
			err = response.NewInternalError(unmarshalErr)
			return
		}
	}

	if !found {
		err = response.NewInternalError(errors.New("claude did not call the get_ai_help tool"))
		return
	}

	if !aiHelpIsComplete(aiHelp) {
		err = response.NewInternalError(errors.New("claude returned an incomplete ai help response"))
		return
	}

	return
}

func toolInputSchema(v any) (schema anthropic.ToolInputSchemaParam, err error) {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties:  false,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
	}

	raw, marshalErr := json.Marshal(reflector.Reflect(v))
	if marshalErr != nil {
		err = response.NewInternalError(marshalErr)
		return
	}

	if unmarshalErr := json.Unmarshal(raw, &schema); unmarshalErr != nil {
		err = response.NewInternalError(unmarshalErr)
		return
	}
	return
}
