package interactor

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"
	"backend/internal/util"
)

type SubmittedSolutionsInteractor struct {
	submittedSolutionsRepository inputport.SubmittedSolutionsRepositoryInputPort
	problemsRepository           inputport.ProblemsRepositoryInputPort
	uuidGenerator                util.UUIDGenerator
	txManager                    tx.Manager
}

func NewSubmittedSolutionInteractor(
	submissionsRepository inputport.SubmittedSolutionsRepositoryInputPort,
	problemsRepository inputport.ProblemsRepositoryInputPort,
	uuidGenerator util.UUIDGenerator,
	txManager tx.Manager,
) *SubmittedSolutionsInteractor {
	return &SubmittedSolutionsInteractor{
		submittedSolutionsRepository: submissionsRepository,
		problemsRepository:           problemsRepository,
		uuidGenerator:                uuidGenerator,
		txManager:                    txManager,
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

func (interactor *SubmittedSolutionsInteractor) SubmitSolution(
	ctx context.Context,
	userId string,
	problemId string,
	submittedSolutionBody entities.SubmittedSolutionsBody,
) (solution entities.SubmittedSolutions, err error) {
	logger.Info("SubmittedSolutions: SubmitSolution")

	solutionId, err := interactor.uuidGenerator.NewV7()
	if err != nil {
		err = response.NewInternalError(err)
		return
	}

	err = interactor.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if _, txErr := interactor.problemsRepository.GetProblemDetails(ctx, userId, problemId); txErr != nil {
			return txErr
		}

		submittedSolution := entities.SubmittedSolutions{
			SolutionId: solutionId,
			ProblemId:  problemId,
			Solution:   submittedSolutionBody.Solution,
			Status:     submittedSolutionBody.Status,
		}

		var txErr error
		solution, txErr = interactor.submittedSolutionsRepository.SubmitSolution(ctx, submittedSolution)
		return txErr
	})
	if err != nil {
		return
	}

	return
}
