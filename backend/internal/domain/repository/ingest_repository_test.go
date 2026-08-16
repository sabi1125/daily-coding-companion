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

func TestIngestRepository_GetIngestByUserId(t *testing.T) {
	ingestDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		userId    string
		retried   bool
		wantErr   bool
		wantCount int
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:    "returns a row when ingest already ran today",
			userId:  "user-1",
			retried: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `ingest_runs` WHERE user_id = ? AND ingest_date = ? AND retried = ?")).
					WithArgs("user-1", ingestDate, false).
					WillReturnRows(sqlmock.NewRows([]string{"ingest_run_id", "user_id", "status", "retried", "ingest_date"}).
						AddRow("run-1", "user-1", "success", false, ingestDate))
			},
			wantCount: 1,
		},
		{
			name:    "empty when nothing ran for today — an older row doesn't count",
			userId:  "user-1",
			retried: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `ingest_runs` WHERE user_id = ? AND ingest_date = ? AND retried = ?")).
					WithArgs("user-1", ingestDate, false).
					WillReturnRows(sqlmock.NewRows([]string{"ingest_run_id", "user_id", "status", "retried", "ingest_date"}))
			},
			wantCount: 0,
		},
		{
			name:    "returns error on unexpected db failure",
			userId:  "user-1",
			retried: false,
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `ingest_runs` WHERE user_id = ? AND ingest_date = ? AND retried = ?")).
					WithArgs("user-1", ingestDate, false).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewIngestRepository(db)
			got, err := repo.GetIngestByUserId(context.Background(), tt.userId, ingestDate, tt.retried)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantCount)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestIngestRepository_CreateIngestWithErr(t *testing.T) {
	errMsg := "refresh token invalid"

	tests := []struct {
		name      string
		input     entities.IngestRuns
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "creates successfully",
			input: entities.IngestRuns{
				IngestRunId: "run-1",
				UserId:      "user-1",
				Status:      "failed",
				Error:       &errMsg,
				Retried:     false,
				IngestDate:  time.Now(),
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `ingest_runs`")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "returns error when insert fails",
			input: entities.IngestRuns{
				IngestRunId: "run-1",
				UserId:      "user-1",
				Status:      "failed",
				Error:       &errMsg,
				Retried:     false,
				IngestDate:  time.Now(),
			},
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `ingest_runs`")).
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

			repo := NewIngestRepository(db)
			err := repo.CreateIngestWithErr(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
