package infrastructure

import (
	"backend/internal/config"
	"backend/internal/controller"
	ingestrunner "backend/internal/domain/ingest_runner"
	"backend/internal/domain/interactor"
	"backend/internal/domain/repository"
	"backend/internal/infrastructure/middleware"
	"backend/internal/tx"
	"backend/internal/util"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func Router(e *echo.Echo, db *gorm.DB) {
	// shared
	userRepository := repository.NewUsersRepository(db)
	oauthRepository := repository.NewOauthRepository(db)
	settingRepository := repository.NewSettingsRepository(db)
	sessionRepository := repository.NewSessionsRepository(db)
	problemRepository := repository.NewProblemsRepository(db)
	ingestRepository := repository.NewIngestRepository(db)

	// config
	googleCfg := config.LoadGoogleConfigFromEnv()
	googleOauthCfg := config.LoadOauthConfig(googleCfg)
	claudeCfg := config.LoadClaudeConfigFromEnv()

	// client
	claudeClient := anthropic.NewClient(option.WithAPIKey(claudeCfg.APIKey))

	// uuid generator
	uuidGenerator := util.NewUUIDGenerator()

	txManager := NewTransactionManager(db)

	// create ingest runner
	ingestRunner := ingestrunner.NewIngestRunner(
		uuidGenerator,
		googleOauthCfg,
		ingestRepository,
		oauthRepository,
		problemRepository,
		claudeClient,
		txManager,
		db,
	)

	// routes
	RegisteredHealthRouter(e, db)
	RegisteredAuthRoutes(e, db, userRepository, oauthRepository, settingRepository, sessionRepository, *googleCfg, googleOauthCfg, uuidGenerator, txManager)
	RegisteredSettingsRoutes(e, db, sessionRepository, txManager)
	RegisteredProblemsRoutes(e, db, problemRepository, sessionRepository, ingestRunner, ingestRepository)
	RegisteredSubmittedSolutionsRoutes(e, db, sessionRepository, problemRepository, uuidGenerator, txManager)
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
	googleCfg config.GoogleConfig,
	googleOauthCfg *oauth2.Config,
	uuidGenerator util.UUIDGenerator,
	txManager tx.Manager,
) {
	auth := e.Group("/auth")
	repository := repository.NewAuthRepository(db)
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
	ingestRunner *ingestrunner.IngestRunner,
	ingestRepository *repository.IngestRepository,
) {
	problems := e.Group("/problems")
	problems.Use(middleware.Auth(sessionRepository))
	interactor := interactor.NewProblemsInteractor(repository, ingestRunner, ingestRepository)
	controller := controller.NewProblemsController(interactor)

	problems.GET("", controller.GetProblems)
	problems.GET("/:id", controller.GetProblemDetail)
	problems.GET("/today", controller.GetTodaysProblem)
}

func RegisteredSubmittedSolutionsRoutes(
	e *echo.Echo,
	db *gorm.DB,
	sessionRepository *repository.SessionsRepository,
	problemRepository *repository.ProblemsRepository,
	uuidGenerator util.UUIDGenerator,
	txManager tx.Manager,
) {
	problems := e.Group("/submissions")
	problems.Use(middleware.Auth(sessionRepository))
	repository := repository.NewSubmittedSolutionsRepository(db)
	interactor := interactor.NewSubmittedSolutionInteractor(repository, problemRepository, uuidGenerator, txManager)
	controller := controller.NewSubmittedSolutionsController(interactor)

	problems.GET("/:id", controller.GetUserSubmissions)
	problems.POST("/:id", controller.SubmitSolutions)
}
