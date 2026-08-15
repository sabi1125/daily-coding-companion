package interactor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"backend/internal/domain/entities"
	inputportMock "backend/internal/domain/repository/inputport/mock"
	logger "backend/internal/log"
	"backend/internal/response"
	txMock "backend/internal/tx/mock"
	utilMock "backend/internal/util/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

func mockWithinTransaction(tx *txMock.MockManager) {
	tx.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		}).
		AnyTimes()
}

func TestMain(m *testing.M) {
	logger.InitForTest()
	m.Run()
}

func TestAuthInteractor_SignIn(t *testing.T) {
	interactor := NewAuthInteractor(
		oauth2.Config{ClientID: "test-client-id"},
		nil, nil, nil, nil, nil, nil, nil,
	)

	authUrl, csrf, err := interactor.SignIn(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, csrf)

	parsedUrl, err := url.Parse(authUrl)
	assert.NoError(t, err)
	assert.Equal(t, csrf, parsedUrl.Query().Get("state"))
}

func TestAuthInteractor_Callback_ExchangeErrors(t *testing.T) {
	tests := []struct {
		name           string
		buildServer    func() *httptest.Server
		wantStatusCode int
	}{
		{
			name: "invalid_grant maps to bad request",
			buildServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"invalid_grant","error_description":"Bad Request"}`))
				}))
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "other oauth error maps to bad gateway",
			buildServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"invalid_client","error_description":"bad client credentials"}`))
				}))
			},
			wantStatusCode: http.StatusBadGateway,
		},
		{
			name: "unreachable google maps to service unavailable",
			buildServer: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				server.Close() // closed before use: connection refused, no RetrieveError to parse
				return server
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.buildServer()
			defer server.Close()

			interactor := NewAuthInteractor(
				oauth2.Config{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					Endpoint: oauth2.Endpoint{
						TokenURL:  server.URL,
						AuthStyle: oauth2.AuthStyleInParams,
					},
				},
				nil, nil, nil, nil, nil, nil, nil,
			)

			_, _, err := interactor.Callback(context.Background(), "test-code")

			var appErr *response.AppError
			ok := errors.As(err, &appErr)
			assert.True(t, ok, "expected a *response.AppError, got %T: %v", err, err)
			assert.Equal(t, tt.wantStatusCode, appErr.Status.Code)
		})
	}
}

func TestAuthInteractor_Callback_PostExchangeErrors(t *testing.T) {
	tests := []struct {
		name           string
		tokenResponse  string
		wantStatusCode int
	}{
		{
			name:           "missing id_token maps to internal error",
			tokenResponse:  `{"access_token":"fake-access-token","token_type":"Bearer","expires_in":3600}`,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "malformed id_token maps to unauthorized",
			tokenResponse:  `{"access_token":"fake-access-token","token_type":"Bearer","expires_in":3600,"id_token":"not-a-valid-jwt"}`,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.tokenResponse))
			}))
			defer server.Close()

			interactor := NewAuthInteractor(
				oauth2.Config{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					Endpoint: oauth2.Endpoint{
						TokenURL:  server.URL,
						AuthStyle: oauth2.AuthStyleInParams,
					},
				},
				nil, nil, nil, nil, nil, nil, nil,
			)

			_, _, err := interactor.Callback(context.Background(), "test-code")

			var appErr *response.AppError
			ok := errors.As(err, &appErr)
			assert.True(t, ok, "expected a *response.AppError, got %T: %v", err, err)
			assert.Equal(t, tt.wantStatusCode, appErr.Status.Code)
		})
	}
}

