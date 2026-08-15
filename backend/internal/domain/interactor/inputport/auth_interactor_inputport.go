//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
	"time"
)

type AuthInteractorInputPort interface {
	SignIn(ctx context.Context) (authUrl string, csrf string, err error)
	Callback(ctx context.Context, code string) (sessionId string, expiresAt time.Time, err error)
}
