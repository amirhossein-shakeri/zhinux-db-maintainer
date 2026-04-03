-- name: FindBackupArtifactByID :one
SELECT id, database_id, backup_job_id, storage_location, size_bytes, checksum, created_at, deleted_at
FROM backup_artifacts
WHERE id = $1;
