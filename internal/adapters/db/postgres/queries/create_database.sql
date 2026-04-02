-- name: CreateDatabase :one
INSERT INTO
    databases (
        id,
        title,
        type,
        host,
        port,
        username,
        password,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    )
RETURNING
    *;

-- name: GetDatabase :one
SELECT * FROM databases WHERE id = $1;

-- name: DeleteDatabase :exec
DELETE FROM databases WHERE id = $1;

-- TODO: Add soft delete, filters, count, etc.
