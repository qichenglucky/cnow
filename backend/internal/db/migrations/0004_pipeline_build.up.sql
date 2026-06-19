-- 0004: Ensure pipeline and build tables exist (idempotent)

CREATE TABLE IF NOT EXISTS pipeline (
    id          BIGSERIAL PRIMARY KEY,
    repo_id     BIGINT NOT NULL DEFAULT 0,
    service_id  BIGINT NOT NULL REFERENCES service(id),
    config_ref  TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS build (
    id           BIGSERIAL PRIMARY KEY,
    pipeline_id  BIGINT NOT NULL REFERENCES pipeline(id),
    commit_sha   TEXT NOT NULL,
    branch       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending',
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    artifact_url TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_pipeline_service ON pipeline (service_id);
CREATE INDEX IF NOT EXISTS idx_build_pipeline ON build (pipeline_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_build_status ON build (status);
