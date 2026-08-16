package config

import (
	"log"
	"os"
)

type ClaudeConfig struct {
	APIKey string
}

func LoadClaudeConfigFromEnv() *ClaudeConfig {
	LoadEnv()

	// gets anthropic api key if provided kills process with logger.fatal
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is not set")
	}

	return &ClaudeConfig{
		APIKey: apiKey,
	}
}
