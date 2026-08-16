package entities

import "time"

type ProblemStatus string

const (
	ProblemOpen   = ProblemStatus("Open")
	ProblemFailed = ProblemStatus("Failed")
	ProblemSolved = ProblemStatus("Solved")
)

type GetProblemParams struct {
	Status string `query:"status" validate:"omitempty,oneof=Open Failed Solved"`
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
