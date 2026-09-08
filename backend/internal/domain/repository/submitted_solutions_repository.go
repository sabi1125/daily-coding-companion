package repository

import (
	"context"
	"time"

	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"

	"gorm.io/gorm"
)

type SubmittedSolutionsRepository struct {
	db *gorm.DB
}

func NewSubmittedSolutionsRepository(db *gorm.DB) *SubmittedSolutionsRepository {
	return &SubmittedSolutionsRepository{
		db: db,
	}
}

func (repository *SubmittedSolutionsRepository) GetSubmittedSolutions(ctx context.Context, userId string, problemId string) (submittedSolutions []entities.SubmittedSolutions, err error) {
	logger.Info("SubmittedSolutionRepository: GetSubmittedSolutions")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Select("submitted_solutions.*").
		Joins("JOIN problems ON problems.problem_id = submitted_solutions.problem_id").
		Where("problems.problem_id = ? AND problems.user_id = ?", problemId, userId).
		Order("submitted_solutions.submitted_at DESC").
		Find(&submittedSolutions).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}

	return
}

func (repository *SubmittedSolutionsRepository) SubmitSolution(ctx context.Context, submittedSolution entities.SubmittedSolutions) (createdSolution entities.SubmittedSolutions, err error) {
	logger.Info("SubmittedSolutionRepository: SubmitSolution")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	createdSolution = submittedSolution
	if err = db.Create(&createdSolution).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}

	return
}

func (repository *SubmittedSolutionsRepository) GetDatesForHeatMap(ctx context.Context, userId string, sixMonthsBeforeToday time.Time) (submittedDates []response.HeatMapDates, err error) {
	logger.Info("SubmittedSolutionRepository: GetDatesForHeatMap")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Select("date(s.submitted_at) as submitted_at, (date(s.submitted_at) = date(p.created_at)) as submitted_same_day_flag").
		Table("submitted_solutions s").
		Joins("LEFT JOIN problems p ON s.problem_id = p.problem_id").
		Where("p.user_id = ? AND s.submitted_at >= ?", userId, sixMonthsBeforeToday).Order("s.submitted_at ASC").
		Find(&submittedDates).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}

	return
}
