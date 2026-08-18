package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/domain/entities"
	inputportMock "backend/internal/domain/repository/inputport/mock"
	logger "backend/internal/log"
	"backend/internal/response"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	logger.InitForTest()
	m.Run()
}

func TestAuth(t *testing.T) {
	sessionCreatedAt := time.Now().Add(-3 * 24 * time.Hour).Truncate(time.Second)

	tests := []struct {
		name           string
		omitCookie     bool
		cookieValue    string
		mockSetup      func(m *inputportMock.MockSessionsRepositoryInputPort)
		expectedStatus int
		wantNext       bool
	}{
		{
			name:        "valid session lets the request through",
			cookieValue: "session-1",
			mockSetup: func(m *inputportMock.MockSessionsRepositoryInputPort) {
				m.EXPECT().GetSessionById(gomock.Any(), "session-1").Return(&entities.Sessions{
					SessionId: "session-1",
					UserId:    "user-1",
					CreatedAt: sessionCreatedAt,
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			wantNext:       true,
		},
		{
			name:           "missing session_id cookie — 401",
			omitCookie:     true,
			mockSetup:      func(m *inputportMock.MockSessionsRepositoryInputPort) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "unknown session — 401",
			cookieValue: "unknown-session",
			mockSetup: func(m *inputportMock.MockSessionsRepositoryInputPort) {
				m.EXPECT().GetSessionById(gomock.Any(), "unknown-session").Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "expired session — 401",
			cookieValue: "session-1",
			mockSetup: func(m *inputportMock.MockSessionsRepositoryInputPort) {
				m.EXPECT().GetSessionById(gomock.Any(), "session-1").Return(&entities.Sessions{
					SessionId: "session-1",
					UserId:    "user-1",
					ExpiresAt: time.Now().Add(-time.Hour),
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "repository error propagates",
			cookieValue: "session-1",
			mockSetup: func(m *inputportMock.MockSessionsRepositoryInputPort) {
				m.EXPECT().GetSessionById(gomock.Any(), "session-1").Return(nil, response.NewInternalError(errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessions := inputportMock.NewMockSessionsRepositoryInputPort(ctrl)
			tt.mockSetup(mockSessions)

			e := echo.New()
			e.HTTPErrorHandler = response.ErrorHandler

			var gotUserID string
			var gotSessionCreatedAt time.Time
			nextCalled := false
			handler := func(c echo.Context) error {
				nextCalled = true
				gotUserID = UserIDFromContext(c.Request().Context())
				gotSessionCreatedAt = SessionCreatedAtFromContext(c.Request().Context())
				return c.NoContent(http.StatusOK)
			}
			e.GET("/protected", handler, Auth(mockSessions))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if !tt.omitCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: tt.cookieValue})
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.wantNext, nextCalled)
			if tt.wantNext {
				assert.Equal(t, "user-1", gotUserID)
				assert.True(t, sessionCreatedAt.Equal(gotSessionCreatedAt))
			}
		})
	}
}

func TestUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-1")
	assert.Equal(t, "user-1", UserIDFromContext(ctx))

	assert.Empty(t, UserIDFromContext(context.Background()))
}

func TestSessionCreatedAtFromContext(t *testing.T) {
	createdAt := time.Now().Add(-24 * time.Hour)
	ctx := context.WithValue(context.Background(), SessionCreatedAtKey, createdAt)
	assert.True(t, createdAt.Equal(SessionCreatedAtFromContext(ctx)))

	assert.True(t, SessionCreatedAtFromContext(context.Background()).IsZero())
}
