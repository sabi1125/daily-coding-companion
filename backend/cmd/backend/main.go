package main

import (
	"backend/internal/config"
	"backend/internal/infrastructure"
	"backend/internal/log"
	"backend/internal/response"
	"backend/internal/validator"

	"github.com/labstack/echo/v4"
)

func main() {
	zapCfg := config.LoadZapConfig()
	logger.Init(zapCfg)
	defer logger.Sync()

	validator.Init()

	dbCfg := config.LoadDbConfig()
	db := infrastructure.Connection(dbCfg)

	// Create an Echo instance
	e := echo.New()
	e.HTTPErrorHandler = response.ErrorHandler
	e.Use(logger.MiddlewareLogger(logger.Get()))
	infrastructure.Router(e, db)

	// Start the server
	logger.Info("starting server")
	e.Logger.Fatal(e.Start(":8080"))
}
