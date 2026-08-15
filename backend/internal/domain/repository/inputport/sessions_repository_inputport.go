//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type SessionsRepositoryInputPort interface {
	CreateSession(ctx context.Context, session *entities.Sessions) (err error)
	GetSessionById(ctx context.Context, sessionId string) (session *entities.Sessions, err error)
}
