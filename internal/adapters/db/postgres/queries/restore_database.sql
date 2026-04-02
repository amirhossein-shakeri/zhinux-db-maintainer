-- name: RestoreDatabase :exec
UPDATE databases SET deleted_at = NULL WHERE id = $1;
