//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package util

import "github.com/google/uuid"

// UUIDGenerator is injected wherever a new identifier is needed — wrapping
// uuid.NewV7 behind an interface (rather than calling it inline) lets tests
// mock a known, predictable value instead of a random one.
type UUIDGenerator interface {
	NewV7() (string, error)
}

type uuidGenerator struct{}

func NewUUIDGenerator() UUIDGenerator {
	return &uuidGenerator{}
}

// NewV7 returns a time-ordered (roughly sortable by creation time) UUID —
// better for DB primary keys than a random v4, since it avoids scattering
// inserts across a B-tree index.
func (g *uuidGenerator) NewV7() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
