package middleware

import (
	"log"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func CORS() echo.MiddlewareFunc {
	origins := os.Getenv("FRONTEND_ORIGINS")
	if origins == "" {
		log.Fatal("FRONTEND_ORIGINS environment variable is not set")
	}

	allowOrigins := strings.Split(origins, ",")
	for i := range allowOrigins {
		allowOrigins[i] = strings.TrimSpace(allowOrigins[i])
	}

	return echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowCredentials: true,
	})
}
