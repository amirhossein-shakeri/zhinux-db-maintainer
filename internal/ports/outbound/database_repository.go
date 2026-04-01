package outbound_ports

import (
	"context"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
)

type DatabaseRepository interface {
	Save(ctx context.Context, db *database.Database) error
	FindByID(ctx context.Context, id string) (*database.Database, error)
	Delete(ctx context.Context, id string) error
}
