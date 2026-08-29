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
		Where("user_id = ?", userId).Order("created_at DESC").Preload("Submissions").Find(&problems).Error; err != nil {
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
	if err = db.Where("user_id = ? and problem_id = ?", userId, problemId).Preload("Submissions").Take(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = response.NewProblemNotFound(err)
			return
		}
		err = response.NewDatabaseError(err)
		return
	}
	problem.Status = deriveProblemStatus(problem.Submissions)
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
		Preload("Submissions").
		Take(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		err = response.NewDatabaseError(err)
		return
	}

	problem.Status = deriveProblemStatus(problem.Submissions)
	return
}

func (repository *ProblemsRepository) UpdateProblemWithAIHelp(ctx context.Context, userId string, problemId string, aiHelp string) (err error) {
	logger.Info("ProblemsRepository: UpdateProblemWithAIHelp")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	err = db.Model(&entities.Problems{}).Where("problem_id = ? and user_id = ?", problemId, userId).Update("ai_help", aiHelp).Error
	if err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}

func (repository *ProblemsRepository) GetProblemsCount(ctx context.Context, userId string) (problemStateCount entities.ProblemStateCount, err error) {
	logger.Infof("ProblemsRepository: GetProblemsCount")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	selectClause := `
		COUNT(DISTINCT CASE WHEN s.status = ? THEN p.problem_id END) AS solved,
		COUNT(DISTINCT p.problem_id) - COUNT(DISTINCT CASE WHEN s.status = ? THEN p.problem_id END) AS unsolved
	`

	err = db.Table("problems AS p").
		Select(selectClause, string(entities.ProblemSolved), string(entities.ProblemSolved)).
		Joins("LEFT JOIN submitted_solutions AS s ON s.problem_id = p.problem_id").
		Where("p.user_id = ?", userId).
		Scan(&problemStateCount).Error
	if err != nil {
		return entities.ProblemStateCount{}, response.NewDatabaseError(err)
	}

	return
}
