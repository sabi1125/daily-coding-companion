//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
	"time"

	"backend/internal/domain/entities"
)

type OauthRepositoryInputPort interface {
	FindUserBySub(ctx context.Context, sub string) (oauthCredentials *entities.OauthCredentials, err error)
	CreateOauthCredentials(ctx context.Context, oauthCreds *entities.OauthCredentials) (err error)
	UpdateOauthInformationWithSub(ctx context.Context, sub string, refreshToken string, expiryAt time.Time) (err error)
}
