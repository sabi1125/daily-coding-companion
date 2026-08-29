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

type ExecuteSubmissionResponse struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Compile  *Run   `json:"compile,omitempty"`
	Run      Run    `json:"run"`
}

type Run struct {
	Stdout   string  `json:"stdout"`
	Stderr   string  `json:"stderr"`
	Output   string  `json:"output"`
	Code     int     `json:"code"`
	Signal   *string `json:"signal"`
	CPUTime  int     `json:"cpu_time"`
	WallTime int     `json:"wall_time"`
	Memory   int     `json:"memory"`
	Message  *string `json:"message"`
	Status   *string `json:"status"`
}
