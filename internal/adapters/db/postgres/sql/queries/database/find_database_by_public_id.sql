-- name: FindDatabaseByPublicID :one
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
    public_id = $1;
