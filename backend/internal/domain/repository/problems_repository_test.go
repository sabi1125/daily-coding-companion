package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestProblemsRepository_GetProblems(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		userId     string
		status     string
		setupMock  func(mock sqlmock.Sqlmock)
		wantErr    bool
		wantIds    []string
		wantStatus map[string]string
	}{
		{
			name:   "derives Open/Failed/Solved and returns all when status is All",
			userId: "user-1",
			status: "All",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("user-1").
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "title", "needs_review_flag", "created_at"}).
						AddRow("p-open", "Open Problem", false, now).
						AddRow("p-failed", "Failed Problem", false, now).
						AddRow("p-solved", "Solved Problem", false, now))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("p-open", "p-failed", "p-solved").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}).
						AddRow("s1", "p-failed", "Failed").
						AddRow("s2", "p-solved", "Failed").
						AddRow("s3", "p-solved", "Solved"))
			},
			wantIds: []string{"p-open", "p-failed", "p-solved"},
			wantStatus: map[string]string{
				"p-open":   "Open",
				"p-failed": "Failed",
				"p-solved": "Solved",
			},
		},
		{
			name:   "empty status behaves the same as All",
			userId: "user-1",
			status: "",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("user-1").
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "title", "needs_review_flag", "created_at"}).
						AddRow("p-open", "Open Problem", false, now))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("p-open").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}))
			},
			wantIds: []string{"p-open"},
		},
		{
			name:   "filters to only the requested status",
			userId: "user-1",
			status: "Solved",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("user-1").
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "title", "needs_review_flag", "created_at"}).
						AddRow("p-open", "Open Problem", false, now).
						AddRow("p-solved", "Solved Problem", false, now))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("p-open", "p-solved").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}).
						AddRow("s1", "p-solved", "Solved"))
			},
			wantIds: []string{"p-solved"},
		},
		{
			name:   "no problems for user returns empty, not an error",
			userId: "user-1",
			status: "All",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("user-1").
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "title", "needs_review_flag", "created_at"}))
			},
			wantIds: []string{},
		},
		{
			name:   "returns error when the problems query fails",
			userId: "user-1",
			status: "All",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("user-1").
					WillReturnError(errors.New("db connection lost"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewProblemsRepository(db)
			problems, err := repo.GetProblems(context.Background(), tt.userId, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
				return
			}

			assert.NoError(t, err)
			gotIds := make([]string, len(problems))
			for i, p := range problems {
				gotIds[i] = p.ProblemId
			}
			assert.ElementsMatch(t, tt.wantIds, gotIds)

			for _, p := range problems {
				if want, ok := tt.wantStatus[p.ProblemId]; ok {
					assert.Equal(t, want, p.Status)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
