-- name: UpsertDatabase :exec
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
        updated_at,
        deleted_at
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
        $9,
        $10
    ) ON CONFLICT (id) DO
UPDATE
SET
    title = EXCLUDED.title,
    type = EXCLUDED.type,
    host = EXCLUDED.host,
    port = EXCLUDED.port,
    username = EXCLUDED.username,
    password = EXCLUDED.password,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;
