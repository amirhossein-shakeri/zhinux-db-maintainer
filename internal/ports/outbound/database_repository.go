package outbound_ports

import (
	"context"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
)

type DatabaseRepository interface {
	// Create or update a database
	Save(ctx context.Context, db *database.Database) error

	// Try finding a specific database by ID, if not found, returns nil, nil not throwing an error like `GetByID`
	FindByID(ctx context.Context, id string) (*database.Database, error)

	// Filter a list of databases with the given filter and pagination
	Filter(ctx context.Context, f *database.Filter,
		p *shared.Pagination) (*shared.Result[*database.Database], error)

	// Mark a database as deleted / moved to trash
	SoftDelete(ctx context.Context, id string) error

	// Restore a soft-deleted database / mark as not deleted
	Restore(ctx context.Context, id string) error

	// Hard deletes a database record physically
	HardDelete(ctx context.Context, id string) error
}
