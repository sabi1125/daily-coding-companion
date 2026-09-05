package repository

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"backend/internal/domain/entities"
	"backend/internal/response"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestProblemsRepository_GetProblems(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		userId     string
		status     string
		difficulty []entities.ProblemDifficulty
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
			name:       "filters by a single difficulty at the SQL level",
			userId:     "user-1",
			status:     "All",
			difficulty: []entities.ProblemDifficulty{"Easy"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("Easy", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "title", "needs_review_flag", "created_at"}).
						AddRow("p-easy", "Easy Problem", false, now))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("p-easy").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}))
			},
			wantIds: []string{"p-easy"},
		},
		{
			name:       "filters by multiple difficulties at the SQL level",
			userId:     "user-1",
			status:     "All",
			difficulty: []entities.ProblemDifficulty{"Easy", "Medium"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .problem_id.,.title.,.needs_review_flag.,.created_at. FROM .problems.").
					WithArgs("Easy", "Medium", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "title", "needs_review_flag", "created_at"}).
						AddRow("p-easy", "Easy Problem", false, now).
						AddRow("p-medium", "Medium Problem", false, now))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("p-easy", "p-medium").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}))
			},
			wantIds: []string{"p-easy", "p-medium"},
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
			problems, err := repo.GetProblems(context.Background(), tt.userId, tt.status, tt.difficulty)

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

func TestProblemsRepository_GetProblemDetails(t *testing.T) {
	tests := []struct {
		name       string
		userId     string
		problemId  string
		wantErr    bool
		wantCode   int
		wantStatus string
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:      "returns the matching problem",
			userId:    "user-1",
			problemId: "problem-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM .problems.").
					WithArgs("user-1", "problem-1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "user_id", "title"}).
						AddRow("problem-1", "user-1", "Two Sum"))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("problem-1").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}).
						AddRow("s1", "problem-1", "Solved"))
			},
			wantStatus: "Solved",
		},
		{
			name:      "returns 404 when no matching problem",
			userId:    "user-1",
			problemId: "unknown-problem",
			wantErr:   true,
			wantCode:  http.StatusNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM .problems.").
					WithArgs("user-1", "unknown-problem", 1).
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "user_id", "title"}))
			},
		},
		{
			name:      "returns error on unexpected db failure",
			userId:    "user-1",
			problemId: "problem-1",
			wantErr:   true,
			wantCode:  http.StatusInternalServerError,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM .problems.").
					WithArgs("user-1", "problem-1", 1).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewProblemsRepository(db)
			problem, err := repo.GetProblemDetails(context.Background(), tt.userId, tt.problemId)

			if tt.wantErr {
				var appErr *response.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, tt.wantCode, appErr.Status.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.problemId, problem.ProblemId)
				assert.Equal(t, tt.wantStatus, problem.Status)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProblemsRepository_CreateProblem(t *testing.T) {
	input := &entities.Problems{
		ProblemId:  "p-1",
		UserId:     "user-1",
		RawProblem: "raw problem text",
	}

	tests := []struct {
		name      string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "inserts successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `problems`")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:    "returns a database error when insert fails",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `problems`")).
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

			repo := NewProblemsRepository(db)
			err := repo.CreateProblem(context.Background(), input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProblemsRepository_GetTodaysproblem(t *testing.T) {
	todaysDate := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		wantErr    bool
		wantId     string
		wantStatus string
		setupMock  func(mock sqlmock.Sqlmock)
	}{
		{
			name:       "returns today's problem with derived status from its submissions",
			wantId:     "p-1",
			wantStatus: "Solved",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .*problems.* FROM .problems. JOIN ingest_runs").
					WithArgs("user-1", todaysDate, 1).
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "user_id"}).
						AddRow("p-1", "user-1"))

				mock.ExpectQuery("SELECT \\* FROM .submitted_solutions.").
					WithArgs("p-1").
					WillReturnRows(sqlmock.NewRows([]string{"solution_id", "problem_id", "status"}).
						AddRow("s-1", "p-1", "Solved"))
			},
		},
		{
			name: "no problem for today returns the zero value, not an error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .*problems.* FROM .problems. JOIN ingest_runs").
					WithArgs("user-1", todaysDate, 1).
					WillReturnRows(sqlmock.NewRows([]string{"problem_id", "user_id"}))
			},
		},
		{
			name:    "returns error on unexpected db failure",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .*problems.* FROM .problems. JOIN ingest_runs").
					WithArgs("user-1", todaysDate, 1).
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewProblemsRepository(db)
			problem, err := repo.GetTodaysproblem(context.Background(), "user-1", todaysDate)

			if tt.wantErr {
				assert.Error(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantId, problem.ProblemId)
			if tt.wantId != "" {
				assert.Equal(t, tt.wantStatus, problem.Status)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProblemsRepository_UpdateProblemWithAIHelp(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		setupMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "updates ai_help successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE `problems` SET")).
					WithArgs(`{"concept":"c"}`, sqlmock.AnyArg(), "problem-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:    "returns a database error when update fails",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE `problems` SET")).
					WithArgs(`{"concept":"c"}`, sqlmock.AnyArg(), "problem-1", "user-1").
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

			repo := NewProblemsRepository(db)
			err := repo.UpdateProblemWithAIHelp(context.Background(), "user-1", "problem-1", `{"concept":"c"}`)

			if tt.wantErr {
				var appErr *response.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, http.StatusInternalServerError, appErr.Status.Code)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProblemsRepository_GetProblemsCount(t *testing.T) {
	tests := []struct {
		name         string
		wantErr      bool
		setupMock    func(mock sqlmock.Sqlmock)
		wantSolved   int64
		wantUnsolved int64
	}{
		{
			name: "returns solved and unsolved counts",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.|\n)*FROM problems AS p LEFT JOIN submitted_solutions AS s ON s.problem_id = p.problem_id WHERE p.user_id = .").
					WithArgs("Solved", "Solved", "user-1").
					WillReturnRows(sqlmock.NewRows([]string{"solved", "unsolved"}).
						AddRow(3, 2))
			},
			wantSolved:   3,
			wantUnsolved: 2,
		},
		{
			name:    "returns a database error when the query fails",
			wantErr: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.|\n)*FROM problems AS p LEFT JOIN submitted_solutions AS s ON s.problem_id = p.problem_id WHERE p.user_id = .").
					WithArgs("Solved", "Solved", "user-1").
					WillReturnError(errors.New("db connection lost"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.setupMock(mock)

			repo := NewProblemsRepository(db)
			problemStateCount, err := repo.GetProblemsCount(context.Background(), "user-1")

			if tt.wantErr {
				var appErr *response.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, http.StatusInternalServerError, appErr.Status.Code)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSolved, problemStateCount.Solved)
			assert.Equal(t, tt.wantUnsolved, problemStateCount.Unsolved)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
