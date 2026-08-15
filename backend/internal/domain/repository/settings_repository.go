package repository

import (
	"context"

	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"

	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{
		db: db,
	}
}

func (repository *SettingsRepository) CreateSetting(ctx context.Context, setting *entities.Settings) (err error) {
	logger.Infof("SettingsRepository: CreateSetting")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Create(&setting).Error; err != nil {
		err = response.NewInternalError(err)
		return
	}
	return
}

func (repository *SettingsRepository) GetUserSetting(ctx context.Context, userId string) (setting entities.Settings, err error) {
	logger.Infof("SettingsRepository: GetUserSettings")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Where("user_id = ?", userId).Take(&setting).Error; err != nil {
		err = response.NewInternalError(err)
		return
	}
	return
}
