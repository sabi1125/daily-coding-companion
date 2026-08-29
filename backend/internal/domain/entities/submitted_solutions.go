package entities

import (
	"errors"
	"time"
)

type GetSubmittedSolutionsParam struct {
	ProblemId string `param:"id" validate:"required,uuid"`
}

type SubmitSolutionsParam struct {
	ProblemId string `param:"id" validate:"required,uuid"`
}

type SubmittedSolutionsBody struct {
	Solution string `json:"solution" validate:"required,min=1"`
	Status   string `json:"status" validate:"required,min=1,oneof=Solved Failed"`
}

type SubmittedSolutions struct {
	SolutionId  string    `json:"solution_id" gorm:"column:solution_id"`
	ProblemId   string    `json:"-" gorm:"column:problem_id"`
	Solution    string    `json:"solution" gorm:"column:solution"`
	Status      string    `json:"status" gorm:"column:status"`
	SubmittedAt time.Time `json:"submitted_at" gorm:"column:submitted_at;autoCreateTime"`
}

type SubmittedSolutionForExecution struct {
	Language string `json:"language" validate:"required,oneof=python javascript cpp go"`
	Content  string `json:"content" validate:"required,min=1"`
}

type PistonFile struct {
	Content string `json:"content"`
}

type SubmittedSolutionRequest struct {
	Language string       `json:"language"`
	Version  string       `json:"version"`
	Files    []PistonFile `json:"files"`
}

func (s *SubmittedSolutionRequest) ResolveVersion() (err error) {
	switch s.Language {
	case "python":
		s.Version = "3.12.0"
	case "javascript":
		s.Version = "20.11.1"
	case "cpp":
		s.Version = "10.2.0"
	case "go":
		s.Version = "1.16.2"
	default:
		err = errors.New("unsupported language: " + s.Language)
	}
	return
}
