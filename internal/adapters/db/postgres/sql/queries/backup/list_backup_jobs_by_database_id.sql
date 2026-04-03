-- name: ListBackupJobsByDatabaseID :many
SELECT id, database_id, trigger_type, status, started_at, finished_at, artifact_id
FROM backup_jobs
WHERE database_id = $1
ORDER BY created_at DESC;
