package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	interactorMock "backend/internal/domain/interactor/inputport/mock"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/validator"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	logger.InitForTest()
	validator.Init()
	m.Run()
}

// newTestEcho wires the same HTTPErrorHandler main.go uses, so a handler
// returning a *response.AppError actually maps to the right status code in
// these tests, same as it would in the real running server.
func newTestEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = response.ErrorHandler
	return e
}

func TestAuthController_SignIn(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(m *interactorMock.MockAuthInteractorInputPort)
		expectedStatus int
	}{
		{
			name: "redirects to the google auth url",
			mockSetup: func(m *interactorMock.MockAuthInteractorInputPort) {
				m.EXPECT().SignIn(gomock.Any()).Return("https://accounts.google.com/o/oauth2/v2/auth?...", "csrf-state", nil)
			},
			expectedStatus: http.StatusFound,
		},
		{
			name: "interactor error propagates through response.ErrorHandler",
			mockSetup: func(m *interactorMock.MockAuthInteractorInputPort) {
				m.EXPECT().SignIn(gomock.Any()).Return("", "", response.NewInternalError(nil))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockAuthInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewAuthController(mockInteractor, "http://localhost:8080/health")
			e.GET("/auth/google", controller.SignIn)

			req := httptest.NewRequest(http.MethodGet, "/auth/google", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAuthController_Callback(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		cookieValue    string
		omitCookie     bool
		mockSetup      func(m *interactorMock.MockAuthInteractorInputPort)
		expectedStatus int
	}{
		{
			name:        "success — state matches, sets session cookie, redirects",
			query:       "code=auth-code&state=matching-state",
			cookieValue: "matching-state",
			mockSetup: func(m *interactorMock.MockAuthInteractorInputPort) {
				m.EXPECT().Callback(gomock.Any(), "auth-code").Return("session-id", time.Now().AddDate(0, 0, 30), nil)
			},
			expectedStatus: http.StatusFound,
		},
		{
			name:           "missing code/state — validator rejects with 400",
			query:          "",
			cookieValue:    "matching-state",
			mockSetup:      func(m *interactorMock.MockAuthInteractorInputPort) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing oauth_state cookie — 401",
			query:          "code=auth-code&state=matching-state",
			omitCookie:     true,
			mockSetup:      func(m *interactorMock.MockAuthInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "state mismatch — 401",
			query:          "code=auth-code&state=matching-state",
			cookieValue:    "different-state",
			mockSetup:      func(m *interactorMock.MockAuthInteractorInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "interactor error propagates",
			query:       "code=auth-code&state=matching-state",
			cookieValue: "matching-state",
			mockSetup: func(m *interactorMock.MockAuthInteractorInputPort) {
				m.EXPECT().Callback(gomock.Any(), "auth-code").Return("", time.Time{}, response.NewBadGateway(nil))
			},
			expectedStatus: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockAuthInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewAuthController(mockInteractor, "http://localhost:8080/health")
			e.GET("/auth/google/callback", controller.Callback)

			req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?"+tt.query, nil)
			if !tt.omitCookie {
				req.AddCookie(&http.Cookie{Name: "oauth_state", Value: tt.cookieValue})
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestAuthController_Callback_ClearsStateCookie asserts oauth_state — a
// single-use CSRF nonce — is cleared as soon as it's read, on both the
// match and mismatch path, not just left to expire on its own.
func TestAuthController_Callback_ClearsStateCookie(t *testing.T) {
	tests := []struct {
		name        string
		cookieValue string
		mockSetup   func(m *interactorMock.MockAuthInteractorInputPort)
	}{
		{
			name:        "state matches",
			cookieValue: "matching-state",
			mockSetup: func(m *interactorMock.MockAuthInteractorInputPort) {
				m.EXPECT().Callback(gomock.Any(), "auth-code").Return("session-id", time.Now().AddDate(0, 0, 30), nil)
			},
		},
		{
			name:        "state mismatch",
			cookieValue: "different-state",
			mockSetup:   func(m *interactorMock.MockAuthInteractorInputPort) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockInteractor := interactorMock.NewMockAuthInteractorInputPort(ctrl)
			tt.mockSetup(mockInteractor)

			e := newTestEcho()
			controller := NewAuthController(mockInteractor, "http://localhost:8080/health")
			e.GET("/auth/google/callback", controller.Callback)

			req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=auth-code&state=matching-state", nil)
			req.AddCookie(&http.Cookie{Name: "oauth_state", Value: tt.cookieValue})
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			var cleared *http.Cookie
			for _, c := range rec.Result().Cookies() {
				if c.Name == "oauth_state" {
					cleared = c
				}
			}

			if assert.NotNil(t, cleared, "expected oauth_state to be re-set in the response") {
				assert.Less(t, cleared.MaxAge, 0)
			}
		})
	}
}
