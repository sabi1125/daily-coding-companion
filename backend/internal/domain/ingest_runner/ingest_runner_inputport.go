//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package ingestrunner

import "context"

type IngestRunnerInputPort interface {
	Ingest(ctx context.Context, userIds []string, retried bool) error
}
