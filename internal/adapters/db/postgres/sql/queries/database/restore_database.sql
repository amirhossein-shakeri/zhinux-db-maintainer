-- name: RestoreDatabase :exec
UPDATE databases
SET deleted_at = NULL,
    updated_at = NOW()
WHERE id = $1;
