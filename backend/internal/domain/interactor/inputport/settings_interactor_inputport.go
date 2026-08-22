//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
	"time"

	"backend/internal/domain/entities"
)

type SettingsInteractorInputPort interface {
	GetUserSetting(ctx context.Context, userId string, sessionCreatedAt time.Time) (setting entities.Settings, problemStateCount entities.ProblemStateCount, email string, err error)
	UpdateUserSetting(ctx context.Context, userId string, preference string, sessionCreatedAt time.Time) (setting entities.Settings, err error)
}
