package interactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/domain/entities"
	inputportMock "backend/internal/domain/repository/inputport/mock"
	"backend/internal/response"
	txMock "backend/internal/tx/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSettingsInteractor_GetUserSetting(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testSetting := entities.Settings{SettingID: "setting-1", UserID: testUserId}
	testProblemStateCount := entities.ProblemStateCount{Solved: 3, Unsolved: 2}
	testUser := entities.Users{UserId: testUserId, Email: "test@example.com"}
	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name             string
		sessionCreatedAt time.Time
		prepareFunc      func(msr *inputportMock.MockSettingsRepositoryInputPort, mpr *inputportMock.MockProblemsRepositoryInputPort, mur *inputportMock.MockUsersRepositoryInputPort)
		wantedError      error
		wantNeedsReauth  bool
	}{
		{
			name:             "success, session within 7 days — not needing reauth",
			sessionCreatedAt: time.Now().Add(-1 * time.Hour),
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort, mpr *inputportMock.MockProblemsRepositoryInputPort, mur *inputportMock.MockUsersRepositoryInputPort) {
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(testSetting, nil)
				mpr.EXPECT().GetProblemsCount(gomock.Any(), testUserId).Return(testProblemStateCount, nil)
				mur.EXPECT().GetUser(gomock.Any(), testUserId).Return(testUser, nil)
			},
			wantNeedsReauth: false,
		},
		{
			name:             "session older than 7 days — needs reauth",
			sessionCreatedAt: time.Now().Add(-8 * 24 * time.Hour),
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort, mpr *inputportMock.MockProblemsRepositoryInputPort, mur *inputportMock.MockUsersRepositoryInputPort) {
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(testSetting, nil)
				mpr.EXPECT().GetProblemsCount(gomock.Any(), testUserId).Return(testProblemStateCount, nil)
				mur.EXPECT().GetUser(gomock.Any(), testUserId).Return(testUser, nil)
			},
			wantNeedsReauth: true,
		},
		{
			name:             "fails to get setting",
			sessionCreatedAt: time.Now(),
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort, mpr *inputportMock.MockProblemsRepositoryInputPort, mur *inputportMock.MockUsersRepositoryInputPort) {
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(entities.Settings{}, errRepo)
			},
			wantedError: errRepo,
		},
		{
			name:             "fails to get problem counts",
			sessionCreatedAt: time.Now(),
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort, mpr *inputportMock.MockProblemsRepositoryInputPort, mur *inputportMock.MockUsersRepositoryInputPort) {
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(testSetting, nil)
				mpr.EXPECT().GetProblemsCount(gomock.Any(), testUserId).Return(entities.ProblemStateCount{}, errRepo)
			},
			wantedError: errRepo,
		},
		{
			name:             "fails to get user",
			sessionCreatedAt: time.Now(),
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort, mpr *inputportMock.MockProblemsRepositoryInputPort, mur *inputportMock.MockUsersRepositoryInputPort) {
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(testSetting, nil)
				mpr.EXPECT().GetProblemsCount(gomock.Any(), testUserId).Return(testProblemStateCount, nil)
				mur.EXPECT().GetUser(gomock.Any(), testUserId).Return(entities.Users{}, errRepo)
			},
			wantedError: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSettings := inputportMock.NewMockSettingsRepositoryInputPort(ctrl)
			mockProblems := inputportMock.NewMockProblemsRepositoryInputPort(ctrl)
			mockUsers := inputportMock.NewMockUsersRepositoryInputPort(ctrl)
			tt.prepareFunc(mockSettings, mockProblems, mockUsers)

			interactor := NewSettingsInteractor(mockSettings, mockUsers, mockProblems, nil)

			setting, problemStateCount, email, err := interactor.GetUserSetting(ctx, testUserId, tt.sessionCreatedAt)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantNeedsReauth, setting.NeedsReauth)
			assert.Equal(t, testProblemStateCount, problemStateCount)
			assert.Equal(t, testUser.Email, email)
		})
	}
}

func TestSettingsInteractor_UpdateUserSetting(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testPreferences := "new preferences"
	testSetting := entities.Settings{SettingID: "setting-1", UserID: testUserId}
	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name        string
		prepareFunc func(msr *inputportMock.MockSettingsRepositoryInputPort)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort) {
				msr.EXPECT().UpdateUserSetting(gomock.Any(), testUserId, testPreferences).Return(nil)
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(testSetting, nil)
			},
		},
		{
			name: "fails to update setting",
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort) {
				msr.EXPECT().UpdateUserSetting(gomock.Any(), testUserId, testPreferences).Return(errRepo)
			},
			wantedError: errRepo,
		},
		{
			name: "fails to read setting back after update",
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort) {
				msr.EXPECT().UpdateUserSetting(gomock.Any(), testUserId, testPreferences).Return(nil)
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(entities.Settings{}, errRepo)
			},
			wantedError: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSettings := inputportMock.NewMockSettingsRepositoryInputPort(ctrl)
			mockTx := txMock.NewMockManager(ctrl)
			mockWithinTransaction(mockTx)

			tt.prepareFunc(mockSettings)

			interactor := NewSettingsInteractor(mockSettings, nil, nil, mockTx)

			setting, err := interactor.UpdateUserSetting(ctx, testUserId, testPreferences, time.Now())

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testSetting, setting)
		})
	}
}
