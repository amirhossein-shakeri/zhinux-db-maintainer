package postgres_backup

import (
	"context"
	"fmt"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type backupJobRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewBackupJobRepository(pool *pgxpool.Pool) outbound_ports.BackupJobRepository {
	return &backupJobRepositoryImpl{pool: pool}
}

func (r *backupJobRepositoryImpl) Save(ctx context.Context, job *backup.BackupJob) error {
	if job == nil {
		return fmt.Errorf("backup job is nil")
	}

	_, err := r.pool.Exec(ctx, backupJobUpsertSQL,
		job.ID,
		job.DatabaseID,
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
	rows, err := r.pool.Query(ctx, backupJobListByDatabaseIDSQL, databaseID)
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
	artifactID *string,
) error {
	_, err := r.pool.Exec(ctx, backupJobMarkFinishedSQL, id, string(status), finishedAt, artifactID)
	return err
}

func scanBackupJob(row interface {
	Scan(dest ...any) error
}) (*backup.BackupJob, error) {
	var item backup.BackupJob
	var trigger string
	var status string

	err := row.Scan(
		&item.ID,
		&item.DatabaseID,
		&trigger,
		&status,
		&item.StartedAt,
		&item.FinishedAt,
		&item.ArtifactID,
	)
	if err != nil {
		return nil, err
	}

	item.TriggerType = backup.BackupTrigger(trigger)
	item.Status = backup.BackupStatus(status)
	return &item, nil
}

const (
	backupJobUpsertSQL = `
INSERT INTO backup_jobs (
	id, database_id, trigger_type, status, started_at, finished_at, artifact_id, created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
) ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	started_at = EXCLUDED.started_at,
	finished_at = EXCLUDED.finished_at,
	artifact_id = EXCLUDED.artifact_id,
	updated_at = EXCLUDED.updated_at
`

	backupJobFindByIDSQL = `
SELECT id, database_id, trigger_type, status, started_at, finished_at, artifact_id
FROM backup_jobs
WHERE id = $1
`

	backupJobListByDatabaseIDSQL = `
SELECT id, database_id, trigger_type, status, started_at, finished_at, artifact_id
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
