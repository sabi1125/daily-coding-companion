//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type SettingsInteractorInputPort interface {
	GetUserSetting(ctx context.Context, userId string) (setting entities.Settings, err error)
	UpdateUserSetting(ctx context.Context, userId string, preference string) (setting entities.Settings, err error)
}
