package infrastructure

import (
	"context"

	"backend/internal/tx"

	"gorm.io/gorm"
)

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) tx.Manager {
	return &TransactionManager{
		db: db,
	}
}

func (m *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		return fn(tx.WithTx(ctx, gormTx))
	})
}
