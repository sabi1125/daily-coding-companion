//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
)

type AuthInteractorInputPort interface {
	SignIn(ctx context.Context) (authUrl string, csrf string, err error)
}
