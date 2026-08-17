//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package util

import "time"

// TimeProvider wraps time.Now() behind an interface — same reasoning as
// UUIDGenerator: inject it instead of calling time.Now() inline, so tests
// can mock a known, fixed "now" instead of depending on the real wall
// clock (critical for anything asserting on computed expiry times).
type TimeProvider interface {
	Now() time.Time
	ExpiryTimeCalculator() time.Time
	TodaysDate() time.Time
}

var JST = time.FixedZone("JST", 9*60*60)

type timeProvider struct{}

func NewTimeProvider() TimeProvider {
	return &timeProvider{}
}

func (t *timeProvider) Now() time.Time {
	return time.Now()
}

func (t *timeProvider) ExpiryTimeCalculator() time.Time {
	now := time.Now()
	expiryTime := now.AddDate(0, 0, 30)
	return expiryTime
}

func (t *timeProvider) TodaysDate() time.Time {
	now := time.Now().In(JST)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, JST)
}
