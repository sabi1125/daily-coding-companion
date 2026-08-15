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

// TestProblemsController_GetProblems_ResponseShape locks in the History
// list shape: {result, total}, with only what the list view renders
// (problem_id/title/status/needs_review_flag/created_at) — full problem
// content belongs to GET /problems/{id}, not this endpoint.
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

	var body getProblemsResponse
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
