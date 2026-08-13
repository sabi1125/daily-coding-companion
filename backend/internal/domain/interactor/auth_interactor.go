package interactor

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"backend/internal/config"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/response"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type AuthInteractor struct {
	googleCfg      config.GoogleConfig
	authRepository inputport.AuthRepositoryInputPort
}

func NewAuthInteractor(authRepository inputport.AuthRepositoryInputPort, googleCfg config.GoogleConfig) *AuthInteractor {
	return &AuthInteractor{
		googleCfg:      googleCfg,
		authRepository: authRepository,
	}
}

func (interactor *AuthInteractor) SignIn(ctx context.Context) (authUrl string, csrf string, err error) {
	logger.Info("AuthInteractor: SignIn")

	// create oauth config
	oauthConfig := &oauth2.Config{
		ClientID:     interactor.googleCfg.GoogleClientID,
		ClientSecret: interactor.googleCfg.GoogleClientSecret,
		RedirectURL:  interactor.googleCfg.RedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			gmail.GmailReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}

	csrf, err = generateState()
	if err != nil {
		err = response.NewInternalError(err)
		return
	}

	// create the oauth url with csrf token
	authUrl = oauthConfig.AuthCodeURL(csrf)

	return
}

func generateState() (csrf string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
