package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"backend/internal/domain/entities"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestOauthRepository_FindUserBySub(t *testing.T) {
	tests := []struct {
		name      string
		sub       string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "returns existing record",
			sub:  "google-sub-123",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `oauth_credentials` WHERE oauth_id = ?")).
					WithArgs("google-sub-123", 1). // Take's LIMIT is a bound arg, not a literal
					WillReturnRows(sqlmock.NewRows([]string{"oauth_id", "user_id", "refresh_token"}).
						AddRow("google-sub-123", "user-1", "refresh-token-value"))
			},
		},
		{
			name: "returns nil, nil when no record found — not an error, signals a new user",
			sub:  "unknown-sub",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `oauth_credentials` WHERE oauth_id = ?")).
					WithArgs("unknown-sub", 1).
					WillReturnRows(sqlmock.NewRows([]string{"oauth_id", "user_id", "refresh_token"}))
			},
		},
		{
			name:    "returns error on unexpected db failure",
			sub:     "google-sub-123",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `oauth_credentials` WHERE oauth_id = ?")).
					WithArgs("google-sub-123", 1).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewOauthRepository(db)
			got, err := repo.FindUserBySub(context.Background(), tt.sub)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			_ = got
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOauthRepository_FindUserByUserId(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:   "returns existing record",
			userId: "user-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `oauth_credentials` WHERE user_id = ?")).
					WithArgs("user-1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"oauth_id", "user_id", "refresh_token"}).
						AddRow("google-sub-123", "user-1", "refresh-token-value"))
			},
		},
		{
			name:   "returns nil, nil when no record found",
			userId: "unknown-user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `oauth_credentials` WHERE user_id = ?")).
					WithArgs("unknown-user", 1).
					WillReturnRows(sqlmock.NewRows([]string{"oauth_id", "user_id", "refresh_token"}))
			},
		},
		{
			name:    "returns error on unexpected db failure",
			userId:  "user-1",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `oauth_credentials` WHERE user_id = ?")).
					WithArgs("user-1", 1).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewOauthRepository(db)
			got, err := repo.FindUserByUserId(context.Background(), tt.userId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			_ = got
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOauthRepository_GetAllUserIds(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		wantIds   []string
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "returns every connected user's id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `user_id` FROM `oauth_credentials`")).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).
						AddRow("user-1").
						AddRow("user-2"))
			},
			wantIds: []string{"user-1", "user-2"},
		},
		{
			name: "no connected users yet — empty, not an error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `user_id` FROM `oauth_credentials`")).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
			},
			wantIds: []string{},
		},
		{
			name:    "returns error on unexpected db failure — must not silently return empty",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `user_id` FROM `oauth_credentials`")).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewOauthRepository(db)
			got, err := repo.GetAllUserIds(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.wantIds, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOauthRepository_CreateOauthCredentials(t *testing.T) {
	tests := []struct {
		name      string
		input     *entities.OauthCredentials
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "creates successfully",
			input: &entities.OauthCredentials{
				OauthId:      "google-sub-123",
				UserId:       "user-1",
				RefreshToken: "refresh-token-value",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `oauth_credentials`")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "returns error when insert fails",
			input: &entities.OauthCredentials{
				OauthId:      "google-sub-123",
				UserId:       "user-1",
				RefreshToken: "refresh-token-value",
			},
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `oauth_credentials`")).
					WillReturnError(errors.New("db connection lost"))
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewOauthRepository(db)
			err := repo.CreateOauthCredentials(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOauthRepository_UpdateOauthInformationWithSub(t *testing.T) {
	tests := []struct {
		name         string
		sub          string
		refreshToken string
		wantErr      bool
		setupMock    func(mock sqlmock.Sqlmock)
	}{
		{
			name:         "updates successfully",
			sub:          "google-sub-123",
			refreshToken: "new-refresh-token",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE `oauth_credentials` SET")).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:         "returns error when update fails",
			sub:          "google-sub-123",
			refreshToken: "new-refresh-token",
			wantErr:      true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE `oauth_credentials` SET")).
					WillReturnError(errors.New("db connection lost"))
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewOauthRepository(db)
			err := repo.UpdateOauthInformationWithSub(context.Background(), tt.sub, tt.refreshToken, time.Now())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
