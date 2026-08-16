package entities

import "time"

type GetSubmittedSolutionsParam struct {
	ProblemId string `param:"id" validate:"required,uuid"`
}

type SubmittedSolutions struct {
	SolutionId  string    `json:"solution_id" gorm:"column:solution_id"`
	ProblemId   string    `json:"-" gorm:"column:problem_id"`
	Solution    string    `json:"solution" gorm:"column:solution"`
	Status      string    `json:"status" gorm:"column:status"`
	SubmittedAt time.Time `json:"submitted_at" gorm:"column:submitted_at"`
}
