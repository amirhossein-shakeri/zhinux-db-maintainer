-- name: UpsertBackupArtifact :exec
INSERT INTO backup_artifacts (
    id,
    database_id,
    backup_job_id,
    storage_location,
    size_bytes,
    checksum,
    created_at,
    deleted_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET storage_location = EXCLUDED.storage_location,
    size_bytes = EXCLUDED.size_bytes,
    checksum = EXCLUDED.checksum,
    deleted_at = EXCLUDED.deleted_at,
    updated_at = EXCLUDED.updated_at;
