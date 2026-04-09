package postgres_backup

import (
	"context"
	"fmt"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
	outboundports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	zhinuxtypes "github.com/amirhossein-shakeri/zhinux-platform/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type backupArtifactRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewBackupArtifactRepository(pool *pgxpool.Pool) outboundports.BackupArtifactRepository {
	return &backupArtifactRepositoryImpl{pool: pool}
}

func (r *backupArtifactRepositoryImpl) Save(ctx context.Context, artifact *backup.BackupArtifact) error {
	if artifact == nil {
		return fmt.Errorf("backup artifact is nil")
	}

	now := time.Now().UTC()
	if artifact.CreatedAt.IsZero() {
		artifact.CreatedAt = now
	}

	publicID, err := parsePublicID(artifact.PublicID)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, backupArtifactUpsertSQL,
		artifact.ID,
		publicID,
		int64(artifact.DatabaseID),
		artifact.BackupJobID,
		artifact.StorageLocation,
		artifact.Size,
		artifact.Checksum,
		artifact.CreatedAt,
		artifact.DeletedAt,
		now,
	)
	return err
}

func (r *backupArtifactRepositoryImpl) FindByID(ctx context.Context, id string) (*backup.BackupArtifact, error) {
	row := r.pool.QueryRow(ctx, backupArtifactFindByIDSQL, id)
	item, err := scanBackupArtifact(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *backupArtifactRepositoryImpl) ListByDatabaseID(ctx context.Context, databaseID string) ([]*backup.BackupArtifact, error) {
	parsedDatabaseID, err := parseDatabaseID(databaseID)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, backupArtifactListByDatabaseIDSQL, parsedDatabaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*backup.BackupArtifact, 0, 16)
	for rows.Next() {
		item, scanErr := scanBackupArtifact(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *backupArtifactRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, backupArtifactSoftDeleteSQL, id)
	return err
}

func scanBackupArtifact(row interface {
	Scan(dest ...any) error
}) (*backup.BackupArtifact, error) {
	var item backup.BackupArtifact
	var databaseID int64

	err := row.Scan(
		&item.ID,
		&item.PublicID,
		&databaseID,
		&item.BackupJobID,
		&item.StorageLocation,
		&item.Size,
		&item.Checksum,
		&item.CreatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	item.DatabaseID = zhinuxtypes.ID(databaseID)
	return &item, nil
}

const (
	backupArtifactUpsertSQL = `
INSERT INTO backup_artifacts (
	id, public_id, database_id, backup_job_id, storage_location, size_bytes, checksum, created_at, deleted_at, updated_at
) VALUES (
	$1, COALESCE($2, gen_random_uuid()), $3, $4, $5, $6, $7, $8, $9, $10
) ON CONFLICT (id) DO UPDATE SET
	public_id = EXCLUDED.public_id,
	storage_location = EXCLUDED.storage_location,
	size_bytes = EXCLUDED.size_bytes,
	checksum = EXCLUDED.checksum,
	deleted_at = EXCLUDED.deleted_at,
	updated_at = EXCLUDED.updated_at
`

	backupArtifactFindByIDSQL = `
SELECT id, public_id::text, database_id, backup_job_id, storage_location, size_bytes, checksum, created_at, deleted_at
FROM backup_artifacts
WHERE id = $1
`

	backupArtifactListByDatabaseIDSQL = `
SELECT id, public_id::text, database_id, backup_job_id, storage_location, size_bytes, checksum, created_at, deleted_at
FROM backup_artifacts
WHERE database_id = $1
ORDER BY created_at DESC
`

	backupArtifactSoftDeleteSQL = `
UPDATE backup_artifacts
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1
`
)
