//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/response"
)

type PistonApiRepositoryInputport interface {
	RunSubmission(ctx context.Context, request entities.SubmittedSolutionRequest) (response response.ExecuteSubmissionResponse, err error)
}
