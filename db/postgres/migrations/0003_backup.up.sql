CREATE TABLE IF NOT EXISTS backup_plans (
    id TEXT PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    database_id BIGINT NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
    schedule TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    retention_policy TEXT NOT NULL CHECK (
        retention_policy IN ('keep_last', 'max_age')
    ),
    compression_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    encryption_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS backup_jobs (
    id TEXT PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid (),
    database_id BIGINT NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
    trigger_type TEXT NOT NULL CHECK (
        trigger_type IN ('manual', 'scheduled')
    ),
    status TEXT NOT NULL CHECK (
        status IN (
            'pending',
            'in_progress',
            'success',
            'failed',
            'canceled'
        )
    ),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    artifact_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS backup_artifacts (
    id TEXT PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid (),
    database_id BIGINT NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
    backup_job_id TEXT NOT NULL REFERENCES backup_jobs (id) ON DELETE CASCADE,
    storage_location TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    checksum TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_backup_plans_database_id ON backup_plans (database_id);

CREATE INDEX IF NOT EXISTS idx_backup_plans_public_id ON backup_plans (public_id);

CREATE INDEX IF NOT EXISTS idx_backup_jobs_database_id ON backup_jobs (database_id);

CREATE INDEX IF NOT EXISTS idx_backup_jobs_public_id ON backup_jobs (public_id);

CREATE INDEX IF NOT EXISTS idx_backup_jobs_status ON backup_jobs (status);

CREATE INDEX IF NOT EXISTS idx_backup_jobs_artifact_id ON backup_jobs (artifact_id);

CREATE INDEX IF NOT EXISTS idx_backup_artifacts_database_id ON backup_artifacts (database_id);

CREATE INDEX IF NOT EXISTS idx_backup_artifacts_public_id ON backup_artifacts (public_id);

CREATE INDEX IF NOT EXISTS idx_backup_artifacts_backup_job_id ON backup_artifacts (backup_job_id);

CREATE INDEX IF NOT EXISTS idx_backup_artifacts_deleted_at ON backup_artifacts (deleted_at);
