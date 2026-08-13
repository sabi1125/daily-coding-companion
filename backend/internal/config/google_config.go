package config

import (
	"log"
	"os"
)

type GoogleConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectUrl        string
}

func LoadGoogleConfigFromEnv() *GoogleConfig {
	LoadEnv()

	// gets google client id if provided kills process with logger.fatal
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID environment variable is not set")
	}

	// gets google client secret if provided kills process with logger.fatal
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET environment variable is not set")
	}

	// gets redirect url if provided kills process with logger.fatal
	redirectUrl := os.Getenv("REDIRECT_URL")
	if redirectUrl == "" {
		log.Fatal("REDIRECT_URL environment variable is not set")
	}

	googleConfig := &GoogleConfig{
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		RedirectUrl:        redirectUrl,
	}

	return googleConfig
}
