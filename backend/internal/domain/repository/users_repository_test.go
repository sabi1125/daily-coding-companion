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

func TestUsersRepository_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		input     *entities.Users
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "creates successfully",
			input: &entities.Users{
				UserId:    "user-1",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "returns error when insert fails",
			input: &entities.Users{
				UserId:    "user-1",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
			},
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
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

			repo := NewUsersRepository(db)
			err := repo.CreateUser(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUsersRepository_GetAllUserIds(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		wantIds   []string
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "returns every signed-up user's id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `user_id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).
						AddRow("user-1").
						AddRow("user-2"))
			},
			wantIds: []string{"user-1", "user-2"},
		},
		{
			name: "no users yet — empty, not an error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `user_id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
			},
			wantIds: []string{},
		},
		{
			name:    "returns error on unexpected db failure — must not silently return empty",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT `user_id` FROM `users`")).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewUsersRepository(db)
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
