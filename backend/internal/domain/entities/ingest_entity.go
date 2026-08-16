package entities

import (
	"time"
)

type IngestRuns struct {
	IngestRunId string    `json:"ingest_run_id" gorm:"column:ingest_run_id"`
	UserId      string    `json:"user_id" gorm:"column:user_id"`
	ProblemId   *string   `json:"problem_id" gorm:"column:problem_id"`
	Status      string    `json:"status" gorm:"column:status"`
	Error       *string   `json:"error" gorm:"column:error"`
	Retried     bool      `json:"retried" gorm:"column:retried"`
	IngestDate  time.Time `json:"ingest_date" gorm:"column:ingest_date"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}
