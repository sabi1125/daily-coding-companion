package interactor

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"backend/internal/domain/entities"
	ingestRunnerMock "backend/internal/domain/ingest_runner/mock"
	inputportMock "backend/internal/domain/repository/inputport/mock"
	"backend/internal/response"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type problemsInteractorMocks struct {
	problems *inputportMock.MockProblemsRepositoryInputPort
	ingest   *inputportMock.MockIngestRepositoryInputPort
	settings *inputportMock.MockSettingsRepositoryInputPort
	runner   *ingestRunnerMock.MockIngestRunnerInputPort
}

func newTestProblemsInteractor(ctrl *gomock.Controller) (*ProblemsInteractor, problemsInteractorMocks) {
	mocks := problemsInteractorMocks{
		problems: inputportMock.NewMockProblemsRepositoryInputPort(ctrl),
		ingest:   inputportMock.NewMockIngestRepositoryInputPort(ctrl),
		settings: inputportMock.NewMockSettingsRepositoryInputPort(ctrl),
		runner:   ingestRunnerMock.NewMockIngestRunnerInputPort(ctrl),
	}
	interactor := NewProblemsInteractor(mocks.problems, mocks.runner, mocks.ingest, mocks.settings, anthropic.Client{})
	return interactor, mocks
}

func TestProblemsInteractor_GetProblems(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testProblems := []entities.Problems{{ProblemId: "p-1", Status: "Open"}}
	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name        string
		status      entities.ProblemStatus
		difficulty  entities.ProblemDifficulty
		prepareFunc func(mpr *inputportMock.MockProblemsRepositoryInputPort)
		wantedError error
	}{
		{
			name:       "success",
			status:     entities.ProblemStatus("All"),
			difficulty: "",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "All", "").Return(testProblems, nil)
			},
		},
		{
			name:       "forwards a specific status filter",
			status:     entities.ProblemStatus("Solved"),
			difficulty: "",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "Solved", "").Return(testProblems, nil)
			},
		},
		{
			name:       "fails to get problems",
			status:     entities.ProblemStatus("All"),
			difficulty: "",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "All", "").Return(nil, errRepo)
			},
			wantedError: errRepo,
		},
		{
			name:       "fails to get problems with difficulty",
			status:     entities.ProblemStatus("All"),
			difficulty: "Hard",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "All", "Hard").Return(nil, errRepo)
			},
			wantedError: errRepo,
		},
		{
			name:       "success with difficulty",
			status:     entities.ProblemStatus("All"),
			difficulty: "Medium",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "All", "Medium").Return(testProblems, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			interactor, mocks := newTestProblemsInteractor(ctrl)
			tt.prepareFunc(mocks.problems)

			problems, err := interactor.GetProblems(ctx, testUserId, tt.status, tt.difficulty)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testProblems, problems)
		})
	}
}

func TestProblemsInteractor_GetProblemDetails(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testProblemId := "test-problem-id"
	testProblem := entities.Problems{ProblemId: testProblemId}
	errNotFound := response.NewProblemNotFound(errors.New("not found"))

	tests := []struct {
		name        string
		prepareFunc func(mpr *inputportMock.MockProblemsRepositoryInputPort)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).Return(testProblem, nil)
			},
		},
		{
			name: "problem not found",
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).Return(entities.Problems{}, errNotFound)
			},
			wantedError: errNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			interactor, mocks := newTestProblemsInteractor(ctrl)
			tt.prepareFunc(mocks.problems)

			problem, err := interactor.GetProblemDetails(ctx, testUserId, testProblemId)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testProblem, problem)
		})
	}
}

