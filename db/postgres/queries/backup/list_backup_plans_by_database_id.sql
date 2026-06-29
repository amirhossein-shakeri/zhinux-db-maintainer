-- name: ListBackupPlansByDatabaseID :many
SELECT
    id,
    public_id,
    database_id,
    schedule,
    enabled,
    retention_policy,
    compression_enabled,
    encryption_enabled
FROM backup_plans
WHERE
    database_id = $1
ORDER BY created_at DESC;
