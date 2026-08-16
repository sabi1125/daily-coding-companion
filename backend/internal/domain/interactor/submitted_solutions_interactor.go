package interactor

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
)

type SubmittedSolutionsInteractor struct {
	submittedSolutionsRepository inputport.SubmittedSolutionsRepositoryInputPort
}

func NewSubmittedSolutionInteractor(submissionsRepository inputport.SubmittedSolutionsRepositoryInputPort) *SubmittedSolutionsInteractor {
	return &SubmittedSolutionsInteractor{
		submittedSolutionsRepository: submissionsRepository,
	}
}

func (interactor *SubmittedSolutionsInteractor) GetSubmittedSolutions(ctx context.Context, userId string, problemId string) (submittedSolutions []entities.SubmittedSolutions, err error) {
	logger.Info("SubmittedSolutions: GetSubmittedSolutions")

	submittedSolutions, err = interactor.submittedSolutionsRepository.GetSubmittedSolutions(ctx, userId, problemId)
	if err != nil {
		return
	}

	return
}
