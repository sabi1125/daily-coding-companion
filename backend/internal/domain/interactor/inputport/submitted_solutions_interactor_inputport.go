//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/response"
)

type SubmittedSolutionsInteractorInputPort interface {
	GetSubmittedSolutions(ctx context.Context, userId string, problemId string) (problems []entities.SubmittedSolutions, err error)
	SubmitSolution(ctx context.Context, userId string, problemId string, submittedSolutionBody entities.SubmittedSolutionsBody) (problems entities.SubmittedSolutions, err error)
	RunSubmission(ctx context.Context, request entities.SubmittedSolutionForExecution) (response response.ExecuteSubmissionResponse, err error)
}
