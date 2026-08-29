package repository

import (
	"context"

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
