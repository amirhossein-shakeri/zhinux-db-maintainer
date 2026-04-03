-- name: FindBackupJobByID :one
SELECT id, database_id, trigger_type, status, started_at, finished_at, artifact_id
FROM backup_jobs
WHERE id = $1;
