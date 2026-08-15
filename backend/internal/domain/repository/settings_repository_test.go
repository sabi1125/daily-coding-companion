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
