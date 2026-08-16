package controller

import (
	"bytes"
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

const testSubmissionProblemID = "c39a04db-e00b-426b-9e4a-9b8e2cb29a10"

func TestSubmittedSolutionsController_GetUserSubmissions(t *testing.T) {
	tests := []struct {
		name           string
		problemID      string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort)
		expectedStatus int
	}{
		{
			name:      "success",
			problemID: testSubmissionProblemID,
			mockSetup: func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {
				m.EXPECT().GetSubmittedSolutions(gomock.Any(), "user-1", testSubmissionProblemID).
					Return([]entities.SubmittedSolutions{{SolutionId: "s-1", ProblemId: testSubmissionProblemID}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "no submissions yet — still 200 with an empty result",
			problemID: testSubmissionProblemID,
			mockSetup: func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {
				m.EXPECT().GetSubmittedSolutions(gomock.Any(), "user-1", testSubmissionProblemID).
					Return([]entities.SubmittedSolutions{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id — 401",
			problemID:      testSubmissionProblemID,
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "id isn't a uuid — 400",
			problemID:      "problem-1",
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "uuid-shaped but malformed id — 400",
			problemID:      "c39a04db-e00b-426b-9e4a-9b8e2cb29a1",
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "interactor error propagates",
			problemID: testSubmissionProblemID,
			mockSetup: func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {
				m.EXPECT().GetSubmittedSolutions(gomock.Any(), "user-1", testSubmissionProblemID).
					Return(nil, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockSubmittedSolutionsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewSubmittedSolutionsController(mockInteractor)
			e.GET("/submissions/:id", controller.GetUserSubmissions)

			req := httptest.NewRequest(http.MethodGet, "/submissions/"+tt.problemID, nil)
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSubmittedSolutionsController_GetUserSubmissions_ResponseShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	submittedAt := time.Date(2026, 8, 10, 14, 30, 0, 0, time.UTC)

	mockInteractor := interactorMock.NewMockSubmittedSolutionsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().GetSubmittedSolutions(gomock.Any(), "user-1", testSubmissionProblemID).
		Return([]entities.SubmittedSolutions{
			{
				SolutionId:  "s-1",
				ProblemId:   testSubmissionProblemID,
				Solution:    "def two_sum(): pass",
				Status:      "Solved",
				SubmittedAt: submittedAt,
			},
		}, nil)

	e := newTestEcho()
	controller := NewSubmittedSolutionsController(mockInteractor)
	e.GET("/submissions/:id", controller.GetUserSubmissions)

	req := httptest.NewRequest(http.MethodGet, "/submissions/"+testSubmissionProblemID, nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var raw map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
	assert.Contains(t, raw, "result", "list response must be wrapped in result/total")
	assert.Contains(t, raw, "total")

	var body response.GetSubmissionsResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, 1, body.Total)
	if assert.Len(t, body.Result, 1) {
		assert.Equal(t, "s-1", body.Result[0].SolutionId)
		assert.Equal(t, testSubmissionProblemID, body.Result[0].ProblemId, "problem_id must survive into the response")
		assert.Equal(t, "def two_sum(): pass", body.Result[0].Solution)
		assert.Equal(t, "Solved", body.Result[0].Status)
		assert.True(t, submittedAt.Equal(body.Result[0].SubmittedAt))
	}

	assert.Contains(t, rec.Body.String(), `"problem_id"`, "problem_id key must actually be on the wire, not just zero-valued")
}

func TestSubmittedSolutionsController_GetUserSubmissions_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInteractor := interactorMock.NewMockSubmittedSolutionsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().GetSubmittedSolutions(gomock.Any(), "user-1", testSubmissionProblemID).
		Return([]entities.SubmittedSolutions{}, nil)

	e := newTestEcho()
	controller := NewSubmittedSolutionsController(mockInteractor)
	e.GET("/submissions/:id", controller.GetUserSubmissions)

	req := httptest.NewRequest(http.MethodGet, "/submissions/"+testSubmissionProblemID, nil)
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body response.GetSubmissionsResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Total)
	assert.Empty(t, body.Result)
	assert.NotNil(t, body.Result, "result must serialize as [] not null")
}

func TestSubmittedSolutionsController_SubmitSolutions(t *testing.T) {
	tests := []struct {
		name           string
		problemID      string
		body           string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort)
		expectedStatus int
	}{
		{
			name:      "success",
			problemID: testSubmissionProblemID,
			body:      `{"solution":"def two_sum(): pass","status":"Solved"}`,
			mockSetup: func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {
				m.EXPECT().SubmitSolution(gomock.Any(), "user-1", testSubmissionProblemID, entities.SubmittedSolutionsBody{
					Solution: "def two_sum(): pass",
					Status:   "Solved",
				}).Return(entities.SubmittedSolutions{SolutionId: "s-1", ProblemId: testSubmissionProblemID}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id — 401",
			problemID:      testSubmissionProblemID,
			body:           `{"solution":"code","status":"Solved"}`,
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "id isn't a uuid — 400",
			problemID:      "problem-1",
			body:           `{"solution":"code","status":"Solved"}`,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "uuid-shaped but malformed id — 400",
			problemID:      "c39a04db-e00b-426b-9e4a-9b8e2cb29a1",
			body:           `{"solution":"code","status":"Solved"}`,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty solution — 400",
			problemID:      testSubmissionProblemID,
			body:           `{"solution":"","status":"Solved"}`,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "status not Solved/Failed — 400",
			problemID:      testSubmissionProblemID,
			body:           `{"solution":"code","status":"Open"}`,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed json body — 400",
			problemID:      testSubmissionProblemID,
			body:           `{"solution":`,
			mockSetup:      func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "problem not found or not owned — 404",
			problemID: testSubmissionProblemID,
			body:      `{"solution":"code","status":"Solved"}`,
			mockSetup: func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {
				m.EXPECT().SubmitSolution(gomock.Any(), "user-1", testSubmissionProblemID, entities.SubmittedSolutionsBody{
					Solution: "code",
					Status:   "Solved",
				}).Return(entities.SubmittedSolutions{}, response.NewProblemNotFound(errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "interactor error propagates",
			problemID: testSubmissionProblemID,
			body:      `{"solution":"code","status":"Solved"}`,
			mockSetup: func(m *interactorMock.MockSubmittedSolutionsInteractorInputPort) {
				m.EXPECT().SubmitSolution(gomock.Any(), "user-1", testSubmissionProblemID, entities.SubmittedSolutionsBody{
					Solution: "code",
					Status:   "Solved",
				}).Return(entities.SubmittedSolutions{}, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockSubmittedSolutionsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewSubmittedSolutionsController(mockInteractor)
			e.POST("/submissions/:id", controller.SubmitSolutions)

			req := httptest.NewRequest(http.MethodPost, "/submissions/"+tt.problemID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestSubmittedSolutionsController_SubmitSolutions_ResponseShape locks in
// that the 200 response is a bare object (unlike GetUserSubmissions' list,
// there's no result/total wrapper) and that problem_id is present —
// entities.SubmittedSolutions tags ProblemId json:"-" so the controller must
// map through response.Submission, same as the GET path.
func TestSubmittedSolutionsController_SubmitSolutions_ResponseShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	submittedAt := time.Date(2026, 8, 10, 14, 30, 0, 0, time.UTC)

	mockInteractor := interactorMock.NewMockSubmittedSolutionsInteractorInputPort(ctrl)
	mockInteractor.EXPECT().SubmitSolution(gomock.Any(), "user-1", testSubmissionProblemID, entities.SubmittedSolutionsBody{
		Solution: "def two_sum(): pass",
		Status:   "Solved",
	}).Return(entities.SubmittedSolutions{
		SolutionId:  "s-1",
		ProblemId:   testSubmissionProblemID,
		Solution:    "def two_sum(): pass",
		Status:      "Solved",
		SubmittedAt: submittedAt,
	}, nil)

	e := newTestEcho()
	controller := NewSubmittedSolutionsController(mockInteractor)
	e.POST("/submissions/:id", controller.SubmitSolutions)

	req := httptest.NewRequest(http.MethodPost, "/submissions/"+testSubmissionProblemID,
		bytes.NewBufferString(`{"solution":"def two_sum(): pass","status":"Solved"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var raw map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
	assert.NotContains(t, raw, "result", "create response must be a bare object, not wrapped")

	var body response.Submission
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "s-1", body.SolutionId)
	assert.Equal(t, testSubmissionProblemID, body.ProblemId, "problem_id must survive into the response")
	assert.Equal(t, "def two_sum(): pass", body.Solution)
	assert.Equal(t, "Solved", body.Status)
	assert.True(t, submittedAt.Equal(body.SubmittedAt))

	assert.Contains(t, rec.Body.String(), `"problem_id"`, "problem_id key must actually be on the wire")
}
