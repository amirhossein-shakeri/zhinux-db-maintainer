-- name: FindRestoreJobByID :one
SELECT
    id,
    public_id,
    artifact_id,
    target_database_id,
    status,
    started_at,
    finished_at
FROM restore_jobs
WHERE
    id = $1;
