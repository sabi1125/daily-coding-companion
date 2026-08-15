//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package tx

import (
	"context"

	"gorm.io/gorm"
)

// Manager lets a caller run a function inside a single database transaction
// without that function needing to know about *gorm.DB directly. Lives in
// its own package so the domain layer (interactors) can depend on this
// interface without importing internal/infrastructure — same reasoning as
// the interactor/repository inputport split.
type Manager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ctxKey string

const gormTxKey ctxKey = "gorm_tx"

// WithTx stashes a transaction-scoped *gorm.DB in ctx. Called by
// infrastructure.TransactionManager.WithinTransaction — kept here, not in
// infrastructure, so repository code can import this package for ExtractTx
// without an import cycle (infrastructure's router.go already imports
// domain/repository, so repository can't import infrastructure back).
func WithTx(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, gormTxKey, db)
}

// ExtractTx returns the transaction-scoped *gorm.DB stashed in ctx by
// WithinTransaction, or nil if there isn't one — callers fall back to
// their own db when nil, so repository methods work the same whether or
// not they're running inside a transaction.
func ExtractTx(ctx context.Context) *gorm.DB {
	db, _ := ctx.Value(gormTxKey).(*gorm.DB)
	return db
}
