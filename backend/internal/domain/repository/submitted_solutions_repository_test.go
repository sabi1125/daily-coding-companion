package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSubmittedSolutionsRepository_GetSubmittedSolutions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		userId    string
		problemId string
		wantErr   bool
		wantIds   []string
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:      "returns the submissions for a problem owned by the caller",
			userId:    "user-1",
			problemId: "problem-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT submitted_solutions\\.\\* FROM .submitted_solutions. JOIN problems ON problems\\.problem_id = submitted_solutions\\.problem_id").
					WithArgs("problem-1", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "solution", "status", "submitted_at"}).
						AddRow("s1", "problem-1", "def two_sum(): pass", "Solved", now).
						AddRow("s2", "problem-1", "def two_sum(): pass", "Failed", now))
			},
			wantIds: []string{"s1", "s2"},
		},
		{
			name:      "problem doesn't exist — empty result, not an error",
			userId:    "user-1",
			problemId: "unknown-problem",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT submitted_solutions\\.\\* FROM .submitted_solutions. JOIN problems ON problems\\.problem_id = submitted_solutions\\.problem_id").
					WithArgs("unknown-problem", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "solution", "status", "submitted_at"}))
			},
			wantIds: []string{},
		},
		{
			name:      "problem belongs to another user — empty result, not a leak",
			userId:    "user-1",
			problemId: "someone-elses-problem",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT submitted_solutions\\.\\* FROM .submitted_solutions. JOIN problems ON problems\\.problem_id = submitted_solutions\\.problem_id").
					WithArgs("someone-elses-problem", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "solution", "status", "submitted_at"}))
			},
			wantIds: []string{},
		},
		{
			name:      "no submissions yet for an owned problem",
			userId:    "user-1",
			problemId: "problem-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT submitted_solutions\\.\\* FROM .submitted_solutions. JOIN problems ON problems\\.problem_id = submitted_solutions\\.problem_id").
					WithArgs("problem-1", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "solution", "status", "submitted_at"}))
			},
			wantIds: []string{},
		},
		{
			name:      "returns a database error when the query fails",
			userId:    "user-1",
			problemId: "problem-1",
			wantErr:   true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT submitted_solutions\\.\\* FROM .submitted_solutions. JOIN problems ON problems\\.problem_id = submitted_solutions\\.problem_id").
					WithArgs("problem-1", "user-1").
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewSubmittedSolutionsRepository(db)
			submissions, err := repo.GetSubmittedSolutions(context.Background(), tt.userId, tt.problemId)

			if tt.wantErr {
				assert.Error(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
				return
			}

			assert.NoError(t, err)
			gotIds := make([]string, len(submissions))
			for i, s := range submissions {
				gotIds[i] = s.SolutionId
			}
			assert.ElementsMatch(t, tt.wantIds, gotIds)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
