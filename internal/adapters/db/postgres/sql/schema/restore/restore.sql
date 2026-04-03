CREATE TABLE IF NOT EXISTS restore_jobs (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL REFERENCES backup_artifacts (id) ON DELETE RESTRICT,
    target_database_id TEXT NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'success', 'failed', 'canceled')),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_restore_jobs_target_database_id ON restore_jobs (target_database_id);
CREATE INDEX IF NOT EXISTS idx_restore_jobs_status ON restore_jobs (status);
CREATE INDEX IF NOT EXISTS idx_restore_jobs_artifact_id ON restore_jobs (artifact_id);
