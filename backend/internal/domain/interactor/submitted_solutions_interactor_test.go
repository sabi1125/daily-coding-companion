package interactor

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain/entities"
	inputportMock "backend/internal/domain/repository/inputport/mock"
	"backend/internal/response"
	txMock "backend/internal/tx/mock"
	utilMock "backend/internal/util/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSubmittedSolutionsInteractor_GetSubmittedSolutions(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testProblemId := "test-problem-id"
	testSolutions := []entities.SubmittedSolutions{{SolutionId: "s-1", ProblemId: testProblemId}}
	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name        string
		prepareFunc func(m *inputportMock.MockSubmittedSolutionsRepositoryInputPort)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(m *inputportMock.MockSubmittedSolutionsRepositoryInputPort) {
				m.EXPECT().GetSubmittedSolutions(gomock.Any(), testUserId, testProblemId).Return(testSolutions, nil)
			},
		},
		{
			name: "fails to get submissions",
			prepareFunc: func(m *inputportMock.MockSubmittedSolutionsRepositoryInputPort) {
				m.EXPECT().GetSubmittedSolutions(gomock.Any(), testUserId, testProblemId).Return(nil, errRepo)
			},
			wantedError: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSubmissions := inputportMock.NewMockSubmittedSolutionsRepositoryInputPort(ctrl)
			tt.prepareFunc(mockSubmissions)

			mockProblems := inputportMock.NewMockProblemsRepositoryInputPort(ctrl)
			mockUUID := utilMock.NewMockUUIDGenerator(ctrl)
			mockTx := txMock.NewMockManager(ctrl)
			mockPiston := inputportMock.NewMockPistonApiRepositoryInputport(ctrl)

			interactor := NewSubmittedSolutionInteractor(mockSubmissions, mockPiston, mockProblems, mockUUID, mockTx)

			solutions, err := interactor.GetSubmittedSolutions(ctx, testUserId, testProblemId)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testSolutions, solutions)
		})
	}
}

func TestSubmittedSolutionsInteractor_SubmitSolution(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testProblemId := "test-problem-id"
	testSolutionId := "test-solution-id"
	testBody := entities.SubmittedSolutionsBody{Solution: "def two_sum(): pass", Status: "Solved"}
	testSolution := entities.SubmittedSolutions{
		SolutionId: testSolutionId,
		ProblemId:  testProblemId,
		Solution:   testBody.Solution,
		Status:     testBody.Status,
	}
	errUUID := errors.New("rand read failed")
	errNotFound := response.NewProblemNotFound(errors.New("not found"))
	errInsert := response.NewDatabaseError(errors.New("insert failed"))

	tests := []struct {
		name        string
		prepareFunc func(
			mSubmissions *inputportMock.MockSubmittedSolutionsRepositoryInputPort,
			mProblems *inputportMock.MockProblemsRepositoryInputPort,
			mUUID *utilMock.MockUUIDGenerator,
		)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(
				mSubmissions *inputportMock.MockSubmittedSolutionsRepositoryInputPort,
				mProblems *inputportMock.MockProblemsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testSolutionId, nil)
				mProblems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{ProblemId: testProblemId}, nil)
				mSubmissions.EXPECT().SubmitSolution(gomock.Any(), testSolution).Return(testSolution, nil)
			},
		},
		{
			// solutionId generation happens before the transaction opens, so
			// neither the ownership check nor the insert should ever run.
			name: "uuid generation fails — not wrapped as a database error",
			prepareFunc: func(
				mSubmissions *inputportMock.MockSubmittedSolutionsRepositoryInputPort,
				mProblems *inputportMock.MockProblemsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return("", errUUID)
			},
			wantedError: response.NewInternalError(errUUID),
		},
		{
			// problem missing/not owned by the caller — the insert must not
			// run, since submitted_solutions has no user_id column of its
			// own to enforce ownership at write time.
			name: "problem not found or not owned — insert never attempted",
			prepareFunc: func(
				mSubmissions *inputportMock.MockSubmittedSolutionsRepositoryInputPort,
				mProblems *inputportMock.MockProblemsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testSolutionId, nil)
				mProblems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{}, errNotFound)
			},
			wantedError: errNotFound,
		},
		{
			name: "insert fails",
			prepareFunc: func(
				mSubmissions *inputportMock.MockSubmittedSolutionsRepositoryInputPort,
				mProblems *inputportMock.MockProblemsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testSolutionId, nil)
				mProblems.EXPECT().GetProblemDetails(gomock.Any(), testUserId, testProblemId).
					Return(entities.Problems{ProblemId: testProblemId}, nil)
				mSubmissions.EXPECT().SubmitSolution(gomock.Any(), testSolution).
					Return(entities.SubmittedSolutions{}, errInsert)
			},
			wantedError: errInsert,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSubmissions := inputportMock.NewMockSubmittedSolutionsRepositoryInputPort(ctrl)
			mockProblems := inputportMock.NewMockProblemsRepositoryInputPort(ctrl)
			mockUUID := utilMock.NewMockUUIDGenerator(ctrl)
			mockTx := txMock.NewMockManager(ctrl)
			mockPiston := inputportMock.NewMockPistonApiRepositoryInputport(ctrl)
			mockWithinTransaction(mockTx)

			tt.prepareFunc(mockSubmissions, mockProblems, mockUUID)

			interactor := NewSubmittedSolutionInteractor(mockSubmissions, mockPiston, mockProblems, mockUUID, mockTx)

			solution, err := interactor.SubmitSolution(ctx, testUserId, testProblemId, testBody)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testSolution, solution)
		})
	}
}
