-- name: ListRestoreJobsByTargetDatabaseID :many
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
    target_database_id = $1
ORDER BY created_at DESC;
