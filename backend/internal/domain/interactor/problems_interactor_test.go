package interactor

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain/entities"
	inputportMock "backend/internal/domain/repository/inputport/mock"
	"backend/internal/response"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProblemsInteractor_GetProblems(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testProblems := []entities.Problems{{ProblemId: "p-1", Status: "Open"}}
	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name        string
		status      entities.ProblemStatus
		prepareFunc func(mpr *inputportMock.MockProblemsRepositoryInputPort)
		wantedError error
	}{
		{
			name:   "success",
			status: entities.ProblemStatus("All"),
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "All").Return(testProblems, nil)
			},
		},
		{
			name:   "forwards a specific status filter",
			status: entities.ProblemStatus("Solved"),
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "Solved").Return(testProblems, nil)
			},
		},
		{
			name:   "fails to get problems",
			status: entities.ProblemStatus("All"),
			prepareFunc: func(mpr *inputportMock.MockProblemsRepositoryInputPort) {
				mpr.EXPECT().GetProblems(gomock.Any(), testUserId, "All").Return(nil, errRepo)
			},
			wantedError: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProblems := inputportMock.NewMockProblemsRepositoryInputPort(ctrl)
			tt.prepareFunc(mockProblems)

			interactor := NewProblemsInteractor(mockProblems)

			problems, err := interactor.GetProblems(ctx, testUserId, tt.status)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testProblems, problems)
		})
	}
}
