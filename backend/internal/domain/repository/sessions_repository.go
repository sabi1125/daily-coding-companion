package repository

import (
	"context"

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
