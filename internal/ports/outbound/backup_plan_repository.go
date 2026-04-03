package outbound_ports

import (
	"context"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
)

type BackupPlanRepository interface {
	Save(ctx context.Context, plan *backup.BackupPlan) error
	FindByID(ctx context.Context, id string) (*backup.BackupPlan, error)
	ListByDatabaseID(ctx context.Context, databaseID string) ([]*backup.BackupPlan, error)
	Delete(ctx context.Context, id string) error
}
