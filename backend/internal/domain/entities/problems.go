package entities

import "time"

type (
	ProblemStatus     string
	ProblemDifficulty string
)

const (
	ProblemOpen   = ProblemStatus("Open")
	ProblemFailed = ProblemStatus("Failed")
	ProblemSolved = ProblemStatus("Solved")
)

type GetProblemParams struct {
	Status     string   `query:"status" validate:"omitempty,oneof=Open Failed Solved"`
	Difficulty []string `query:"difficulty" validate:"omitempty,dive,oneof=Easy Medium Hard"`
}

type GetProblemDetailParams struct {
	ProblemId string `param:"id" validate:"required,uuid"`
}

type Problems struct {
	ProblemId       string    `json:"problem_id" gorm:"column:problem_id"`
	UserId          string    `json:"-" gorm:"column:user_id"`
	RawProblem      string    `json:"raw_problem" gorm:"column:raw_problem"`
	Title           *string   `json:"title" gorm:"column:title"`
	ProblemText     *string   `json:"problem_text" gorm:"column:problem_text"`
	AlgorithmTag    *string   `json:"algorithm_tag" gorm:"column:algorithm_tag"`
	Difficulty      *string   `json:"difficulty" gorm:"column:difficulty"`
	AiHelp          *string   `json:"ai_help" gorm:"column:ai_help"`
	NeedsReviewFlag bool      `json:"needs_review_flag" gorm:"column:needs_review_flag"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	Status      string               `json:"status" gorm:"-"`
	Submissions []SubmittedSolutions `json:"-" gorm:"foreignKey:ProblemId;references:ProblemId"`
}

type AIHelp struct {
	Concept     string `json:"concept" jsonschema:"required,description=The underlying concept/technique this problem relies on, explained on its own, not just how it applies to this problem"`
	Nudge       string `json:"nudge" jsonschema:"required,description=A short question or hint that points toward the right idea without stating it outright"`
	Approach    string `json:"approach" jsonschema:"required,description=The strategy/algorithm to use, described in plain language"`
	Walkthrough string `json:"walkthrough" jsonschema:"required,description=A prose talk-through of how to apply the approach to this problem's inputs, step by step — not a full working solution or runnable code"`
}

type ProblemStateCount struct {
	Solved   int64 `json:"solved" gorm:"solved"`
	Unsolved int64 `json:"unsolved" gorm:"unsolved"`
}
