package infrastructure

import (
	"fmt"

	"backend/internal/config"
	logger "backend/internal/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connection(cfg *config.DBConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Asia%%2FTokyo",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	logger.Info("connecting to the database...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("failed to connect to the database: %v", err)
	}

	logger.Info("connected to the database")
	return db
}
