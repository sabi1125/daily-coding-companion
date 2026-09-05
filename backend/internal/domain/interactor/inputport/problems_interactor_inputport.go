//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type ProblemsInteractorInputPort interface {
	GetProblems(ctx context.Context, userId string, status entities.ProblemStatus, difficulty []entities.ProblemDifficulty) (problems []entities.Problems, err error)
	GetProblemDetails(ctx context.Context, userId string, problemId string) (problems entities.Problems, err error)
	GetTodaysProblem(ctx context.Context, userId string) (problems entities.Problems, err error)
	GetAIHelp(ctx context.Context, userId string, problemId string) (aiHelp entities.AIHelp, err error)
}
