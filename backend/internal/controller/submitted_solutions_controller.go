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

type SubmittedSolutionsController struct {
	submittedSolutionsInteractor inputport.SubmittedSolutionsInteractorInputPort
}

func NewSubmittedSolutionsController(submittedSolutionsInteractor inputport.SubmittedSolutionsInteractorInputPort) *SubmittedSolutionsController {
	return &SubmittedSolutionsController{
		submittedSolutionsInteractor: submittedSolutionsInteractor,
	}
}

func (controller *SubmittedSolutionsController) GetUserSubmissions(c echo.Context) error {
	logger.Info("SubmissionsController: GetUserSubmissions")

	ctx := c.Request().Context()

	userId := middleware.UserIDFromContext(ctx)
	if userId == "" {
		err := response.NewUnauthorized(errors.New("invalid user"))
		return err
	}

	var params entities.GetSubmittedSolutionsParam
	if err := c.Bind(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	if err := validator.ValidateStruct(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	solutions, err := controller.submittedSolutionsInteractor.GetSubmittedSolutions(ctx, userId, params.ProblemId)
	if err != nil {
		return err
	}

	result := make([]response.Submission, len(solutions))
	for i, solution := range solutions {
		result[i] = response.Submission{
			SolutionId:  solution.SolutionId,
			ProblemId:   solution.ProblemId,
			Solution:    solution.Solution,
			Status:      solution.Status,
			SubmittedAt: solution.SubmittedAt,
		}
	}

	return c.JSON(http.StatusOK, response.GetSubmissionsResponse{Result: result, Total: len(result)})
}
