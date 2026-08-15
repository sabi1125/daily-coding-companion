package interactor

import (
	"context"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
)

type ProblemsInteractor struct {
	problemsRepository inputport.ProblemsRepositoryInputPort
}

func NewProblemsInteractor(problemsRepository inputport.ProblemsRepositoryInputPort) *ProblemsInteractor {
	return &ProblemsInteractor{
		problemsRepository: problemsRepository,
	}
}

func (interactor *ProblemsInteractor) GetProblems(ctx context.Context, userId string, status entities.ProblemStatus) (problems []entities.Problems, err error) {
	logger.Info("ProblemInteractor: GetProblems")

	problems, err = interactor.problemsRepository.GetProblems(ctx, userId, string(status))
	if err != nil {
		return
	}
	return
}
