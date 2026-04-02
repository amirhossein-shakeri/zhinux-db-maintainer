-- name: SoftDeleteDatabase :exec
UPDATE databases SET deleted_at = NOW() WHERE id = $1;
