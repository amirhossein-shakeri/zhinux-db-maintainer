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

type backupJobRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewBackupJobRepository(pool *pgxpool.Pool) outboundports.BackupJobRepository {
	return &backupJobRepositoryImpl{pool: pool}
}

func (r *backupJobRepositoryImpl) Save(ctx context.Context, job *backup.BackupJob) error {
	if job == nil {
		return fmt.Errorf("backup job is nil")
	}

	publicID, err := parsePublicID(job.PublicID)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, backupJobUpsertSQL,
		job.ID,
		publicID,
		int64(job.DatabaseID),
		string(job.TriggerType),
		string(job.Status),
		job.StartedAt,
		job.FinishedAt,
		job.ArtifactID,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	return err
}

func (r *backupJobRepositoryImpl) FindByID(ctx context.Context, id string) (*backup.BackupJob, error) {
	row := r.pool.QueryRow(ctx, backupJobFindByIDSQL, id)
	item, err := scanBackupJob(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *backupJobRepositoryImpl) ListByDatabaseID(ctx context.Context, databaseID string) ([]*backup.BackupJob, error) {
	parsedDatabaseID, err := parseDatabaseID(databaseID)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, backupJobListByDatabaseIDSQL, parsedDatabaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*backup.BackupJob, 0, 16)
	for rows.Next() {
		item, scanErr := scanBackupJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *backupJobRepositoryImpl) MarkStarted(ctx context.Context, id string, startedAt time.Time) error {
	_, err := r.pool.Exec(ctx, backupJobMarkStartedSQL, id, startedAt)
	return err
}

func (r *backupJobRepositoryImpl) MarkFinished(
	ctx context.Context,
	id string,
	status backup.BackupStatus,
	finishedAt time.Time,
	artifactID *zhinuxtypes.ID,
) error {
	var artifactIDValue *int64
	if artifactID != nil {
		value := int64(*artifactID)
		artifactIDValue = &value
	}
	_, err := r.pool.Exec(ctx, backupJobMarkFinishedSQL, id, string(status), finishedAt, artifactIDValue)
	return err
}

func scanBackupJob(row interface {
	Scan(dest ...any) error
}) (*backup.BackupJob, error) {
	var item backup.BackupJob
	var trigger string
	var status string
	var databaseID int64
	var artifactID *int64

	err := row.Scan(
		&item.ID,
		&item.PublicID,
		&databaseID,
		&trigger,
		&status,
		&item.StartedAt,
		&item.FinishedAt,
		&artifactID,
	)
	if err != nil {
		return nil, err
	}

	item.DatabaseID = zhinuxtypes.ID(databaseID)
	item.TriggerType = backup.BackupTrigger(trigger)
	item.Status = backup.BackupStatus(status)
	if artifactID != nil {
		typedArtifactID := zhinuxtypes.ID(*artifactID)
		item.ArtifactID = &typedArtifactID
	}
	return &item, nil
}

const (
	backupJobUpsertSQL = `
INSERT INTO backup_jobs (
	id, public_id, database_id, trigger_type, status, started_at, finished_at, artifact_id, created_at, updated_at
) VALUES (
	$1, COALESCE($2, gen_random_uuid()), $3, $4, $5, $6, $7, $8, $9, $10
) ON CONFLICT (id) DO UPDATE SET
	public_id = EXCLUDED.public_id,
	status = EXCLUDED.status,
	started_at = EXCLUDED.started_at,
	finished_at = EXCLUDED.finished_at,
	artifact_id = EXCLUDED.artifact_id,
	updated_at = EXCLUDED.updated_at
`

	backupJobFindByIDSQL = `
SELECT id, public_id::text, database_id, trigger_type, status, started_at, finished_at, artifact_id
FROM backup_jobs
WHERE id = $1
`

	backupJobListByDatabaseIDSQL = `
SELECT id, public_id::text, database_id, trigger_type, status, started_at, finished_at, artifact_id
FROM backup_jobs
WHERE database_id = $1
ORDER BY created_at DESC
`

	backupJobMarkStartedSQL = `
UPDATE backup_jobs
SET status = 'in_progress', started_at = $2, updated_at = NOW()
WHERE id = $1
`

	backupJobMarkFinishedSQL = `
UPDATE backup_jobs
SET status = $2, finished_at = $3, artifact_id = $4, updated_at = NOW()
WHERE id = $1
`
)
