package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"

	"gorm.io/gorm"
)

type OauthRepository struct {
	db *gorm.DB
}

func NewOauthRepository(db *gorm.DB) *OauthRepository {
	return &OauthRepository{
		db: db,
	}
}

func (repository *OauthRepository) FindUserBySub(ctx context.Context, sub string) (oauthCredentials *entities.OauthCredentials, err error) {
	logger.Infof("OauthRepository: FindUserBySub")

	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	var creds entities.OauthCredentials
	if err = db.Where("oauth_id = ?", sub).Take(&creds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Infof("No records found for given sub")
			err = nil
			return
		}
		err = response.NewDatabaseError(err)
		return
	}

	oauthCredentials = &creds
	return
}

func (repository *OauthRepository) CreateOauthCredentials(ctx context.Context, oauthCreds *entities.OauthCredentials) (err error) {
	logger.Infof("OauthRepository: CreateOauthCredentials")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Create(oauthCreds).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}

func (repository *OauthRepository) UpdateOauthInformationWithSub(ctx context.Context, sub string, refreshToken string, expiryAt time.Time) (err error) {
	logger.Infof("OauthRepository: UpdateOauthInformationWithSub")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Model(&entities.OauthCredentials{}).Where("oauth_id = ?", sub).Updates(entities.OauthCredentials{RefreshToken: refreshToken, ExpiryAt: expiryAt}).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}

	return
}
