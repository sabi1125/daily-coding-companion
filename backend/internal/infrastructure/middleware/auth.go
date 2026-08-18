package middleware

import (
	"context"
	"errors"
	"time"

	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/response"

	"github.com/labstack/echo/v4"
)

type ctxKey string

const (
	UserIDKey           ctxKey = "user_id"
	SessionCreatedAtKey ctxKey = "session_created_at"
)

func Auth(sessionsRepository inputport.SessionsRepositoryInputPort) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger.Info("Middleware: Auth")

			cookie, err := c.Cookie("session_id")
			if err != nil {
				return response.NewUnauthorized(err)
			}

			session, err := sessionsRepository.GetSessionById(c.Request().Context(), cookie.Value)
			if err != nil {
				return err
			}
			if session == nil || session.ExpiresAt.Before(time.Now()) {
				return response.NewUnauthorized(errors.New("session invalid or expired"))
			}

			ctx := context.WithValue(c.Request().Context(), UserIDKey, session.UserId)
			ctx = context.WithValue(ctx, SessionCreatedAtKey, session.CreatedAt)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func UserIDFromContext(ctx context.Context) string {
	userId, _ := ctx.Value(UserIDKey).(string)
	return userId
}

func SessionCreatedAtFromContext(ctx context.Context) time.Time {
	createdAt, _ := ctx.Value(SessionCreatedAtKey).(time.Time)
	return createdAt
}
