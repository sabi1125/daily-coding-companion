package controller

import (
	"errors"
	"net/http"

	"backend/internal/domain/entities"
	"backend/internal/domain/interactor/inputport"
	"backend/internal/infrastructure/middleware"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/validator"

	"github.com/labstack/echo/v4"
)

type ProblemsController struct {
	problemsInteractor inputport.ProblemsInteractorInputPort
}

func NewProblemsController(problemsInteractor inputport.ProblemsInteractorInputPort) *ProblemsController {
	return &ProblemsController{
		problemsInteractor: problemsInteractor,
	}
}

func (controller *ProblemsController) GetProblems(c echo.Context) error {
	logger.Info("ProblemsController: GetProblems")

	ctx := c.Request().Context()
	userId := middleware.UserIDFromContext(ctx)
	if userId == "" {
		err := response.NewUnauthorized(errors.New("invalid user"))
		return err
	}

	var params entities.GetProblemParams
	if err := c.Bind(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	if err := validator.ValidateStruct(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	problemStatus := entities.ProblemStatus("All")
	if params.Status != "" {
		problemStatus = entities.ProblemStatus(params.Status)
	}

	problems, err := controller.problemsInteractor.GetProblems(ctx, userId, problemStatus)
	if err != nil {
		return err
	}

	result := make([]response.ProblemSummary, len(problems))
	for i, problem := range problems {
		result[i] = response.ProblemSummary{
			ProblemId:       problem.ProblemId,
			Title:           problem.Title,
			Status:          problem.Status,
			NeedsReviewFlag: problem.NeedsReviewFlag,
			CreatedAt:       problem.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, response.GetProblemsResponse{Result: result, Total: len(result)})
}

func (controller *ProblemsController) GetProblemDetail(c echo.Context) error {
	logger.Info("ProblemsController: GetProblemDetails")

	ctx := c.Request().Context()
	userId := middleware.UserIDFromContext(ctx)
	if userId == "" {
		err := response.NewUnauthorized(errors.New("invalid user"))
		return err
	}

	var params entities.GetProblemDetailParams
	if err := c.Bind(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	if err := validator.ValidateStruct(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	problem, err := controller.problemsInteractor.GetProblemDetails(ctx, userId, params.ProblemId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.ProblemDetail{
		ProblemId:       problem.ProblemId,
		RawProblem:      problem.RawProblem,
		Title:           problem.Title,
		ProblemText:     problem.ProblemText,
		AlgorithmTag:    problem.AlgorithmTag,
		Difficulty:      problem.Difficulty,
		AiHelp:          problem.AiHelp,
		NeedsReviewFlag: problem.NeedsReviewFlag,
		CreatedAt:       problem.CreatedAt,
		UpdatedAt:       problem.UpdatedAt,
	})
}

func (controller *ProblemsController) GetTodaysProblem(c echo.Context) error {
	logger.Info("ProblemsController: GetTodaysProblem")

	ctx := c.Request().Context()
	userId := middleware.UserIDFromContext(ctx)
	if userId == "" {
		err := response.NewUnauthorized(errors.New("invalid user"))
		return err
	}

	problem, err := controller.problemsInteractor.GetTodaysProblem(ctx, userId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.TodaysProblemResponse{
		Result: response.TodaysProblem{
			ProblemId:       problem.ProblemId,
			RawProblem:      problem.RawProblem,
			Title:           problem.Title,
			ProblemText:     problem.ProblemText,
			AlgorithmTag:    problem.AlgorithmTag,
			Difficulty:      problem.Difficulty,
			Status:          problem.Status,
			AiHelp:          problem.AiHelp,
			NeedsReviewFlag: problem.NeedsReviewFlag,
			CreatedAt:       problem.CreatedAt,
			UpdatedAt:       problem.UpdatedAt,
		},
	})
}

func (controller *ProblemsController) GetAIHelp(c echo.Context) error {
	logger.Info("ProblemsController: GetAIHelp")

	ctx := c.Request().Context()
	userId := middleware.UserIDFromContext(ctx)
	if userId == "" {
		err := response.NewUnauthorized(errors.New("invalid user"))
		return err
	}

	var params entities.GetProblemDetailParams
	if err := c.Bind(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	if err := validator.ValidateStruct(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	aiHelp, err := controller.problemsInteractor.GetAIHelp(ctx, userId, params.ProblemId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.AIHelp{
		Concept:     aiHelp.Concept,
		Nudge:       aiHelp.Nudge,
		Approach:    aiHelp.Approach,
		Walkthrough: aiHelp.Walkthrough,
	})
}