func TestAuthInteractor_createUserRecords(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testSettingId := "test-setting-id"
	testExpires := time.Now().Add(time.Hour).Unix()

	testIdentity := entities.GoogleIdentity{
		Sub:        "test-sub",
		Email:      "test@example.com",
		GivenName:  "Test",
		FamilyName: "User",
	}
	testToken := &oauth2.Token{RefreshToken: "test-refresh-token"}

	errUUID := errors.New("uuid generation failed")
	errRepo := errors.New("repository call failed")

	tests := []struct {
		name        string
		prepareFunc func(
			mur *inputportMock.MockUsersRepositoryInputPort,
			motr *inputportMock.MockOauthRepositoryInputPort,
			msr *inputportMock.MockSettingsRepositoryInputPort,
			mUUID *utilMock.MockUUIDGenerator,
		)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(
				mur *inputportMock.MockUsersRepositoryInputPort,
				motr *inputportMock.MockOauthRepositoryInputPort,
				msr *inputportMock.MockSettingsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testUserId, nil)
				mUUID.EXPECT().NewV7().Return(testSettingId, nil)
				mur.EXPECT().CreateUser(gomock.Any(), &entities.Users{
					UserId:    testUserId,
					Email:     testIdentity.Email,
					FirstName: testIdentity.GivenName,
					LastName:  testIdentity.FamilyName,
				}).Return(nil)
				motr.EXPECT().CreateOauthCredentials(gomock.Any(), &entities.OauthCredentials{
					OauthId:      testIdentity.Sub,
					UserId:       testUserId,
					RefreshToken: testToken.RefreshToken,
					ExpiryAt:     time.Unix(testExpires, 0),
				}).Return(nil)
				msr.EXPECT().CreateSetting(gomock.Any(), &entities.Settings{
					SettingID: testSettingId,
					UserID:    testUserId,
				}).Return(nil)
			},
			wantedError: nil,
		},
		{
			name: "fails to generate user id",
			prepareFunc: func(
				mur *inputportMock.MockUsersRepositoryInputPort,
				motr *inputportMock.MockOauthRepositoryInputPort,
				msr *inputportMock.MockSettingsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return("", errUUID)
			},
			wantedError: response.NewInternalError(errUUID),
		},
		{
			name: "fails to generate setting id",
			prepareFunc: func(
				mur *inputportMock.MockUsersRepositoryInputPort,
				motr *inputportMock.MockOauthRepositoryInputPort,
				msr *inputportMock.MockSettingsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testUserId, nil)
				mUUID.EXPECT().NewV7().Return("", errUUID)
			},
			wantedError: response.NewInternalError(errUUID),
		},
		{
			name: "fails to create user",
			prepareFunc: func(
				mur *inputportMock.MockUsersRepositoryInputPort,
				motr *inputportMock.MockOauthRepositoryInputPort,
				msr *inputportMock.MockSettingsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testUserId, nil)
				mUUID.EXPECT().NewV7().Return(testSettingId, nil)
				mur.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(response.NewDatabaseError(errRepo))
			},
			wantedError: response.NewDatabaseError(errRepo),
		},
		{
			name: "fails to create oauth credentials",
			prepareFunc: func(
				mur *inputportMock.MockUsersRepositoryInputPort,
				motr *inputportMock.MockOauthRepositoryInputPort,
				msr *inputportMock.MockSettingsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testUserId, nil)
				mUUID.EXPECT().NewV7().Return(testSettingId, nil)
				mur.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
				motr.EXPECT().CreateOauthCredentials(gomock.Any(), gomock.Any()).Return(response.NewDatabaseError(errRepo))
			},
			wantedError: response.NewDatabaseError(errRepo),
		},
		{
			name: "fails to create setting",
			prepareFunc: func(
				mur *inputportMock.MockUsersRepositoryInputPort,
				motr *inputportMock.MockOauthRepositoryInputPort,
				msr *inputportMock.MockSettingsRepositoryInputPort,
				mUUID *utilMock.MockUUIDGenerator,
			) {
				mUUID.EXPECT().NewV7().Return(testUserId, nil)
				mUUID.EXPECT().NewV7().Return(testSettingId, nil)
				mur.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
				motr.EXPECT().CreateOauthCredentials(gomock.Any(), gomock.Any()).Return(nil)
				msr.EXPECT().CreateSetting(gomock.Any(), gomock.Any()).Return(response.NewDatabaseError(errRepo))
			},
			wantedError: response.NewDatabaseError(errRepo),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsers := inputportMock.NewMockUsersRepositoryInputPort(ctrl)
			mockOauth := inputportMock.NewMockOauthRepositoryInputPort(ctrl)
			mockSettings := inputportMock.NewMockSettingsRepositoryInputPort(ctrl)
			mockUUID := utilMock.NewMockUUIDGenerator(ctrl)
			mockTx := txMock.NewMockManager(ctrl)
			mockWithinTransaction(mockTx)

			tt.prepareFunc(mockUsers, mockOauth, mockSettings, mockUUID)

			interactor := NewAuthInteractor(
				oauth2.Config{}, nil, mockUsers, mockOauth, mockSettings, nil, mockUUID, mockTx,
			)

			userId, err := interactor.createUserRecords(ctx, testIdentity, testToken, testExpires)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				assert.Empty(t, userId)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testUserId, userId)
		})
	}
}

