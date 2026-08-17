//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package ingestrunner

import "context"

type IngestRunnerInputPort interface {
	RunForUser(ctx context.Context, userId string, retried bool) error
}
