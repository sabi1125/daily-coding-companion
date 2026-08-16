//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type ProblemsRepositoryInputPort interface {
	GetProblems(ctx context.Context, userId string, status string) (problems []entities.Problems, err error)
	GetProblemDetails(ctx context.Context, userId string, problemId string) (problem entities.Problems, err error)
	CreateProblem(ctx context.Context, problem *entities.Problems) (err error)
}
