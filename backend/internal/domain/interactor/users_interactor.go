package interactor

import (
	"backend/internal/domain/repository/inputport"
)

type UsersInteractor struct {
	usersRepository inputport.UsersRepositoryInputPort
}

func NewUsersInteractor(usersRepository inputport.UsersRepositoryInputPort) *UsersInteractor {
	return &UsersInteractor{
		usersRepository: usersRepository,
	}
}
