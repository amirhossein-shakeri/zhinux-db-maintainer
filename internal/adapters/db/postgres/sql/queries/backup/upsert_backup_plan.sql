-- name: UpsertBackupPlan :exec
INSERT INTO backup_plans (
    id,
    database_id,
    schedule,
    enabled,
    retention_policy,
    compression_enabled,
    encryption_enabled,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET schedule = EXCLUDED.schedule,
    enabled = EXCLUDED.enabled,
    retention_policy = EXCLUDED.retention_policy,
    compression_enabled = EXCLUDED.compression_enabled,
    encryption_enabled = EXCLUDED.encryption_enabled,
    updated_at = EXCLUDED.updated_at;
