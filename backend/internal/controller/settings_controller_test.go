package controller

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/domain/entities"
	interactorMock "backend/internal/domain/interactor/inputport/mock"
	"backend/internal/infrastructure/middleware"
	"backend/internal/response"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// withUserID stands in for the Auth middleware — it stashes user_id in the
// request context the same way middleware.Auth does, so these tests can
// exercise the handler without going through a real session lookup.
func withUserID(req *http.Request, userId string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userId)
	return req.WithContext(ctx)
}

func TestSettingsController_GetUserSettings(t *testing.T) {
	tests := []struct {
		name           string
		omitUserID     bool
		mockSetup      func(m *interactorMock.MockSettingsInteractorInputPort)
		expectedStatus int
	}{
		{
			name: "success — returns setting",
			mockSetup: func(m *interactorMock.MockSettingsInteractorInputPort) {
				m.EXPECT().GetUserSetting(gomock.Any(), "user-1", gomock.Any()).Return(entities.Settings{SettingID: "setting-1"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when user id is missing",
			omitUserID:     true,
			mockSetup:      func(m *interactorMock.MockSettingsInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unexpected error",
			mockSetup: func(m *interactorMock.MockSettingsInteractorInputPort) {
				m.EXPECT().GetUserSetting(gomock.Any(), "user-1", gomock.Any()).Return(entities.Settings{}, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "session older than 7 days — needs_reauth true in response",
			mockSetup: func(m *interactorMock.MockSettingsInteractorInputPort) {
				m.EXPECT().GetUserSetting(gomock.Any(), "user-1", gomock.Any()).
					Return(entities.Settings{SettingID: "setting-1", NeedsReauth: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockSettingsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewSettingsController(mockInteractor)
			e.GET("/settings", controller.GetUserSettings)

			req := httptest.NewRequest(http.MethodGet, "/settings", nil)
			if !tt.omitUserID {
				req = withUserID(req, "user-1")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSettingsController_UpdateUserSettings(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *interactorMock.MockSettingsInteractorInputPort)
		expectedStatus int
	}{
		{
			name: "success — updates setting",
			body: `{"get_help_preferences":"new preferences"}`,
			mockSetup: func(m *interactorMock.MockSettingsInteractorInputPort) {
				m.EXPECT().UpdateUserSetting(gomock.Any(), "user-1", "new preferences", gomock.Any()).
					Return(entities.Settings{SettingID: "setting-1"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when error is empty",
			body:           `{"get_help_preferences":""}`,
			mockSetup:      func(m *interactorMock.MockSettingsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "when error exceeds 1000 characters",
			body:           `{"get_help_preferences":"` + string(bytes.Repeat([]byte("a"), 1001)) + `"}`,
			mockSetup:      func(m *interactorMock.MockSettingsInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unexpected error",
			body: `{"get_help_preferences":"new preferences"}`,
			mockSetup: func(m *interactorMock.MockSettingsInteractorInputPort) {
				m.EXPECT().UpdateUserSetting(gomock.Any(), "user-1", "new preferences", gomock.Any()).
					Return(entities.Settings{}, response.NewDatabaseError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockSettingsInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewSettingsController(mockInteractor)
			e.PATCH("/settings", controller.UpdateUserSettings)

			req := httptest.NewRequest(http.MethodPatch, "/settings", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, "user-1")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
