package postgres_backup

import (
	"context"
	"fmt"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
	outbound_ports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type backupPlanRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewBackupPlanRepository(pool *pgxpool.Pool) outbound_ports.BackupPlanRepository {
	return &backupPlanRepositoryImpl{pool: pool}
}

func (r *backupPlanRepositoryImpl) Save(ctx context.Context, plan *backup.BackupPlan) error {
	if plan == nil {
		return fmt.Errorf("backup plan is nil")
	}

	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, backupPlanUpsertSQL,
		plan.ID,
		plan.DatabaseID,
		plan.Schedule,
		plan.Enabled,
		string(plan.RetentionPolicy),
		plan.CompressionEnabled,
		plan.EncryptionEnabled,
		now,
		now,
	)
	return err
}

func (r *backupPlanRepositoryImpl) FindByID(ctx context.Context, id string) (*backup.BackupPlan, error) {
	row := r.pool.QueryRow(ctx, backupPlanFindByIDSQL, id)
	plan, err := scanBackupPlan(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return plan, nil
}

func (r *backupPlanRepositoryImpl) ListByDatabaseID(ctx context.Context, databaseID string) ([]*backup.BackupPlan, error) {
	rows, err := r.pool.Query(ctx, backupPlanListByDatabaseIDSQL, databaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*backup.BackupPlan, 0, 8)
	for rows.Next() {
		item, scanErr := scanBackupPlan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *backupPlanRepositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, backupPlanDeleteSQL, id)
	return err
}

func scanBackupPlan(row interface {
	Scan(dest ...any) error
}) (*backup.BackupPlan, error) {
	var item backup.BackupPlan
	var retentionPolicy string

	err := row.Scan(
		&item.ID,
		&item.DatabaseID,
		&item.Schedule,
		&item.Enabled,
		&retentionPolicy,
		&item.CompressionEnabled,
		&item.EncryptionEnabled,
	)
	if err != nil {
		return nil, err
	}

	item.RetentionPolicy = shared.RetentionPolicy(retentionPolicy)
	return &item, nil
}

const (
	backupPlanUpsertSQL = `
INSERT INTO backup_plans (
	id, database_id, schedule, enabled, retention_policy, compression_enabled, encryption_enabled, created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
) ON CONFLICT (id) DO UPDATE SET
	schedule = EXCLUDED.schedule,
	enabled = EXCLUDED.enabled,
	retention_policy = EXCLUDED.retention_policy,
	compression_enabled = EXCLUDED.compression_enabled,
	encryption_enabled = EXCLUDED.encryption_enabled,
	updated_at = EXCLUDED.updated_at
`

	backupPlanFindByIDSQL = `
SELECT id, database_id, schedule, enabled, retention_policy, compression_enabled, encryption_enabled
FROM backup_plans
WHERE id = $1
`

	backupPlanListByDatabaseIDSQL = `
SELECT id, database_id, schedule, enabled, retention_policy, compression_enabled, encryption_enabled
FROM backup_plans
WHERE database_id = $1
ORDER BY created_at DESC
`

	backupPlanDeleteSQL = `
DELETE FROM backup_plans WHERE id = $1
`
)
