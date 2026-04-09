package postgres_restore

import (
	"context"
	"fmt"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/restore"
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type restoreJobRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRestoreJobRepository(pool *pgxpool.Pool) outbound_ports.RestoreJobRepository {
	return &restoreJobRepositoryImpl{pool: pool}
}

func (r *restoreJobRepositoryImpl) Save(ctx context.Context, job *restore.RestoreJob) error {
	if job == nil {
		return fmt.Errorf("restore job is nil")
	}

	_, err := r.pool.Exec(ctx, restoreJobUpsertSQL,
		job.ID,
		job.ArtifactID,
		job.TargetDatabaseID,
		string(job.Status),
		job.StartedAt,
		job.FinishedAt,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	return err
}

func (r *restoreJobRepositoryImpl) FindByID(ctx context.Context, id string) (*restore.RestoreJob, error) {
	row := r.pool.QueryRow(ctx, restoreJobFindByIDSQL, id)
	item, err := scanRestoreJob(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *restoreJobRepositoryImpl) ListByTargetDatabaseID(ctx context.Context, databaseID string) ([]*restore.RestoreJob, error) {
	rows, err := r.pool.Query(ctx, restoreJobListByTargetDatabaseIDSQL, databaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*restore.RestoreJob, 0, 16)
	for rows.Next() {
		item, scanErr := scanRestoreJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *restoreJobRepositoryImpl) MarkStarted(ctx context.Context, id string, startedAt time.Time) error {
	_, err := r.pool.Exec(ctx, restoreJobMarkStartedSQL, id, startedAt)
	return err
}

func (r *restoreJobRepositoryImpl) MarkFinished(
	ctx context.Context,
	id string,
	status restore.RestoreStatus,
	finishedAt time.Time,
) error {
	_, err := r.pool.Exec(ctx, restoreJobMarkFinishedSQL, id, string(status), finishedAt)
	return err
}

func scanRestoreJob(row interface {
	Scan(dest ...any) error
}) (*restore.RestoreJob, error) {
	var item restore.RestoreJob
	var status string

	err := row.Scan(
		&item.ID,
		&item.ArtifactID,
		&item.TargetDatabaseID,
		&status,
		&item.StartedAt,
		&item.FinishedAt,
	)
	if err != nil {
		return nil, err
	}

	item.Status = restore.RestoreStatus(status)
	return &item, nil
}

const (
	restoreJobUpsertSQL = `
INSERT INTO restore_jobs (
	id, artifact_id, target_database_id, status, started_at, finished_at, created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	started_at = EXCLUDED.started_at,
	finished_at = EXCLUDED.finished_at,
	updated_at = EXCLUDED.updated_at
`

	restoreJobFindByIDSQL = `
SELECT id, artifact_id, target_database_id, status, started_at, finished_at
FROM restore_jobs
WHERE id = $1
`

	restoreJobListByTargetDatabaseIDSQL = `
SELECT id, artifact_id, target_database_id, status, started_at, finished_at
FROM restore_jobs
WHERE target_database_id = $1
ORDER BY created_at DESC
`

	restoreJobMarkStartedSQL = `
UPDATE restore_jobs
SET status = 'in_progress', started_at = $2, updated_at = NOW()
WHERE id = $1
`

	restoreJobMarkFinishedSQL = `
UPDATE restore_jobs
SET status = $2, finished_at = $3, updated_at = NOW()
WHERE id = $1
`
)
