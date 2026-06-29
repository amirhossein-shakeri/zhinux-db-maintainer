-- name: SoftDeleteBackupArtifact :exec
UPDATE backup_artifacts
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
