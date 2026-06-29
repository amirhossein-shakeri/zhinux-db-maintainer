-- name: MarkBackupJobFinished :exec
UPDATE backup_jobs
SET status = $2,
    finished_at = $3,
    artifact_id = $4,
    updated_at = NOW()
WHERE id = $1;
