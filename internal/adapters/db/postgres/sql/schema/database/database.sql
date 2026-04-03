CREATE TABLE IF NOT EXISTS databases (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('postgres', 'mongo')),
    host TEXT NOT NULL,
    port INT NOT NULL CHECK (port > 0),
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_databases_type ON databases (type);
CREATE INDEX IF NOT EXISTS idx_databases_host_port ON databases (host, port);
CREATE INDEX IF NOT EXISTS idx_databases_deleted_at ON databases (deleted_at);
