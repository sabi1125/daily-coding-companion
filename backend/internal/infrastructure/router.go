package infrastructure

import (
	"backend/internal/config"
	"backend/internal/controller"
	"backend/internal/domain/interactor"
	"backend/internal/domain/repository"
	"backend/internal/infrastructure/middleware"
	"backend/internal/tx"
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
	problemRepository := repository.NewProblemsRepository(db)

	txManager := NewTransactionManager(db)

	RegisteredHealthRouter(e, db)
	RegisteredAuthRoutes(e, db, userRepository, oauthRepository, settingRepository, sessionRepository, txManager)
	RegisteredSettingsRoutes(e, db, sessionRepository, txManager)
	RegisteredProblemsRoutes(e, db, problemRepository, sessionRepository)
	RegisteredSubmittedSolutionsRoutes(e, db, sessionRepository, problemRepository, txManager)
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
	txManager tx.Manager,
) {
	googleCfg := config.LoadGoogleConfigFromEnv()
	googleOauthCfg := config.LoadOauthConfig(googleCfg)

	auth := e.Group("/auth")
	repository := repository.NewAuthRepository(db)
	uuidGenerator := util.NewUUIDGenerator()
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
	auth.POST("/signout", controller.Signout, middleware.Auth(sessionRepository))
}

func RegisteredSettingsRoutes(
	e *echo.Echo,
	db *gorm.DB,
	sessionRepository *repository.SessionsRepository,
	txManager tx.Manager,
) {
	settings := e.Group("/settings")
	settings.Use(middleware.Auth(sessionRepository))
	repository := repository.NewSettingsRepository(db)
	interactor := interactor.NewSettingsInteractor(repository, txManager)
	controller := controller.NewSettingsController(interactor)

	settings.GET("", controller.GetUserSettings)
	settings.PATCH("", controller.UpdateUserSettings)
}

func RegisteredProblemsRoutes(
	e *echo.Echo,
	db *gorm.DB,
	repository *repository.ProblemsRepository,
	sessionRepository *repository.SessionsRepository,
) {
	problems := e.Group("/problems")
	problems.Use(middleware.Auth(sessionRepository))
	interactor := interactor.NewProblemsInteractor(repository)
	controller := controller.NewProblemsController(interactor)

	problems.GET("", controller.GetProblems)
	problems.GET("/:id", controller.GetProblemDetail)
}

func RegisteredSubmittedSolutionsRoutes(
	e *echo.Echo,
	db *gorm.DB,
	sessionRepository *repository.SessionsRepository,
	problemRepository *repository.ProblemsRepository,
	txManager tx.Manager,
) {
	problems := e.Group("/submissions")
	problems.Use(middleware.Auth(sessionRepository))
	repository := repository.NewSubmittedSolutionsRepository(db)
	uuidGenerator := util.NewUUIDGenerator()
	interactor := interactor.NewSubmittedSolutionInteractor(repository, problemRepository, uuidGenerator, txManager)
	controller := controller.NewSubmittedSolutionsController(interactor)

	problems.GET("/:id", controller.GetUserSubmissions)
	problems.POST("/:id", controller.SubmitSolutions)
}
