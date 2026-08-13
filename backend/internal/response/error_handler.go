package response

import (
	"errors"
	"net/http"

	logger "backend/internal/log"

	"github.com/labstack/echo/v4"
)

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Status.Category != Expected {
			logger.Error(appErr)
		}
		c.JSON(appErr.Status.Code, map[string]string{"message": appErr.Status.Message})
		return
	}

	logger.Error(err)
	c.JSON(http.StatusInternalServerError, map[string]string{"message": InternalError.Message})
}
