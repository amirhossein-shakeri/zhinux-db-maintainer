package postgres_database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbq "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/adapters/persistence/postgres/gen"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	zhinuxtypes "github.com/amirhossein-shakeri/zhinux-platform/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRepositoryImpl struct {
	q    *dbq.Queries
	pool *pgxpool.Pool
}

func NewDatabaseRepository(
	pool *pgxpool.Pool,
) outbound_ports.DatabaseRepository {
	return &databaseRepositoryImpl{
		q:    dbq.New(pool),
		pool: pool,
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

	err := r.q.UpsertDatabase(ctx, dbq.UpsertDatabaseParams{
		ID:        int64(db.ID),
		Title:     db.Title,
		Type:      string(db.Typ),
		Host:      db.Host,
		Port:      int32(db.Port),
		Username:  db.User,
		Password:  db.Password,
		CreatedAt: toPGTimestamp(db.CreatedAt),
		UpdatedAt: toPGTimestamp(db.UpdatedAt),
		DeletedAt: toPGTimestampPointer(db.DeletedAt),
	})
	return err
}

func (r *databaseRepositoryImpl) FindByID(ctx context.Context, id string) (*database.Database, error) {
	parsedID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.FindDatabaseByID(ctx, parsedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return mapQueryRowToDomain(row), nil
}

func (r *databaseRepositoryImpl) FindByPublicID(ctx context.Context, publicID uuid.UUID) (*database.Database, error) {
	// TODO: Extract utility helper to parse string to pgtype UUID validating through google's UUID
	// parsedUUID, err := uuid.Parse(publicID)
	// if err != nil {
	// 	return nil, err // todo: Wrap error
	// }

	uuid := pgtype.UUID{Bytes: publicID, Valid: true}

	row, err := r.q.FindDatabaseByPublicID(ctx, uuid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return mapQueryRowToDomain(row), nil
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
			parsedIDs := make([]int64, 0, len(f.IDs))
			for _, id := range f.IDs {
				parsedID, parseErr := parseID(id)
				if parseErr != nil {
					return nil, parseErr
				}
				parsedIDs = append(parsedIDs, parsedID)
			}
			addClause("id = ANY($%d)", parsedIDs)
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
	parsedID, err := parseID(id)
	if err != nil {
		return err
	}
	return r.q.SoftDeleteDatabase(ctx, parsedID)
}

func (r *databaseRepositoryImpl) Restore(ctx context.Context, id string) error {
	parsedID, err := parseID(id)
	if err != nil {
		return err
	}
	return r.q.RestoreDatabase(ctx, parsedID)
}

func (r *databaseRepositoryImpl) HardDelete(ctx context.Context, id string) error {
	parsedID, err := parseID(id)
	if err != nil {
		return err
	}
	return r.q.HardDeleteDatabase(ctx, parsedID)
}

func mapQueryRowToDomain(row dbq.Database) *database.Database {
	item := &database.Database{
		ID:       zhinuxtypes.ID(row.ID),
		Title:    row.Title,
		Typ:      database.DatabaseType(row.Type),
		Host:     row.Host,
		Port:     uint(row.Port),
		User:     row.Username,
		Password: row.Password,
	}

	if row.CreatedAt.Valid {
		item.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		item.UpdatedAt = row.UpdatedAt.Time
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		item.DeletedAt = &deletedAt
	}

	return item
}

func parseID(rawID string) (int64, error) {
	id := strings.TrimSpace(rawID)
	if id == "" {
		return 0, fmt.Errorf("id is required")
	}

	parsedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse id %q: %w", rawID, err)
	}

	return parsedID, nil
}

func toPGTimestamp(value time.Time) pgtype.Timestamptz {
	if value.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{
		Time:  value.UTC(),
		Valid: true,
	}
}

func toPGTimestampPointer(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return toPGTimestamp(*value)
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
