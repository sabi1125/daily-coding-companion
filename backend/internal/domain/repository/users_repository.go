package repository

import (
	"context"

	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"
	"backend/internal/tx"

	"gorm.io/gorm"
)

type UsersRepository struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}

func (repository *UsersRepository) CreateUser(ctx context.Context, user *entities.Users) (err error) {
	logger.Infof("UserRepository: CreateUser")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Create(&user).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}

func (repository *UsersRepository) GetUser(ctx context.Context, userId string) (user entities.Users, err error) {
	logger.Infof("UserRepository: GetUser")
	db := tx.ExtractTx(ctx)
	if db == nil {
		db = repository.db
	}

	if err = db.Where("user_id = ?", userId).Take(&user).Error; err != nil {
		err = response.NewDatabaseError(err)
		return
	}
	return
}
