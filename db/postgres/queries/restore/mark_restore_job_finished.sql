-- name: MarkRestoreJobFinished :exec
UPDATE restore_jobs
SET status = $2,
    finished_at = $3,
    updated_at = NOW()
WHERE id = $1;
