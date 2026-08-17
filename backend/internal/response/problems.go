package response

import "time"

// ProblemSummary is the History-list shape — only what the list view
// renders. Full problem content (raw_problem, problem_text, algorithm_tag,
// difficulty, ai_help) is what GET /problems/{id} is for.
type ProblemSummary struct {
	ProblemId       string    `json:"problem_id"`
	Title           *string   `json:"title"`
	Status          string    `json:"status"`
	NeedsReviewFlag bool      `json:"needs_review_flag"`
	CreatedAt       time.Time `json:"created_at"`
}

type GetProblemsResponse struct {
	Result []ProblemSummary `json:"result"`
	Total  int              `json:"total"`
}

// ProblemDetail is the GET /problems/{id} shape — full problem content, no
// status (submissions, which status is derived from, come from a separate
// GET /submissions/{id} call, not this one).
type ProblemDetail struct {
	ProblemId       string    `json:"problem_id"`
	RawProblem      string    `json:"raw_problem"`
	Title           *string   `json:"title"`
	ProblemText     *string   `json:"problem_text"`
	AlgorithmTag    *string   `json:"algorithm_tag"`
	Difficulty      *string   `json:"difficulty"`
	AiHelp          *string   `json:"ai_help"`
	NeedsReviewFlag bool      `json:"needs_review_flag"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TodaysProblem is the GET /problems/today shape — full problem content
// plus status, since /today can return a problem the caller already has
// submissions against from earlier in the day.
type TodaysProblem struct {
	ProblemId       string    `json:"problem_id"`
	RawProblem      string    `json:"raw_problem"`
	Title           *string   `json:"title"`
	ProblemText     *string   `json:"problem_text"`
	AlgorithmTag    *string   `json:"algorithm_tag"`
	Difficulty      *string   `json:"difficulty"`
	Status          string    `json:"status"`
	AiHelp          *string   `json:"ai_help"`
	NeedsReviewFlag bool      `json:"needs_review_flag"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TodaysProblemResponse struct {
	Result TodaysProblem `json:"result"`
}
