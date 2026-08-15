package infrastructure

import (
	"backend/internal/config"
	"backend/internal/controller"
	"backend/internal/domain/interactor"
	"backend/internal/domain/repository"
	"backend/internal/util"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func Router(e *echo.Echo, db *gorm.DB) {
	// shared
	userRepository := repository.NewUsersRepository(db)
	oauthRepository := repository.NewOauthRepository(db)
	settingRepository := repository.NewSettingsRepository(db)
	sessionRepository := repository.NewSessionsRepository(db)

	RegisteredHealthRouter(e, db)
	RegisteredAuthRoutes(e, db, userRepository, oauthRepository, settingRepository, sessionRepository)
}

func RegisteredHealthRouter(e *echo.Echo, db *gorm.DB) {
	health := e.Group("/health")
	repository := repository.NewHealthRepository(db)
	interactor := interactor.NewHealthInteractor(repository)
	controller := controller.NewHealthController(interactor)

	health.GET("", controller.Health)
}

func RegisteredAuthRoutes(
	e *echo.Echo,
	db *gorm.DB,
	userRepository *repository.UsersRepository,
	oauthRepository *repository.OauthRepository,
	settingRepository *repository.SettingsRepository,
	sessionRepository *repository.SessionsRepository,
) {
	googleCfg := config.LoadGoogleConfigFromEnv()
	googleOauthCfg := config.LoadOauthConfig(googleCfg)

	auth := e.Group("/auth")
	repository := repository.NewAuthRepository(db)
	uuidGenerator := util.NewUUIDGenerator()
	txManager := NewTransactionManager(db)
	interactor := interactor.NewAuthInteractor(
		*googleOauthCfg,
		repository,
		userRepository,
		oauthRepository,
		settingRepository,
		sessionRepository,
		uuidGenerator,
		txManager,
	)
	controller := controller.NewAuthController(interactor, googleCfg.CallbackRedirectUrl)

	auth.GET("/google", controller.SignIn)
	auth.GET("/google/callback", controller.Callback)
}
