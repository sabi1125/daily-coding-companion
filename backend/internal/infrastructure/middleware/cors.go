package middleware

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func CORS() echo.MiddlewareFunc {
	origins := os.Getenv("FRONTEND_ORIGINS")
	if origins == "" {
		log.Fatal("FRONTEND_ORIGINS environment variable is not set")
	}

	return echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{origins},
		AllowCredentials: true,
	})
}
