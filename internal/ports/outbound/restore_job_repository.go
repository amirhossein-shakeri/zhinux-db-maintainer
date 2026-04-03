package outbound_ports

import (
	"context"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/restore"
)

type RestoreJobRepository interface {
	Save(ctx context.Context, job *restore.RestoreJob) error
	FindByID(ctx context.Context, id string) (*restore.RestoreJob, error)
	ListByTargetDatabaseID(ctx context.Context, databaseID string) ([]*restore.RestoreJob, error)
	MarkStarted(ctx context.Context, id string, startedAt time.Time) error
	MarkFinished(ctx context.Context, id string, status restore.RestoreStatus, finishedAt time.Time) error
}
