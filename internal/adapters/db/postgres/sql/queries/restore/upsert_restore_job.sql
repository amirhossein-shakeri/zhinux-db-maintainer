-- name: UpsertRestoreJob :exec
INSERT INTO
    restore_jobs (
        id,
        public_id,
        artifact_id,
        target_database_id,
        status,
        started_at,
        finished_at,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    ) ON CONFLICT (id) DO
UPDATE
SET
    public_id = EXCLUDED.public_id,
    status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    finished_at = EXCLUDED.finished_at,
    updated_at = EXCLUDED.updated_at;
