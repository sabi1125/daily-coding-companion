package controller

import (
	"errors"
	"net/http"

	"backend/internal/domain/entities"
	"backend/internal/domain/interactor/inputport"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/validator"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authInteractor      inputport.AuthInteractorInputPort
	callbackRedirectUrl string
}

func NewAuthController(authInteractor inputport.AuthInteractorInputPort, callbackRedirectUrl string) *AuthController {
	return &AuthController{
		authInteractor:      authInteractor,
		callbackRedirectUrl: callbackRedirectUrl,
	}
}

func (controller *AuthController) SignIn(c echo.Context) error {
	logger.Info("AuthController: SignIn")

	ctx := c.Request().Context()
	authUrl, csrf, err := controller.authInteractor.SignIn(ctx)
	if err != nil {
		return err
	}

	// set cookie for checking with callback
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    csrf,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})

	return c.Redirect(http.StatusFound, authUrl)
}

func (controller *AuthController) Callback(c echo.Context) error {
	logger.Info("AuthController: Callback")
	ctx := c.Request().Context()

	// google callback params
	var params entities.GoogleCallbackParameters

	// 1. get query params sent from google
	if err := c.Bind(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	if err := validator.ValidateStruct(&params); err != nil {
		err = response.NewBadRequest(err)
		return err
	}

	cookie, err := c.Cookie("oauth_state")
	if err != nil {
		err = response.NewUnauthorized(err)
		return err
	}

	// oauth_state is a single-use CSRF nonce — clear it as soon as it's read
	// so it can't be replayed, regardless of whether the check below passes.
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	// 2. check if state matches
	if cookie.Value != params.State {
		err = response.NewUnauthorized(errors.New("cookie state and state returned by google do not match"))
		return err
	}

	sessionId, expiresAt, err := controller.authInteractor.Callback(ctx, params.Code)
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusFound, controller.callbackRedirectUrl)
}
