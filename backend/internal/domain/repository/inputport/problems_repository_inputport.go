//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
	"time"

	"backend/internal/domain/entities"
)

type ProblemsRepositoryInputPort interface {
	GetProblems(ctx context.Context, userId string, status string, difficulty string) (problems []entities.Problems, err error)
	GetProblemDetails(ctx context.Context, userId string, problemId string) (problem entities.Problems, err error)
	CreateProblem(ctx context.Context, problem *entities.Problems) (err error)
	GetTodaysproblem(ctx context.Context, userId string, todaysDate time.Time) (problem entities.Problems, err error)
	UpdateProblemWithAIHelp(ctx context.Context, userId string, problemId string, aiHelp string) (err error)
	GetProblemsCount(ctx context.Context, userId string) (problemStateCount entities.ProblemStateCount, err error)
}
