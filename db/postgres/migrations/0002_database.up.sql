CREATE TABLE IF NOT EXISTS databases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), -- Postgres 17+ generates UUIDv7
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('postgres', 'mongo')), -- TODO: Use DB Enums
    host TEXT NOT NULL,
    port INT NOT NULL CHECK (port > 0),
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_databases_public_id ON databases (public_id);

CREATE INDEX IF NOT EXISTS idx_databases_type ON databases (type);

CREATE INDEX IF NOT EXISTS idx_databases_host_port ON databases (host, port);

CREATE INDEX IF NOT EXISTS idx_databases_deleted_at ON databases (deleted_at);
