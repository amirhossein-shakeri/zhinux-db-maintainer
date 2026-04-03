package postgres

import (
	"context"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRepositoryImpl struct {
	q *Queries
}

// TODO: Add logging

func NewDatabaseRepository(pool *pgxpool.Pool) outbound_ports.DatabaseRepository {
	return &databaseRepositoryImpl{
		q: New(pool),
	}
}

func (r *databaseRepositoryImpl) Save(
	ctx context.Context, db *database.Database,
) error {
	//
}

func (r *databaseRepositoryImpl) FindByID(
	ctx context.Context, id string,
) (*database.Database, error) {
	row, err := r.q.GetDatabase(ctx, id)
	if err != nil {
		return nil, err
	}

	return &database.Database{
		ID:       row.ID,
		Title:    row.Title,
		Host:     row.Host,
		Port:     uint(row.Port),
		User:     row.Username,
		Password: row.Password,
	}, nil
}
