//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

import (
	"context"
	"time"

	"backend/internal/domain/entities"
)

type IngestRepositoryInputPort interface {
	GetIngestByUserId(ctx context.Context, userId string, ingestDate time.Time, retried bool) (ingest []entities.IngestRuns, err error)
	CreateIngestWithErr(ctx context.Context, ingest entities.IngestRuns) (err error)
}
