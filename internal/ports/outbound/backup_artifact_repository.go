package outbound_ports

import (
	"context"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
)

type BackupArtifactRepository interface {
	Save(ctx context.Context, artifact *backup.BackupArtifact) error
	FindByID(ctx context.Context, id string) (*backup.BackupArtifact, error)
	ListByDatabaseID(ctx context.Context, databaseID string) ([]*backup.BackupArtifact, error)
	SoftDelete(ctx context.Context, id string) error
}
