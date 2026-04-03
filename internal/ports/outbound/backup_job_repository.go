package outbound_ports

import (
	"context"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
)

type BackupJobRepository interface {
	Save(ctx context.Context, job *backup.BackupJob) error
	FindByID(ctx context.Context, id string) (*backup.BackupJob, error)
	ListByDatabaseID(ctx context.Context, databaseID string) ([]*backup.BackupJob, error)
	MarkStarted(ctx context.Context, id string, startedAt time.Time) error
	MarkFinished(ctx context.Context, id string, status backup.BackupStatus, finishedAt time.Time, artifactID *string) error
}
