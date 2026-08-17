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

type ProblemsRepository struct {
	db *gorm.DB
}

func NewProblemsRepository(db *gorm.DB) *ProblemsRepository {
	return &ProblemsRepository{
		db: db,
	}
}

func (repository *ProblemsRepository) GetProblems(ctx context.Context, userId string, status string) (problems []entities.Problems, err error) {
	logger.Infof("ProblemsRepository: GetProblems")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Select("problem_id", "title", "needs_review_flag", "created_at").
		Where("user_id = ?", userId).Preload("Submissions").Find(&problems).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}

	for i := range problems {
		problems[i].Status = deriveProblemStatus(problems[i].Submissions)
	}

	if status == "" || status == "All" {
		return
	}

	filtered := make([]entities.Problems, 0, len(problems))
	for _, problem := range problems {
		if problem.Status == status {
			filtered = append(filtered, problem)
		}
	}
	problems = filtered

	return
}

func deriveProblemStatus(submissions []entities.SubmittedSolutions) string {
	if len(submissions) == 0 {
		return string(entities.ProblemOpen)
	}

	for _, submission := range submissions {
		if submission.Status == string(entities.ProblemSolved) {
			return string(entities.ProblemSolved)
		}
	}

	return string(entities.ProblemFailed)
}

func (repository *ProblemsRepository) GetProblemDetails(ctx context.Context, userId string, problemId string) (problem entities.Problems, err error) {
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}
	if err = db.Where("user_id = ? and problem_id = ?", userId, problemId).Take(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = response.NewProblemNotFound(err)
			return
		}
		err = response.NewDatabaseError(err)
		return
	}
	return
}

func (repository *ProblemsRepository) CreateProblem(ctx context.Context, problem *entities.Problems) (err error) {
	logger.Infof("ProblemsRepository: CreateProblem")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Create(problem).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}

func (repository *ProblemsRepository) GetTodaysproblem(ctx context.Context, userId string, todaysDate time.Time) (problem entities.Problems, err error) {
	logger.Infof("ProblemsRepository: GetTodaysproblem")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.
		Joins("JOIN ingest_runs ON ingest_runs.problem_id = problems.problem_id").
		Where("problems.user_id = ? AND ingest_runs.ingest_date = ?", userId, todaysDate).
		Take(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		err = response.NewDatabaseError(err)
		return
	}
	return
}
