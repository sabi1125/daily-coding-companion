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
	logger.Info("ProblemsInteractor: GetTodaysProblem")
	todaysDate := util.NewTimeProvider().TodaysDate()

	problem, err = interactor.problemsRepository.GetTodaysproblem(ctx, userId, todaysDate)
	if err != nil {
		return
	}
	if err = verifyProblemOwnership(problem, userId); err != nil {
		problem = entities.Problems{}
		return
	}
	if problem.ProblemId != "" {
		return
	}

	alreadyRetried, getErr := interactor.ingestRepository.GetIngestByUserId(ctx, userId, todaysDate, true)
	if getErr != nil {
		err = getErr
		return
	}
	if len(alreadyRetried) > 0 {
		err = response.NewNoProblemToday(errors.New("retry already used up for today"))
		return
	}

	if ingestErr := interactor.ingestRunner.RunForUser(ctx, userId, true); ingestErr != nil {
		err = ingestErr
		return
	}

	problem, err = interactor.problemsRepository.GetTodaysproblem(ctx, userId, todaysDate)
	if err != nil {
		return
	}
	if err = verifyProblemOwnership(problem, userId); err != nil {
		problem = entities.Problems{}
		return
	}
	if problem.ProblemId == "" {
		err = response.NewNoProblemToday(errors.New("ingest did not produce a problem for today"))
	}
	return
}

func verifyProblemOwnership(problem entities.Problems, userId string) error {
	if problem.ProblemId != "" && problem.UserId != userId {
		return response.NewUnauthorized(errors.New("today's problem does not belong to the caller"))
	}
	return nil
}
