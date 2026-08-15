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

func TestSessionsRepository_CreateSession(t *testing.T) {
	tests := []struct {
		name      string
		input     *entities.Sessions
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "creates successfully",
			input: &entities.Sessions{
				SessionId: "session-1",
				UserId:    "user-1",
				ExpiresAt: time.Now().AddDate(0, 0, 30),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sessions`")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "returns error when insert fails",
			input: &entities.Sessions{
				SessionId: "session-1",
				UserId:    "user-1",
				ExpiresAt: time.Now().AddDate(0, 0, 30),
			},
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sessions`")).
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

			repo := NewSessionsRepository(db)
			err := repo.CreateSession(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
