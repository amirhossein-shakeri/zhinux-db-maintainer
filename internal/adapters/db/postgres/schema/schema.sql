CREATE TABLE databases (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    type TEXT NOT NULL,
    host TEXT NOT NULL,
    port INT NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- todo: How about migrations or schema changes?
-- todo: How about idempotency and `IF NOT EXISTS`
