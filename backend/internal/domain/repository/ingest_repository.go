package repository

import (
	"context"
	"time"

	"backend/internal/domain/entities"
	"backend/internal/response"
	"backend/internal/tx"

	"gorm.io/gorm"
)

type IngestRepository struct {
	db *gorm.DB
}

func NewIngestRepository(db *gorm.DB) *IngestRepository {
	return &IngestRepository{
		db: db,
	}
}

func (repository *IngestRepository) GetIngestByUserId(ctx context.Context, userId string, ingestDate time.Time, retried bool) (ingest []entities.IngestRuns, err error) {
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Where("user_id = ? AND ingest_date = ? AND retried = ?", userId, ingestDate, retried).Find(&ingest).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}

func (repository *IngestRepository) CreateIngestWithErr(ctx context.Context, ingest entities.IngestRuns) (err error) {
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Create(&ingest).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}
