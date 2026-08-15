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

func TestSessionsRepository_GetSessionById(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:      "returns existing session",
			sessionId: "session-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `sessions` WHERE session_id = ?")).
					WithArgs("session-1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"session_id", "user_id"}).
						AddRow("session-1", "user-1"))
			},
		},
		{
			name:      "returns nil, nil when no session found",
			sessionId: "unknown-session",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `sessions` WHERE session_id = ?")).
					WithArgs("unknown-session", 1).
					WillReturnRows(sqlmock.NewRows([]string{"session_id", "user_id"}))
			},
		},
		{
			name:      "returns error on unexpected db failure",
			sessionId: "session-1",
			wantErr:   true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `sessions` WHERE session_id = ?")).
					WithArgs("session-1", 1).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewSessionsRepository(db)
			got, err := repo.GetSessionById(context.Background(), tt.sessionId)

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

func TestSessionsRepository_DeleteUserSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionId string
		userId    string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:      "deletes successfully",
			sessionId: "session-1",
			userId:    "user-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `sessions`")).
					WithArgs("session-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:      "no matching row still succeeds — treated as already signed out",
			sessionId: "session-1",
			userId:    "user-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `sessions`")).
					WithArgs("session-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
		},
		{
			name:      "returns error when delete fails",
			sessionId: "session-1",
			userId:    "user-1",
			wantErr:   true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `sessions`")).
					WithArgs("session-1", "user-1").
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
			err := repo.DeleteUserSession(context.Background(), tt.sessionId, tt.userId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
