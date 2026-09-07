//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
	"time"

	"backend/internal/domain/entities"
	"backend/internal/response"
)

type SubmittedSolutionsRepositoryInputPort interface {
	GetSubmittedSolutions(ctx context.Context, userId string, problemId string) (solutions []entities.SubmittedSolutions, err error)
	SubmitSolution(ctx context.Context, submittedSolution entities.SubmittedSolutions) (createdSolution entities.SubmittedSolutions, err error)
	GetDatedForHeatMap(ctx context.Context, userId string, sixMonthsBeforeToday time.Time) (submittedDated []response.HeatMapDates, err error)
}
