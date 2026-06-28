-- name: MarkRestoreJobStarted :exec
UPDATE restore_jobs
SET status = 'in_progress',
    started_at = $2,
    updated_at = NOW()
WHERE id = $1;
