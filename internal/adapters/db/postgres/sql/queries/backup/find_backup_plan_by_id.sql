-- name: FindBackupPlanByID :one
SELECT id, database_id, schedule, enabled, retention_policy, compression_enabled, encryption_enabled
FROM backup_plans
WHERE id = $1;
