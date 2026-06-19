-- 0002: Add constraints, indexes, and idempotency table

-- Service name uniqueness
ALTER TABLE service ADD CONSTRAINT uq_service_name UNIQUE (name);

-- Environment uniqueness per service
ALTER TABLE environment ADD CONSTRAINT uq_env_service_name UNIQUE (service_id, name);

-- Release uniqueness per service + version
ALTER TABLE "release" ADD CONSTRAINT uq_release_service_version UNIQUE (service_id, version);

-- Status enum constraints
ALTER TABLE service ADD CONSTRAINT chk_service_status
    CHECK (status IN ('draft', 'creating', 'ready', 'degraded', 'archived'));
ALTER TABLE "release" ADD CONSTRAINT chk_release_status
    CHECK (status IN ('created', 'reviewing', 'approved', 'deploying', 'verifying', 'observing', 'succeeded', 'failed', 'rollback_pending', 'rolling_back', 'rolled_back'));
ALTER TABLE approval ADD CONSTRAINT chk_approval_status
    CHECK (status IN ('pending', 'approved', 'rejected', 'expired'));
ALTER TABLE build ADD CONSTRAINT chk_build_status
    CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled'));
ALTER TABLE incident ADD CONSTRAINT chk_incident_severity
    CHECK (severity IN ('low', 'medium', 'high', 'critical'));

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_release_status ON "release" (status);
CREATE INDEX IF NOT EXISTS idx_release_created_at ON "release" (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_release_service_id ON "release" (service_id);
CREATE INDEX IF NOT EXISTS idx_release_env_id ON "release" (environment_id);
CREATE INDEX IF NOT EXISTS idx_release_event_release ON release_event (release_id, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_action_role ON audit_log (action, actor_role);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_log (resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_ai_run_service ON ai_run (service_id, created_at);
CREATE INDEX IF NOT EXISTS idx_build_pipeline ON build (pipeline_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_env_service ON environment (service_id);
CREATE INDEX IF NOT EXISTS idx_incident_service ON incident (service_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_approval_release ON approval (release_id);

-- Idempotency table
CREATE TABLE IF NOT EXISTS idempotency_key (
    key          TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    response_json JSONB,
    status       TEXT NOT NULL DEFAULT 'reserved',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_idempotency_expires ON idempotency_key (expires_at);
