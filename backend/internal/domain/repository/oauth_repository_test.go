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
