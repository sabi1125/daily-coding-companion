package infrastructure

import (
	"backend/internal/config"
	"backend/internal/controller"
	"backend/internal/domain/interactor"
	"backend/internal/domain/repository"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Router(e *echo.Echo, db *gorm.DB) {
	RegisteredHealthRouter(e, db)
	RegisteredAuthRoutes(e, db)
}

func RegisteredHealthRouter(e *echo.Echo, db *gorm.DB) {
	health := e.Group("/health")
	repository := repository.NewHealthRepository(db)
	interactor := interactor.NewHealthInteractor(repository)
	controller := controller.NewHealthController(interactor)

	health.GET("", controller.Health)
}

func RegisteredAuthRoutes(e *echo.Echo, db *gorm.DB) {
	googleConfig := config.LoadGoogleConfigFromEnv()

	auth := e.Group("/auth")
	repository := repository.NewAuthRepository(db)
	interactor := interactor.NewAuthInteractor(repository, *googleConfig)
	controller := controller.NewAuthController(interactor)

	auth.GET("/google", controller.SignIn)
}
