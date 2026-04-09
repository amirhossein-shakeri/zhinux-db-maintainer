package postgres_database

import (
	"context"
	"fmt"
	"strings"
	"time"

	databaseq "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/adapters/db/postgres/gen/database"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRepositoryImpl struct {
	q *databaseq.Queries
}

func NewDatabaseRepository(
	pool *pgxpool.Pool,
) outbound_ports.DatabaseRepository {
	return &databaseRepositoryImpl{
		q: databaseq.New(pool),
	}
}

func (r *databaseRepositoryImpl) Save(
	ctx context.Context, db *database.Database,
) error {
	// Reject if entity is not provided
	if db == nil {
		return fmt.Errorf("database is nil") // todo: use named errors
	}

	now := time.Now().UTC()

	// Set created_at if not set already
	if db.CreatedAt.IsZero() {
		db.CreatedAt = now
	}

	// Always update updated_at to now
	db.UpdatedAt = now

	err := r.q.UpsertDatabase(ctx, databaseq.UpsertDatabaseParams{
		ID:        int64(db.ID),
		Title:     db.Title,
		Type:      string(db.Typ),
		Host:      db.Host,
		Port:      int32(db.Port),
		Username:  db.User,
		Password:  db.Password,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
		DeletedAt: db.DeletedAt,
	})
	return err
}

func (r *databaseRepositoryImpl) FindByID(ctx context.Context, id string) (*database.Database, error) {
	row := r.pool.QueryRow(ctx, databaseFindByIDSQL, id)

	item, err := scanDatabase(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return item, nil
}

func (r *databaseRepositoryImpl) Filter(
	ctx context.Context,
	f *database.Filter,
	p *shared.Pagination,
) (*shared.Result[*database.Database], error) {
	clauses := make([]string, 0, 8)
	args := make([]any, 0, 16)
	argPos := 1

	addClause := func(sql string, arg any) {
		clauses = append(clauses, fmt.Sprintf(sql, argPos))
		args = append(args, arg)
		argPos++
	}

	if f != nil {
		if len(f.IDs) > 0 {
			addClause("id = ANY($%d)", f.IDs)
		}
		if f.Title != nil {
			addClause("title ILIKE ('%%' || $%d || '%%')", *f.Title)
		}
		if len(f.Types) > 0 {
			types := make([]string, 0, len(f.Types))
			for _, t := range f.Types {
				types = append(types, string(t))
			}
			addClause(`type = ANY($%d)`, types)
		}
		if f.Host != nil {
			addClause("host = $%d", *f.Host)
		}
		if f.Port != nil {
			addClause("port = $%d", int32(*f.Port))
		}
		if f.User != nil {
			addClause("username = $%d", *f.User)
		}
		if f.Password != nil {
			addClause("password = $%d", *f.Password)
		}

		if f.CreatedAtSince != nil {
			addClause("created_at >= $%d", *f.CreatedAtSince)
		}
		if f.CreatedAtUntil != nil {
			addClause("created_at <= $%d", *f.CreatedAtUntil)
		}
		if f.UpdatedAtSince != nil {
			addClause("updated_at >= $%d", *f.UpdatedAtSince)
		}
		if f.UpdatedAtUntil != nil {
			addClause("updated_at <= $%d", *f.UpdatedAtUntil)
		}
		if f.DeletedAtSince != nil {
			addClause("deleted_at >= $%d", *f.DeletedAtSince)
		}
		if f.DeletedAtUntil != nil {
			addClause("deleted_at <= $%d", *f.DeletedAtUntil)
		}
	}

	includeDeleted := false
	if f != nil && f.Deleted != nil {
		includeDeleted = *f.Deleted
	}
	if includeDeleted {
		clauses = append(clauses, "deleted_at IS NOT NULL")
	} else {
		clauses = append(clauses, "deleted_at IS NULL")
	}

	whereSQL := ""
	if len(clauses) > 0 {
		whereSQL = " WHERE " + strings.Join(clauses, " AND ")
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM databases" + whereSQL
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	limit := 50
	offset := 0
	if p != nil {
		if p.Limit > 0 {
			limit = p.Limit
		}
		if p.Offset > 0 {
			offset = p.Offset
		}
	}

	querySQL := "SELECT id, title, type, host, port, username, password, created_at, updated_at, deleted_at FROM databases" +
		whereSQL +
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*database.Database, 0, limit)
	for rows.Next() {
		item, scanErr := scanDatabase(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	result := &shared.Result[*database.Database]{
		Items: items,
		Total: total,
	}
	if p != nil {
		result.Pagination = *p
		result.Pagination.Total = total
	}

	return result, nil
}

func (r *databaseRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, databaseSoftDeleteSQL, id)
	return err
}

func (r *databaseRepositoryImpl) Restore(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, databaseRestoreSQL, id)
	return err
}

func (r *databaseRepositoryImpl) HardDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, databaseHardDeleteSQL, id)
	return err
}

func scanDatabase(row interface {
	Scan(dest ...any) error
}) (*database.Database, error) {
	var item database.Database
	var typ string

	err := row.Scan(
		&item.ID,
		&item.Title,
		&typ,
		&item.Host,
		&item.Port,
		&item.User,
		&item.Password,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	item.Typ = database.DatabaseType(typ)
	return &item, nil
}

// const (
// 	databaseUpsertSQL = `
// INSERT INTO databases (
// 	id, title, type, host, port, username, password, created_at, updated_at, deleted_at
// ) VALUES (
// 	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
// ) ON CONFLICT (id) DO UPDATE SET
// 	title = EXCLUDED.title,
// 	type = EXCLUDED.type,
// 	host = EXCLUDED.host,
// 	port = EXCLUDED.port,
// 	username = EXCLUDED.username,
// 	password = EXCLUDED.password,
// 	updated_at = EXCLUDED.updated_at,
// 	deleted_at = EXCLUDED.deleted_at
// `

// 	databaseFindByIDSQL = `
// SELECT id, title, type, host, port, username, password, created_at, updated_at, deleted_at
// FROM databases
// WHERE id = $1
// `

// 	databaseSoftDeleteSQL = `
// UPDATE databases
// SET deleted_at = NOW(), updated_at = NOW()
// WHERE id = $1
// `

// 	databaseRestoreSQL = `
// UPDATE databases
// SET deleted_at = NULL, updated_at = NOW()
// WHERE id = $1
// `

// 	databaseHardDeleteSQL = `
// DELETE FROM databases WHERE id = $1
// `
// )
