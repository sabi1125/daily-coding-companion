package controller

import (
	"net/http"

	"backend/internal/domain/interactor/inputport"
	logger "backend/internal/log"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authInteractor inputport.AuthInteractorInputPort
}

func NewAuthController(authInteractor inputport.AuthInteractorInputPort) *AuthController {
	return &AuthController{
		authInteractor: authInteractor,
	}
}

func (controller *AuthController) SignIn(c echo.Context) error {
	logger.Info("AuthControlled: SignIn")

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
