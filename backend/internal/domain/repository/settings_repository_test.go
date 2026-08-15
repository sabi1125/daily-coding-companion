package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"backend/internal/domain/entities"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSettingsRepository_CreateSetting(t *testing.T) {
	tests := []struct {
		name      string
		input     *entities.Settings
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "creates successfully",
			input: &entities.Settings{
				SettingID: "setting-1",
				UserID:    "user-1",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `settings`")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "returns error when insert fails",
			input: &entities.Settings{
				SettingID: "setting-1",
				UserID:    "user-1",
			},
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `settings`")).
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

			repo := NewSettingsRepository(db)
			err := repo.CreateSetting(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSettingsRepository_GetUserSetting(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:   "returns existing setting",
			userId: "user-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `settings` WHERE user_id = ?")).
					WithArgs("user-1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"setting_id", "user_id"}).
						AddRow("setting-1", "user-1"))
			},
		},
		{
			name:    "returns error when no setting found",
			userId:  "unknown-user",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `settings` WHERE user_id = ?")).
					WithArgs("unknown-user", 1).
					WillReturnRows(sqlmock.NewRows([]string{"setting_id", "user_id"}))
			},
		},
		{
			name:    "returns error on unexpected db failure",
			userId:  "user-1",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `settings` WHERE user_id = ?")).
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

			repo := NewSettingsRepository(db)
			got, err := repo.GetUserSetting(context.Background(), tt.userId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "setting-1", got.SettingID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSettingsRepository_UpdateUserSetting(t *testing.T) {
	tests := []struct {
		name        string
		userId      string
		preferences string
		wantErr     bool
		setupMock   func(mock sqlmock.Sqlmock)
	}{
		{
			name:        "updates successfully",
			userId:      "user-1",
			preferences: "new preferences",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE `settings` SET")).
					WithArgs("new preferences", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:        "returns error when update fails",
			userId:      "user-1",
			preferences: "new preferences",
			wantErr:     true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE `settings` SET")).
					WithArgs("new preferences", "user-1").
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

			repo := NewSettingsRepository(db)
			err := repo.UpdateUserSetting(context.Background(), tt.userId, tt.preferences)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
