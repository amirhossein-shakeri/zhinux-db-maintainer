-- name: DeleteBackupPlan :exec
DELETE FROM backup_plans
WHERE id = $1;
