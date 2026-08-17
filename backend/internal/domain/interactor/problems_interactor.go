package interactor

import (
	"context"
	"errors"

	"backend/internal/domain/entities"
	ingestInputPort "backend/internal/domain/ingest_runner"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/util"
)

type ProblemsInteractor struct {
	problemsRepository inputport.ProblemsRepositoryInputPort
	ingestRunner       ingestInputPort.IngestRunnerInputPort
	ingestRepository   inputport.IngestRepositoryInputPort
}

func NewProblemsInteractor(problemsRepository inputport.ProblemsRepositoryInputPort, ingestRunner ingestInputPort.IngestRunnerInputPort, ingestRepository inputport.IngestRepositoryInputPort) *ProblemsInteractor {
	return &ProblemsInteractor{
		problemsRepository: problemsRepository,
		ingestRunner:       ingestRunner,
		ingestRepository:   ingestRepository,
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

func (interactor *ProblemsInteractor) GetProblemDetails(ctx context.Context, userId string, problemId string) (problem entities.Problems, err error) {
	logger.Info("ProblemsInteractor: GetProblemDetails")

	problem, err = interactor.problemsRepository.GetProblemDetails(ctx, userId, problemId)
	if err != nil {
		return
	}

	return
}

func (interactor *ProblemsInteractor) GetTodaysProblem(ctx context.Context, userId string) (problem entities.Problems, err error) {
	logger.Info("ProblemsInteractor: GetProblemDetails")
	timeProvider := util.NewTimeProvider()
	todaysDate := timeProvider.TodaysDate()

	problem, err = interactor.problemsRepository.GetTodaysproblem(ctx, userId, todaysDate)
	if err != nil {
		return
	}

	if problem.ProblemId == "" {
		// check if retried
		ingest, getErr := interactor.ingestRepository.GetIngestByUserId(ctx, userId, todaysDate, true)
		if getErr != nil {
			err = getErr
			return
		}

		if len(ingest) > 0 {
			err = response.NewProblemNotFound(errors.New("ingest retry has already ran. todays problems does not exist. contact the developer"))
			return
		}

		// retry
		if ingestErr := interactor.ingestRunner.Ingest(ctx, []string{userId}, true); ingestErr != nil {
			err = ingestErr
			return
		}

		problem, err = interactor.problemsRepository.GetTodaysproblem(ctx, userId, todaysDate)
		if err != nil {
			return
		}
	}

	return
}
