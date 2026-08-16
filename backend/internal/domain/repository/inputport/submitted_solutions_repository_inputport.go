//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type SubmittedSolutionsRepositoryInputPort interface {
	GetSubmittedSolutions(ctx context.Context, userId string, problemId string) (solutions []entities.SubmittedSolutions, err error)
	SubmitSolution(ctx context.Context, submittedSolution entities.SubmittedSolutions) (createdSolution entities.SubmittedSolutions, err error)
}
