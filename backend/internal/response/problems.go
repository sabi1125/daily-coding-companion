package response

import "time"

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

type AIHelp struct {
	Concept     string `json:"concept"`
	Nudge       string `json:"nudge"`
	Approach    string `json:"approach"`
	Walkthrough string `json:"walkthrough"`
}
