package response

import (
	"errors"
	"net/http"

	logger "backend/internal/log"

	"github.com/labstack/echo/v4"
)

// ErrorHandler is Echo's central error hook (wired via e.HTTPErrorHandler in
// main.go, not e.Use — this isn't middleware). *AppError is already logged
// by the time it gets here (see newAppError in apperror.go — logging
// happens at construction, at the real call site, not centrally) — this
// only writes the client response. A bare error that was never wrapped as
// an *AppError is the one thing logged here: it means something returned a
// raw error instead of going through response.NewX, worth knowing about.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status.Code, map[string]string{"message": appErr.Status.Message})
		return
	}

	logger.Error(err)
	c.JSON(http.StatusInternalServerError, map[string]string{"message": InternalError.Message})
}
