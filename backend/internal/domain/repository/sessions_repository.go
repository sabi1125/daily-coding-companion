package repository

import (
	"context"
	"errors"

	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"

	"gorm.io/gorm"
)

type SessionsRepository struct {
	db *gorm.DB
}

func NewSessionsRepository(db *gorm.DB) *SessionsRepository {
	return &SessionsRepository{
		db: db,
	}
}

func (repository *SessionsRepository) CreateSession(ctx context.Context, session *entities.Sessions) (err error) {
	logger.Infof("SessionsRepository: CreateSession")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Create(&session).Error; err != nil {
		err = response.NewInternalError(err)
		return
	}
	return
}

func (repository *SessionsRepository) GetSessionById(ctx context.Context, sessionId string) (session *entities.Sessions, err error) {
	logger.Infof("SessionsRepository: GetSessionById")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	var found entities.Sessions
	if err = db.Where("session_id = ?", sessionId).Take(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		err = response.NewInternalError(err)
		return
	}

	session = &found
	return
}

func (repository *SessionsRepository) DeleteUserSession(ctx context.Context, sessionId string, userId string) (err error) {
	logger.Info("SessionRepository: DeleteUserSession")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	// checking the row affected i not necessary the auth middleware already checks if session actually exists
	// if session never exists we will not be able to reach here
	res := db.Where("session_id = ? and user_id = ?", sessionId, userId).Delete(&entities.Sessions{})
	if res.Error != nil {
		err = response.NewDatabaseError(res.Error)
		return
	}

	return
}
