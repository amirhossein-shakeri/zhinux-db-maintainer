-- name: ListBackupArtifactsByDatabaseID :many
SELECT
    id,
    public_id,
    database_id,
    backup_job_id,
    storage_location,
    size_bytes,
    checksum,
    created_at,
    deleted_at
FROM backup_artifacts
WHERE
    database_id = $1
ORDER BY created_at DESC;
