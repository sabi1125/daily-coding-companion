package entities

import "time"

type Sessions struct {
	SessionId string    `json:"session_id" gorm:"column:session_id"`
	UserId    string    `json:"user_id" gorm:"column:user_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at"`
}
