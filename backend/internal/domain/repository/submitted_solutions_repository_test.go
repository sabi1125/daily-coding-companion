package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"backend/internal/domain/entities"
	"backend/internal/response"

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

func TestSubmittedSolutionsRepository_SubmitSolution(t *testing.T) {
	input := entities.SubmittedSolutions{
		SolutionId: "s-1",
		ProblemId:  "problem-1",
		Solution:   "def two_sum(): pass",
		Status:     "Solved",
	}

	tests := []struct {
		name      string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "inserts successfully and returns the created row",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `submitted_solutions`")).
					WithArgs("s-1", "problem-1", "def two_sum(): pass", "Solved", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:    "returns a database error when insert fails",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `submitted_solutions`")).
					WithArgs("s-1", "problem-1", "def two_sum(): pass", "Solved", sqlmock.AnyArg()).
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

			repo := NewSubmittedSolutionsRepository(db)
			created, err := repo.SubmitSolution(context.Background(), input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
				return
			}

			assert.NoError(t, err)
			// created must actually carry the row back — SubmitSolution had
			// a bug where the named return was never assigned and always
			// came back zero-valued despite a successful insert.
			assert.Equal(t, input.SolutionId, created.SolutionId)
			assert.Equal(t, input.ProblemId, created.ProblemId)
			assert.Equal(t, input.Solution, created.Solution)
			assert.Equal(t, input.Status, created.Status)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSubmittedSolutionsRepository_GetDatesForHeatMap(t *testing.T) {
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	tests := []struct {
		name      string
		userId    string
		wantErr   bool
		want      []response.HeatMapDates
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name:   "returns submission dates with the same-day flag",
			userId: "user-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT date(s.submitted_at) as submitted_at, (date(s.submitted_at) = date(p.created_at)) as submitted_same_day_flag FROM submitted_solutions s LEFT JOIN problems p ON s.problem_id = p.problem_id WHERE p.user_id = ? AND s.submitted_at >= ? ORDER BY s.submitted_at ASC")).
					WithArgs("user-1", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"submitted_at", "submitted_same_day_flag"}).
						AddRow("2026-04-19", true).
						AddRow("2026-07-03", false))
			},
			want: []response.HeatMapDates{
				{SubmittedAt: "2026-04-19", SubmittedSameDayFlag: true},
				{SubmittedAt: "2026-07-03", SubmittedSameDayFlag: false},
			},
		},
		{
			name:   "no submissions in the last 6 months — empty result, not an error",
			userId: "user-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT date(s.submitted_at) as submitted_at, (date(s.submitted_at) = date(p.created_at)) as submitted_same_day_flag FROM submitted_solutions s LEFT JOIN problems p ON s.problem_id = p.problem_id WHERE p.user_id = ? AND s.submitted_at >= ? ORDER BY s.submitted_at ASC")).
					WithArgs("user-1", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"submitted_at", "submitted_same_day_flag"}))
			},
			want: []response.HeatMapDates{},
		},
		{
			name:    "returns a database error when the query fails",
			userId:  "user-1",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT date(s.submitted_at) as submitted_at, (date(s.submitted_at) = date(p.created_at)) as submitted_same_day_flag FROM submitted_solutions s LEFT JOIN problems p ON s.problem_id = p.problem_id WHERE p.user_id = ? AND s.submitted_at >= ? ORDER BY s.submitted_at ASC")).
					WithArgs("user-1", sqlmock.AnyArg()).
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
			dates, err := repo.GetDatesForHeatMap(context.Background(), tt.userId, sixMonthsAgo)

			if tt.wantErr {
				assert.Error(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, dates)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
