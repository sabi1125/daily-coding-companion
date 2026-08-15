package interactor

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/tx"
)

type SettingsInteractor struct {
	settingRepository inputport.SettingsRepositoryInputPort

	txManager tx.Manager
}

func NewSettingsInteractor(settingsInteractor inputport.SettingsRepositoryInputPort, txManager tx.Manager) *SettingsInteractor {
	return &SettingsInteractor{
		settingRepository: settingsInteractor,
		txManager:         txManager,
	}
}

func (interactor *SettingsInteractor) GetUserSetting(ctx context.Context, userId string) (setting entities.Settings, err error) {
	logger.Info("SettingsInteractor: GetUserSetting")

	setting, err = interactor.settingRepository.GetUserSetting(ctx, userId)
	if err != nil {
		return
	}

	return
}

func (interactor *SettingsInteractor) UpdateUserSetting(ctx context.Context, userId string, preferences string) (setting entities.Settings, err error) {
	logger.Info("SettingsInteractor: UpdateUserSetting")

	err = interactor.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		txError := interactor.settingRepository.UpdateUserSetting(ctx, userId, preferences)
		if txError != nil {
			return txError
		}

		setting, txError = interactor.settingRepository.GetUserSetting(ctx, userId)
		if txError != nil {
			return txError
		}
		return txError
	})
	if err != nil {
		return
	}

	return
}
