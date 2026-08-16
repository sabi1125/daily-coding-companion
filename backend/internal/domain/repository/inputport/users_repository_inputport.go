//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"

	"backend/internal/domain/entities"
)

type UsersRepositoryInputPort interface {
	CreateUser(ctx context.Context, user *entities.Users) (err error)
	GetAllUserIds(ctx context.Context) (userIds []string, err error)
}