func TestProblemsInteractor_GetTodaysProblem(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	todaysProblem := entities.Problems{ProblemId: "p-1", UserId: testUserId, Status: "Open"}

	tests := []struct {
		name        string
		prepareFunc func(m problemsInteractorMocks)
		wantErr     bool
		wantCode    int
		wantProblem entities.Problems
	}{
		{
			name: "problem already exists today — returns it, no ingest triggered",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(todaysProblem, nil)
			},
			wantProblem: todaysProblem,
		},
		{
			name: "db read failure propagates",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).
					Return(entities.Problems{}, response.NewDatabaseError(errors.New("db down")))
			},
			wantErr:  true,
			wantCode: 500,
		},
		{
			name: "returned problem belongs to another user — 401, not leaked",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).
					Return(entities.Problems{ProblemId: "p-1", UserId: "someone-else"}, nil)
			},
			wantErr:  true,
			wantCode: 401,
		},
		{
			name: "no problem yet, retry already used up today — 404",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(entities.Problems{}, nil)
				m.ingest.EXPECT().GetIngestByUserId(gomock.Any(), testUserId, gomock.Any(), true).
					Return([]entities.IngestRuns{{IngestRunId: "run-1", Retried: true}}, nil)
			},
			wantErr:  true,
			wantCode: 404,
		},
		{
			name: "checking retry status fails",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(entities.Problems{}, nil)
				m.ingest.EXPECT().GetIngestByUserId(gomock.Any(), testUserId, gomock.Any(), true).
					Return(nil, response.NewDatabaseError(errors.New("db down")))
			},
			wantErr:  true,
			wantCode: 500,
		},
		{
			name: "retry not used yet, ingest run fails",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(entities.Problems{}, nil)
				m.ingest.EXPECT().GetIngestByUserId(gomock.Any(), testUserId, gomock.Any(), true).Return(nil, nil)
				m.runner.EXPECT().RunForUser(gomock.Any(), testUserId, true).
					Return(response.NewNoProblemToday(errors.New("refresh token invalid")))
			},
			wantErr:  true,
			wantCode: 404,
		},
		{
			name: "retry succeeds and produces a problem",
			prepareFunc: func(m problemsInteractorMocks) {
				gomock.InOrder(
					m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(entities.Problems{}, nil),
					m.ingest.EXPECT().GetIngestByUserId(gomock.Any(), testUserId, gomock.Any(), true).Return(nil, nil),
					m.runner.EXPECT().RunForUser(gomock.Any(), testUserId, true).Return(nil),
					m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(todaysProblem, nil),
				)
			},
			wantProblem: todaysProblem,
		},
		{
			name: "retry succeeds but still produces no problem — 404, not an empty 200",
			prepareFunc: func(m problemsInteractorMocks) {
				gomock.InOrder(
					m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(entities.Problems{}, nil),
					m.ingest.EXPECT().GetIngestByUserId(gomock.Any(), testUserId, gomock.Any(), true).Return(nil, nil),
					m.runner.EXPECT().RunForUser(gomock.Any(), testUserId, true).Return(nil),
					m.problems.EXPECT().GetTodaysproblem(gomock.Any(), testUserId, gomock.Any()).Return(entities.Problems{}, nil),
				)
			},
			wantErr:  true,
			wantCode: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			interactor, mocks := newTestProblemsInteractor(ctrl)
			tt.prepareFunc(mocks)

			problem, err := interactor.GetTodaysProblem(ctx, testUserId)

			if tt.wantErr {
				var appErr *response.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantCode, appErr.Status.Code)
				}
				assert.Equal(t, entities.Problems{}, problem)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantProblem, problem)
		})
	}
}

func TestProblemsInteractor_GetAIHelp(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testProblemId := "test-problem-id"

	cachedHelp := entities.AIHelp{Concept: "hashmaps", Nudge: "can you avoid the nested loop?", Approach: "use a map", Walkthrough: "walk through it"}
	cachedHelpJSON, err := json.Marshal(cachedHelp)
	assert.NoError(t, err)
	cachedHelpStr := string(cachedHelpJSON)

	tests := []struct {
		name        string
		prepareFunc func(m problemsInteractorMocks)
		wantErr     bool
		wantCode    int
		wantHelp    entities.AIHelp
	}{
		{
			name: "cache hit — unmarshals stored ai_help, no settings lookup, no claude call",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{ProblemId: testProblemId, AiHelp: &cachedHelpStr}, nil)
			},
			wantHelp: cachedHelp,
		},
		{
			name: "problem not found — 404",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{}, response.NewProblemNotFound(errors.New("not found")))
			},
			wantErr:  true,
			wantCode: 404,
		},
		{
			name: "reading the problem fails",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{}, response.NewDatabaseError(errors.New("db down")))
			},
			wantErr:  true,
			wantCode: 500,
		},
		{
			name: "stored ai_help is corrupt — surfaces as an error, not an empty response",
			prepareFunc: func(m problemsInteractorMocks) {
				corrupt := "{not valid json"
				m.problems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{ProblemId: testProblemId, AiHelp: &corrupt}, nil)
			},
			wantErr:  true,
			wantCode: 500,
		},
		{
			name: "cache miss, reading settings fails",
			prepareFunc: func(m problemsInteractorMocks) {
				m.problems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{ProblemId: testProblemId}, nil)
				m.settings.EXPECT().GetUserSetting(gomock.Any(), testUserId).
					Return(entities.Settings{}, response.NewDatabaseError(errors.New("db down")))
			},
			wantErr:  true,
			wantCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			interactor, mocks := newTestProblemsInteractor(ctrl)
			tt.prepareFunc(mocks)

			aiHelp, err := interactor.GetAIHelp(ctx, testUserId, testProblemId)

			if tt.wantErr {
				var appErr *response.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantCode, appErr.Status.Code)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantHelp, aiHelp)
		})
	}
}

// getAIHelpFromClaude's success path requires an actual Claude API call —
// claudeClient is a concrete anthropic.Client, not an interface, so it can't
// be mocked here. Only the schema-building half is exercised directly.
func TestToolInputSchema(t *testing.T) {
	schema, err := toolInputSchema(entities.AIHelp{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"concept", "nudge", "approach", "walkthrough"}, schema.Required)
}
