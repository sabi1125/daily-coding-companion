package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/domain/entities"
	interactorMock "backend/internal/domain/interactor/inputport/mock"
	"backend/internal/response"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProblemsController_GetProblems(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockProblemsInteractorInputPort)
		expectedStatus int
	}{
		{
			name:  "no status query param — fetches all",
			query: "",
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetProblems(gomock.Any(), "user-1", entities.ProblemStatus("All")).
					Return([]entities.Problems{{ProblemId: "p-1"}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "valid status query param",
			query: "?status=Solved",
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetProblems(gomock.Any(), "user-1", entities.ProblemStatus("Solved")).
					Return([]entities.Problems{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid status query param — 400",
			query:          "?status=NotAStatus",
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing user_id — 401",
			query:          "",
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "interactor error propagates",
			query: "",
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetProblems(gomock.Any(), "user-1", entities.ProblemStatus("All")).
					Return(nil, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewProblemsController(mockInteractor)
			e.GET("/problems", controller.GetProblems)

			req := httptest.NewRequest(http.MethodGet, "/problems"+tt.query, nil)
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestProblemsController_GetProblems_ResponseShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	title := "Two Sum"
	createdAt := time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)
	rawProblem := "some raw problem text that should never reach the client here"

	mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().GetProblems(gomock.Any(), "user-1", entities.ProblemStatus("All")).
		Return([]entities.Problems{
			{
				ProblemId:       "p-1",
				Title:           &title,
				Status:          "Open",
				NeedsReviewFlag: true,
				CreatedAt:       createdAt,
				RawProblem:      rawProblem,
			},
		}, nil)

	e := newTestEcho()
	controller := NewProblemsController(mockInteractor)
	e.GET("/problems", controller.GetProblems)

	req := httptest.NewRequest(http.MethodGet, "/problems", nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), rawProblem, "full problem content must not leak into the list response")

	var body response.GetProblemsResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, 1, body.Total)
	if assert.Len(t, body.Result, 1) {
		assert.Equal(t, "p-1", body.Result[0].ProblemId)
		assert.Equal(t, &title, body.Result[0].Title)
		assert.Equal(t, "Open", body.Result[0].Status)
		assert.True(t, body.Result[0].NeedsReviewFlag)
		assert.True(t, createdAt.Equal(body.Result[0].CreatedAt))
	}
}

const testProblemID = "c39a04db-e00b-426b-9e4a-9b8e2cb29a10"

func TestProblemsController_GetProblemDetail(t *testing.T) {
	tests := []struct {
		name           string
		problemID      string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockProblemsInteractorInputPort)
		expectedStatus int
	}{
		{
			name:      "success",
			problemID: testProblemID,
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetProblemDetails(gomock.Any(), "user-1", testProblemID).
					Return(entities.Problems{ProblemId: testProblemID, RawProblem: "raw text"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id — 401",
			problemID:      testProblemID,
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "id isn't a uuid — 400",
			problemID:      "problem-1",
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "uuid-shaped but malformed id — 400",
			problemID:      "c39a04db-e00b-426b-9e4a-9b8e2cb29a1",
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "problem not found — 404",
			problemID: testProblemID,
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetProblemDetails(gomock.Any(), "user-1", testProblemID).
					Return(entities.Problems{}, response.NewProblemNotFound(errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "interactor error propagates",
			problemID: testProblemID,
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetProblemDetails(gomock.Any(), "user-1", testProblemID).
					Return(entities.Problems{}, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewProblemsController(mockInteractor)
			e.GET("/problems/:id", controller.GetProblemDetail)

			req := httptest.NewRequest(http.MethodGet, "/problems/"+tt.problemID, nil)
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestProblemsController_GetProblemDetail_ResponseShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	title := "Two Sum"

	mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().GetProblemDetails(gomock.Any(), "user-1", testProblemID).
		Return(entities.Problems{
			ProblemId:  testProblemID,
			RawProblem: "raw text",
			Title:      &title,
			Status:     "Open", // must not leak into the response
		}, nil)

	e := newTestEcho()
	controller := NewProblemsController(mockInteractor)
	e.GET("/problems/:id", controller.GetProblemDetail)

	req := httptest.NewRequest(http.MethodGet, "/problems/"+testProblemID, nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotContains(t, body, "result", "detail response must be a bare object, not wrapped")
	assert.NotContains(t, body, "status", "status is not part of this endpoint's response")

	var detail response.ProblemDetail
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detail))
	assert.Equal(t, testProblemID, detail.ProblemId)
	assert.Equal(t, "raw text", detail.RawProblem)
	assert.Equal(t, &title, detail.Title)
}

func TestProblemsController_GetTodaysProblem(t *testing.T) {
	tests := []struct {
		name           string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockProblemsInteractorInputPort)
		expectedStatus int
	}{
		{
			name: "success",
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetTodaysProblem(gomock.Any(), "user-1").
					Return(entities.Problems{ProblemId: testProblemID, Status: "Open"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id — 401",
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "no problem available today — 404",
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetTodaysProblem(gomock.Any(), "user-1").
					Return(entities.Problems{}, response.NewNoProblemToday(errors.New("retry already used up")))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "interactor error propagates",
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetTodaysProblem(gomock.Any(), "user-1").
					Return(entities.Problems{}, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewProblemsController(mockInteractor)
			e.GET("/problems/today", controller.GetTodaysProblem)

			req := httptest.NewRequest(http.MethodGet, "/problems/today", nil)
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestProblemsController_GetTodaysProblem_ResponseShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().GetTodaysProblem(gomock.Any(), "user-1").
		Return(entities.Problems{ProblemId: testProblemID, RawProblem: "raw text", Status: "Open"}, nil)

	e := newTestEcho()
	controller := NewProblemsController(mockInteractor)
	e.GET("/problems/today", controller.GetTodaysProblem)

	req := httptest.NewRequest(http.MethodGet, "/problems/today", nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body response.TodaysProblem
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, testProblemID, body.ProblemId)
	assert.Equal(t, "Open", body.Status)
}

func TestProblemsController_GetAIHelp(t *testing.T) {
	tests := []struct {
		name           string
		problemID      string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockProblemsInteractorInputPort)
		expectedStatus int
	}{
		{
			name:      "success",
			problemID: testProblemID,
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetAIHelp(gomock.Any(), "user-1", testProblemID).
					Return(entities.AIHelp{Concept: "hashmaps", Nudge: "n", Approach: "a", Walkthrough: "w"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id — 401",
			problemID:      testProblemID,
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "id isn't a uuid — 400",
			problemID:      "problem-1",
			mockSetup:      func(m *interactorMock.MockProblemsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "problem not found — 404",
			problemID: testProblemID,
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetAIHelp(gomock.Any(), "user-1", testProblemID).
					Return(entities.AIHelp{}, response.NewProblemNotFound(errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "interactor error propagates",
			problemID: testProblemID,
			mockSetup: func(m *interactorMock.MockProblemsInteractorInputPort) {
				m.EXPECT().GetAIHelp(gomock.Any(), "user-1", testProblemID).
					Return(entities.AIHelp{}, response.NewInternalError(errors.New("claude call failed")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewProblemsController(mockInteractor)
			e.GET("/problems/:id/help", controller.GetAIHelp)

			req := httptest.NewRequest(http.MethodGet, "/problems/"+tt.problemID+"/help", nil)
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestProblemsController_GetAIHelp_ResponseShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInteractor := interactorMock.NewMockProblemsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().GetAIHelp(gomock.Any(), "user-1", testProblemID).
		Return(entities.AIHelp{Concept: "hashmaps", Nudge: "n", Approach: "a", Walkthrough: "w"}, nil)

	e := newTestEcho()
	controller := NewProblemsController(mockInteractor)
	e.GET("/problems/:id/help", controller.GetAIHelp)

	req := httptest.NewRequest(http.MethodGet, "/problems/"+testProblemID+"/help", nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotContains(t, body, "problem_id", "problem_id is redundant here — the caller already has it from the path")

	var help response.AIHelp
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &help))
	assert.Equal(t, "hashmaps", help.Concept)
	assert.Equal(t, "n", help.Nudge)
	assert.Equal(t, "a", help.Approach)
	assert.Equal(t, "w", help.Walkthrough)
}
