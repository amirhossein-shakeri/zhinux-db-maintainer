-- name: UpsertRestoreJob :exec
INSERT INTO restore_jobs (
    id,
    artifact_id,
    target_database_id,
    status,
    started_at,
    finished_at,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    finished_at = EXCLUDED.finished_at,
    updated_at = EXCLUDED.updated_at;
