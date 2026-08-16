package interactor

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"backend/internal/domain/entities"
	"backend/internal/domain/repository/inputport"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"
	"backend/internal/util"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type AuthInteractor struct {
	googleOauthCfg    oauth2.Config
	authRepository    inputport.AuthRepositoryInputPort
	userRepository    inputport.UsersRepositoryInputPort
	oauthRepository   inputport.OauthRepositoryInputPort
	settingRepository inputport.SettingsRepositoryInputPort
	sessionRepository inputport.SessionsRepositoryInputPort
	uuidGenerator     util.UUIDGenerator
	txManager         tx.Manager
}

func NewAuthInteractor(
	googleOauthCfg oauth2.Config,
	authRepository inputport.AuthRepositoryInputPort,
	userRepository inputport.UsersRepositoryInputPort,
	oauthRepository inputport.OauthRepositoryInputPort,
	settingRepository inputport.SettingsRepositoryInputPort,
	sessionRepository inputport.SessionsRepositoryInputPort,
	uuidGenerator util.UUIDGenerator,
	txManager tx.Manager,
) *AuthInteractor {
	return &AuthInteractor{
		googleOauthCfg:    googleOauthCfg,
		authRepository:    authRepository,
		userRepository:    userRepository,
		oauthRepository:   oauthRepository,
		settingRepository: settingRepository,
		sessionRepository: sessionRepository,
		uuidGenerator:     uuidGenerator,
		txManager:         txManager,
	}
}

func (interactor *AuthInteractor) SignIn(ctx context.Context) (authUrl string, csrf string, err error) {
	logger.Info("AuthInteractor: SignIn")

	csrf, err = generateState()
	if err != nil {
		err = response.NewInternalError(err)
		return
	}

	authUrl = interactor.googleOauthCfg.AuthCodeURL(csrf, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	return
}

func (interactor *AuthInteractor) Callback(ctx context.Context, code string) (sessionId string, expiresAt time.Time, err error) {
	logger.Info("AuthInteractor: Callback")

	// 3. exchange the token
	token, err := interactor.googleOauthCfg.Exchange(ctx, code)
	if err != nil {
		var receivedErr *oauth2.RetrieveError
		if errors.As(err, &receivedErr) {
			if receivedErr.ErrorCode == "invalid_grant" {
				err = response.NewBadRequest(err)
				return
			}
			err = response.NewBadGateway(err)
			return
		}
		err = response.NewServiceUnavailable(err)
		return
	}

	// get id token
	extractedIdToken, ok := token.Extra("id_token").(string)
	if !ok || extractedIdToken == "" {
		err = response.NewDatabaseError(errors.New("id_token missing or wrong type in token response"))
		return
	}

	// 4. verify id_token
	payload, err := idtoken.Validate(ctx, extractedIdToken, interactor.googleOauthCfg.ClientID)
	if err != nil {
		err = response.NewUnauthorized(err)
		return
	}

	identity, err := extractGoogleIdentity(payload)
	if err != nil {
		err = response.NewDatabaseError(err)
		return
	}

	// 5. check if returning user
	oauthCredentials, err := interactor.oauthRepository.FindUserBySub(ctx, identity.Sub)
	if err != nil {
		return
	}

	var userId string

	if oauthCredentials != nil {
		// 5.a returning user: update refresh token, expiry_at
		userId = oauthCredentials.UserId
		err = interactor.oauthRepository.UpdateOauthInformationWithSub(ctx, identity.Sub, token.RefreshToken, time.Unix(payload.Expires, 0))
		if err != nil {
			return
		}
	} else {
		// 5.b first-time sign-in: create user + oauth_credentials + settings together
		userId, err = interactor.createUserRecords(ctx, identity, token, payload.Expires)
		if err != nil {
			return
		}
	}

	sessionId, expiresAt, err = interactor.createSession(ctx, userId)
	if err != nil {
		return
	}

	return
}

func (interactor *AuthInteractor) createUserRecords(ctx context.Context, identity entities.GoogleIdentity, token *oauth2.Token, expires int64) (string, error) {
	userId, err := interactor.uuidGenerator.NewV7()
	if err != nil {
		return "", response.NewInternalError(err)
	}

	settingId, err := interactor.uuidGenerator.NewV7()
	if err != nil {
		return "", response.NewInternalError(err)
	}

	err = interactor.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		user := &entities.Users{
			UserId:    userId,
			Email:     identity.Email,
			FirstName: identity.GivenName,
			LastName:  identity.FamilyName,
		}
		if err := interactor.userRepository.CreateUser(ctx, user); err != nil {
			return err
		}

		oauthCreds := &entities.OauthCredentials{
			OauthId:      identity.Sub,
			UserId:       userId,
			RefreshToken: token.RefreshToken,
			ExpiryAt:     time.Unix(expires, 0),
		}
		if err := interactor.oauthRepository.CreateOauthCredentials(ctx, oauthCreds); err != nil {
			return err
		}

		setting := &entities.Settings{
			SettingID: settingId,
			UserID:    userId,
		}
		if err := interactor.settingRepository.CreateSetting(ctx, setting); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (interactor *AuthInteractor) createSession(ctx context.Context, userId string) (string, time.Time, error) {
	sessionId, err := interactor.uuidGenerator.NewV7()
	if err != nil {
		return "", time.Time{}, response.NewInternalError(err)
	}

	timeProvider := util.NewTimeProvider()
	expiresAt := timeProvider.ExpiryTimeCalculator()

	session := &entities.Sessions{
		SessionId: sessionId,
		UserId:    userId,
		ExpiresAt: expiresAt,
	}
	if err := interactor.sessionRepository.CreateSession(ctx, session); err != nil {
		return "", time.Time{}, err
	}

	return sessionId, expiresAt, nil
}

func extractGoogleIdentity(payload *idtoken.Payload) (entities.GoogleIdentity, error) {
	marshaled, err := json.Marshal(payload.Claims)
	if err != nil {
		return entities.GoogleIdentity{}, err
	}

	var identity entities.GoogleIdentity
	if err := json.Unmarshal(marshaled, &identity); err != nil {
		return entities.GoogleIdentity{}, err
	}

	identity.Sub = payload.Subject
	return identity, nil
}

func generateState() (csrf string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (interactor *AuthInteractor) DeleteUserSession(ctx context.Context, userId string, sessionId string) (err error) {
	logger.Info("AuthInterector: DeleteUserSession")

	err = interactor.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		innerErr := interactor.sessionRepository.DeleteUserSession(ctx, sessionId, userId)
		return innerErr
	})
	if err != nil {
		return
	}

	return
}
