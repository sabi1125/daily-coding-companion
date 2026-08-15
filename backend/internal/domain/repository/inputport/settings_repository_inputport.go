//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type SettingsRepositoryInputPort interface {
	CreateSetting(ctx context.Context, setting *entities.Settings) (err error)
}
