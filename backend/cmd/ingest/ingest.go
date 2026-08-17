package ingest

import (
	"context"

	"backend/internal/config"
	ingestrunner "backend/internal/domain/ingest_runner"
	"backend/internal/domain/repository"
	"backend/internal/infrastructure"
	"backend/internal/util"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"gorm.io/gorm"
)

func Ingest(db *gorm.DB) error {
	ctx := context.Background()

	// config setup
	googleCfg := config.LoadGoogleConfigFromEnv()
	claudeCfg := config.LoadClaudeConfigFromEnv()
	oauthCfg := config.LoadOauthConfig(googleCfg)

	// client setup
	claudeClient := anthropic.NewClient(option.WithAPIKey(claudeCfg.APIKey))

	// required repository setup
	ingestRepository := *repository.NewIngestRepository(db)
	oauthRepository := *repository.NewOauthRepository(db)
	problemsRepository := *repository.NewProblemsRepository(db)

	// managers
	uuidGenerator := util.NewUUIDGenerator()
	txManager := infrastructure.NewTransactionManager(db)

	// get all userIds required to run daily ingest
	userIds, err := oauthRepository.GetAllUserIds(ctx)
	if err != nil {
		return err
	}

	// create new ingest
	ingest := ingestrunner.NewIngestRunner(
		uuidGenerator,
		oauthCfg,
		&ingestRepository,
		&oauthRepository,
		&problemsRepository,
		claudeClient,
		txManager,
		db,
	)

	// run ingest with all userIds
	ingestErr := ingest.Ingest(ctx, userIds, false)
	if ingestErr != nil {
		return ingestErr
	}

	return nil
}
