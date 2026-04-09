-- name: UpsertDatabase :exec
INSERT INTO
    databases (
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
        $10,
        $11
    ) ON CONFLICT (id) DO
UPDATE
SET
    public_id = EXCLUDED.public_id,
    title = EXCLUDED.title,
    type = EXCLUDED.type,
    host = EXCLUDED.host,
    port = EXCLUDED.port,
    username = EXCLUDED.username,
    password = EXCLUDED.password,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;
