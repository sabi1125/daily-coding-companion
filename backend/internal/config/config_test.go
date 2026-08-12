package config

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLoad(t *testing.T) {
	// setup
	t.Setenv("GO_ENV", "test")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("PORT", "8080")

	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("expected port 8080 but got %s", cfg.Port)
	}
}

func TestLoadDBConfig(t *testing.T) {
	// setup
	t.Setenv("GO_ENV", "test")
	t.Setenv("DB_USER", "dailycodingcompanion")
	t.Setenv("DB_PASSWORD", "dailycodingcompanionpass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "dailycodingcompanion")

	dbCfg := LoadDbConfig()

	if dbCfg.DBUser != "dailycodingcompanion" {
		t.Errorf("expected port 8080 but got %s", dbCfg.DBUser)
	}

	if dbCfg.DBPassword != "dailycodingcompanionpass" {
		t.Errorf("expected password dailycodingcompanionpass but got %s", dbCfg.DBPassword)
	}

	if dbCfg.DBHost != "localhost" {
		t.Errorf("expected htt localhost but got %s", dbCfg.DBHost)
	}

	if dbCfg.DBPort != "3306" {
		t.Errorf("expected DBPort 3306 but got %s", dbCfg.DBPort)
	}

	if dbCfg.DBName != "dailycodingcompanion" {
		t.Errorf("expected DBName dailycodingcompanion but got %s", dbCfg.DBName)
	}
}

// TestLoadZapConfig
func TestLoadZapConfig(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		zapLogLevel zapcore.Level
	}{
		{
			name:        "WHEN LEVEL IS INFO",
			logLevel:    "INFO",
			zapLogLevel: zapcore.InfoLevel,
		},
		{
			name:        "WHEN LEVEL IS DEBUG",
			logLevel:    "DEBUG",
			zapLogLevel: zapcore.DebugLevel,
		},
		{
			name:        "WHEH LEVEL IS NOT GIVEN",
			logLevel:    "",
			zapLogLevel: zapcore.InfoLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			t.Setenv("GO_ENV", "test")
			t.Setenv("LOG_LEVEL", tt.logLevel)
			zapCfg := LoadZapConfig()

			if zapCfg.LogLevel != tt.zapLogLevel {
				t.Errorf("expected log level to be %s but got %s", tt.zapLogLevel, zapCfg.LogLevel)
			}
		})
	}
}
