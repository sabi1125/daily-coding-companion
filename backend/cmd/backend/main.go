package main

import (
	"os"
	_ "time/tzdata"

	"backend/cmd/ingest"
	"backend/internal/config"
	"backend/internal/infrastructure"
	"backend/internal/infrastructure/middleware"
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

	args := os.Args

	if len(args) == 2 {
		if args[1] == "ingest" {
			ingest.Ingest(db)
			return
		} else {
			logger.Fatal("invalid argument")
			return
		}
	} else if len(args) > 1 {
		logger.Fatal("invalid number of arguments")
		return
	}

	// Create an Echo instance
	e := echo.New()
	e.HTTPErrorHandler = response.ErrorHandler
	e.Use(middleware.CORS())
	e.Use(logger.MiddlewareLogger(logger.Get()))
	infrastructure.Router(e, db)

	// Start the server
	logger.Info("starting server")
	e.Logger.Fatal(e.Start(":8080"))
}
