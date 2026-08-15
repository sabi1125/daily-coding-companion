package config

import (
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type GoogleConfig struct {
	GoogleClientID      string
	GoogleClientSecret  string
	RedirectUrl         string
	CallbackRedirectUrl string
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

	callbackRedirectUrl := os.Getenv("CALLBACK_REDIRECT_URL")
	if callbackRedirectUrl == "" {
		log.Fatal("CALLBACK_REDIRECT_URL environment variable is not set")
	}

	googleConfig := &GoogleConfig{
		GoogleClientID:      googleClientID,
		GoogleClientSecret:  googleClientSecret,
		RedirectUrl:         redirectUrl,
		CallbackRedirectUrl: callbackRedirectUrl,
	}

	return googleConfig
}

func LoadOauthConfig(googleCfg *GoogleConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     googleCfg.GoogleClientID,
		ClientSecret: googleCfg.GoogleClientSecret,
		RedirectURL:  googleCfg.RedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			gmail.GmailReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}
}
