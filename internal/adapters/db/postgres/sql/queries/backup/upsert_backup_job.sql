-- name: UpsertBackupJob :exec
INSERT INTO
    backup_jobs (
        id,
        public_id,
        database_id,
        trigger_type,
        status,
        started_at,
        finished_at,
        artifact_id,
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
        $9,
        $10
    ) ON CONFLICT (id) DO
UPDATE
SET
    public_id = EXCLUDED.public_id,
    status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    finished_at = EXCLUDED.finished_at,
    artifact_id = EXCLUDED.artifact_id,
    updated_at = EXCLUDED.updated_at;
