package interactor

import (
	"context"
	"errors"
	"testing"

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
	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name        string
		prepareFunc func(msr *inputportMock.MockSettingsRepositoryInputPort)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort) {
				msr.EXPECT().GetUserSetting(gomock.Any(), testUserId).Return(testSetting, nil)
			},
		},
		{
			name: "fails to get setting",
			prepareFunc: func(msr *inputportMock.MockSettingsRepositoryInputPort) {
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
			tt.prepareFunc(mockSettings)

			interactor := NewSettingsInteractor(mockSettings, nil)

			setting, err := interactor.GetUserSetting(ctx, testUserId)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testSetting, setting)
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

			interactor := NewSettingsInteractor(mockSettings, mockTx)

			setting, err := interactor.UpdateUserSetting(ctx, testUserId, testPreferences)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testSetting, setting)
		})
	}
}
