package interactor

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
)

type SettingsInteractor struct {
	settingsInteractor inputport.SettingsRepositoryInputPort
}

func NewSettingsInteractor(settingsInteractor inputport.SettingsRepositoryInputPort) *SettingsInteractor {
	return &SettingsInteractor{
		settingsInteractor: settingsInteractor,
	}
}

func (interactor *SettingsInteractor) GetUserSetting(ctx context.Context, userId string) (setting entities.Settings, err error) {
	logger.Info("SettingsInteractor: GetUserSetting")

	setting, err = interactor.settingsInteractor.GetUserSetting(ctx, userId)
	if err != nil {
		return
	}

	return
}
