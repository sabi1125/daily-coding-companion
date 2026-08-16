package response

import "time"

type Submission struct {
	SolutionId  string    `json:"solution_id"`
	ProblemId   string    `json:"problem_id"`
	Solution    string    `json:"solution"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type GetSubmissionsResponse struct {
	Result []Submission `json:"result"`
	Total  int          `json:"total"`
}
