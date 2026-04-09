-- name: SoftDeleteDatabase :exec
UPDATE databases
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1;
