//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type ProblemsInteractorInputPort interface {
	GetProblems(ctx context.Context, userId string, status entities.ProblemStatus) (problems []entities.Problems, err error)
}
