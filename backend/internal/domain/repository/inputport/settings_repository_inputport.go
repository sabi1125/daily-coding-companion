//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type SettingsRepositoryInputPort interface {
	CreateSetting(ctx context.Context, setting *entities.Settings) (err error)
	GetUserSetting(ctx context.Context, userId string) (setting entities.Settings, err error)
	UpdateUserSetting(ctx context.Context, userId string, preferences string) (err error)
}
