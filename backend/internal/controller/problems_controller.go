package controller

import (
	"errors"
	"net/http"
	"time"

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

// problemSummary is the History-list shape — only what the list view
// renders. Full problem content (raw_problem, problem_text, algorithm_tag,
// difficulty, ai_help) is what GET /problems/{id} is for.
type problemSummary struct {
	ProblemId       string    `json:"problem_id"`
	Title           *string   `json:"title"`
	Status          string    `json:"status"`
	NeedsReviewFlag bool      `json:"needs_review_flag"`
	CreatedAt       time.Time `json:"created_at"`
}

type getProblemsResponse struct {
	Result []problemSummary `json:"result"`
	Total  int              `json:"total"`
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

	result := make([]problemSummary, len(problems))
	for i, problem := range problems {
		result[i] = problemSummary{
			ProblemId:       problem.ProblemId,
			Title:           problem.Title,
			Status:          problem.Status,
			NeedsReviewFlag: problem.NeedsReviewFlag,
			CreatedAt:       problem.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, getProblemsResponse{Result: result, Total: len(result)})
}
