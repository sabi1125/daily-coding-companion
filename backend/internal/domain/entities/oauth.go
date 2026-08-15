package entities

import "time"

type OauthCredentials struct {
	OauthId      string    `json:"oauth_id" gorm:"column:oauth_id"`
	UserId       string    `json:"user_id" gorm:"column:user_id"`
	RefreshToken string    `json:"refresh_token" gorm:"column:refresh_token"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	ExpiryAt     time.Time `json:"expiry_at" gorm:"column:expiry_at"`
}
