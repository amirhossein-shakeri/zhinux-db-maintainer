-- name: ListActiveDatabases :many
SELECT
    id,
    public_id,
    title,
    type,
    host,
    port,
    username,
    password,
    created_at,
    updated_at,
    deleted_at
FROM databases
WHERE
    deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1
OFFSET
    $2;
