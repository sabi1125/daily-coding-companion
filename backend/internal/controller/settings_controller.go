package controller

import (
	"errors"
	"net/http"

	"backend/internal/domain/entities"
	"backend/internal/domain/interactor/inputport"
	"backend/internal/infrastructure/middleware"
	logger "backend/internal/log"
	"backend/internal/response"

	"github.com/labstack/echo/v4"
)

type SettingsController struct {
	settingsInteractor inputport.SettingsInteractorInputPort
}

func NewSettingsController(settingsInteractor inputport.SettingsInteractorInputPort) *SettingsController {
	return &SettingsController{
		settingsInteractor: settingsInteractor,
	}
}

func (controller *SettingsController) GetUserSettings(c echo.Context) error {
	logger.Info("SettingsController: GetUserSettings")

	ctx := c.Request().Context()
	userId := middleware.UserIDFromContext(ctx)
	if userId == "" {
		err := response.NewUnauthorized(errors.New("invalid user"))
		return err
	}

	setting, err := controller.settingsInteractor.GetUserSetting(ctx, userId)
	if err != nil {
		return err
	}

	res := &entities.Settings{
		SettingID:          setting.SettingID,
		GetHelpPreferences: setting.GetHelpPreferences,
	}

	return c.JSON(http.StatusOK, res)
}