func TestAuthInteractor_createSession(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testSessionId := "test-session-id"

	errUUID := errors.New("uuid generation failed")
	errRepo := errors.New("repository call failed")

	tests := []struct {
		name        string
		prepareFunc func(msr *inputportMock.MockSessionsRepositoryInputPort, mUUID *utilMock.MockUUIDGenerator)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(msr *inputportMock.MockSessionsRepositoryInputPort, mUUID *utilMock.MockUUIDGenerator) {
				mUUID.EXPECT().NewV7().Return(testSessionId, nil)
				msr.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantedError: nil,
		},
		{
			name: "fails to generate session id",
			prepareFunc: func(msr *inputportMock.MockSessionsRepositoryInputPort, mUUID *utilMock.MockUUIDGenerator) {
				mUUID.EXPECT().NewV7().Return("", errUUID)
			},
			wantedError: response.NewInternalError(errUUID),
		},
		{
			name: "fails to create session",
			prepareFunc: func(msr *inputportMock.MockSessionsRepositoryInputPort, mUUID *utilMock.MockUUIDGenerator) {
				mUUID.EXPECT().NewV7().Return(testSessionId, nil)
				msr.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(response.NewDatabaseError(errRepo))
			},
			wantedError: response.NewDatabaseError(errRepo),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessions := inputportMock.NewMockSessionsRepositoryInputPort(ctrl)
			mockUUID := utilMock.NewMockUUIDGenerator(ctrl)
			tt.prepareFunc(mockSessions, mockUUID)

			interactor := NewAuthInteractor(
				oauth2.Config{}, nil, nil, nil, nil, mockSessions, mockUUID, nil,
			)

			before := time.Now()
			sessionId, expiresAt, err := interactor.createSession(ctx, testUserId)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				assert.Empty(t, sessionId)
				assert.True(t, expiresAt.IsZero())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testSessionId, sessionId)
			assert.WithinDuration(t, before.AddDate(0, 0, 30), expiresAt, time.Minute)
		})
	}
}

func TestAuthInteractor_DeleteUserSession(t *testing.T) {
	ctx := context.Background()
	testUserId := "test-user-id"
	testSessionId := "test-session-id"

	errRepo := response.NewDatabaseError(errors.New("repository call failed"))

	tests := []struct {
		name        string
		prepareFunc func(msr *inputportMock.MockSessionsRepositoryInputPort)
		wantedError error
	}{
		{
			name: "success",
			prepareFunc: func(msr *inputportMock.MockSessionsRepositoryInputPort) {
				msr.EXPECT().DeleteUserSession(gomock.Any(), testSessionId, testUserId).Return(nil)
			},
		},
		{
			name: "fails to delete session",
			prepareFunc: func(msr *inputportMock.MockSessionsRepositoryInputPort) {
				msr.EXPECT().DeleteUserSession(gomock.Any(), testSessionId, testUserId).Return(errRepo)
			},
			wantedError: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessions := inputportMock.NewMockSessionsRepositoryInputPort(ctrl)
			mockTx := txMock.NewMockManager(ctrl)
			mockWithinTransaction(mockTx)

			tt.prepareFunc(mockSessions)

			interactor := NewAuthInteractor(
				oauth2.Config{}, nil, nil, nil, nil, mockSessions, nil, mockTx,
			)

			err := interactor.DeleteUserSession(ctx, testUserId, testSessionId)

			if tt.wantedError != nil {
				assert.EqualError(t, err, tt.wantedError.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestExtractGoogleIdentity(t *testing.T) {
	tests := []struct {
		name         string
		payload      *idtoken.Payload
		wantIdentity entities.GoogleIdentity
		wantErr      bool
	}{
		{
			name: "success",
			payload: &idtoken.Payload{
				Subject: "test-sub",
				Claims: map[string]interface{}{
					"email":       "test@example.com",
					"given_name":  "Test",
					"family_name": "User",
				},
			},
			wantIdentity: entities.GoogleIdentity{
				Sub:        "test-sub",
				Email:      "test@example.com",
				GivenName:  "Test",
				FamilyName: "User",
			},
			wantErr: false,
		},
		{
			name: "fails to marshal claims",
			payload: &idtoken.Payload{
				Subject: "test-sub",
				Claims: map[string]interface{}{
					"email": make(chan int),
				},
			},
			wantErr: true,
		},
		{
			name: "fails to unmarshal claims into identity",
			payload: &idtoken.Payload{
				Subject: "test-sub",
				Claims: map[string]interface{}{
					"email": 12345,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := extractGoogleIdentity(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantIdentity, identity)
		})
	}
}

func TestGenerateState(t *testing.T) {
	first, err := generateState()
	assert.NoError(t, err)
	assert.NotEmpty(t, first)

	second, err := generateState()
	assert.NoError(t, err)
	assert.NotEmpty(t, second)

	assert.NotEqual(t, first, second)
}
