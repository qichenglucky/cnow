CREATE TABLE service (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL UNIQUE,
  display_name VARCHAR(128) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  owner_id BIGINT NOT NULL DEFAULT 0,
  team_id BIGINT NOT NULL DEFAULT 0,
  tech_stack VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  default_repo_id BIGINT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE repo (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  url TEXT NOT NULL,
  default_branch VARCHAR(128) NOT NULL DEFAULT 'main',
  visibility VARCHAR(32) NOT NULL DEFAULT 'private',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE pipeline (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  type VARCHAR(16) NOT NULL,
  name VARCHAR(128) NOT NULL,
  definition_ref TEXT NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  last_run_id BIGINT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE build (
  id BIGSERIAL PRIMARY KEY,
  pipeline_id BIGINT NOT NULL REFERENCES pipeline(id) ON DELETE CASCADE,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  commit_sha VARCHAR(64) NOT NULL,
  branch VARCHAR(128) NOT NULL,
  triggered_by BIGINT NOT NULL DEFAULT 0,
  artifact_id TEXT NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  log_ref TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NULL,
  finished_at TIMESTAMPTZ NULL
);

CREATE TABLE environment (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  name VARCHAR(128) NOT NULL,
  type VARCHAR(16) NOT NULL,
  version VARCHAR(64) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  domain_id BIGINT NULL,
  log_source_id BIGINT NULL,
  metric_panel_id BIGINT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(service_id, name)
);

CREATE TABLE domain (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  environment_id BIGINT NOT NULL REFERENCES environment(id) ON DELETE CASCADE,
  domain_name VARCHAR(255) NOT NULL UNIQUE,
  is_wildcard BOOLEAN NOT NULL DEFAULT FALSE,
  protocol VARCHAR(16) NOT NULL DEFAULT 'https',
  certificate_id BIGINT NULL,
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE certificate (
  id BIGSERIAL PRIMARY KEY,
  domain_id BIGINT NOT NULL REFERENCES domain(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  issued_at TIMESTAMPTZ NULL,
  expires_at TIMESTAMPTZ NULL,
  renewal_policy VARCHAR(32) NOT NULL DEFAULT 'auto',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE log_source (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  environment_id BIGINT NOT NULL REFERENCES environment(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  project VARCHAR(128) NOT NULL DEFAULT '',
  logstore VARCHAR(128) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE metric_panel (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  environment_id BIGINT NOT NULL REFERENCES environment(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  dashboard_url TEXT NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE alert_rule (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  environment_id BIGINT NOT NULL REFERENCES environment(id) ON DELETE CASCADE,
  name VARCHAR(128) NOT NULL,
  metric VARCHAR(64) NOT NULL,
  threshold NUMERIC(18,6) NOT NULL,
  severity VARCHAR(16) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE release (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  environment_id BIGINT NOT NULL REFERENCES environment(id) ON DELETE CASCADE,
  version VARCHAR(64) NOT NULL,
  commit_sha VARCHAR(64) NOT NULL,
  image_tag VARCHAR(255) NOT NULL DEFAULT '',
  strategy VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  triggered_by BIGINT NOT NULL DEFAULT 0,
  approved_by BIGINT NULL,
  started_at TIMESTAMPTZ NULL,
  finished_at TIMESTAMPTZ NULL,
  summary TEXT NOT NULL DEFAULT '',
  risk_level VARCHAR(16) NOT NULL DEFAULT 'unknown'
);

CREATE TABLE approval (
  id BIGSERIAL PRIMARY KEY,
  release_id BIGINT NOT NULL REFERENCES release(id) ON DELETE CASCADE,
  status VARCHAR(32) NOT NULL,
  approver_id BIGINT NULL,
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rollback_record (
  id BIGSERIAL PRIMARY KEY,
  release_id BIGINT NOT NULL REFERENCES release(id) ON DELETE CASCADE,
  target_version VARCHAR(64) NOT NULL,
  triggered_by BIGINT NOT NULL DEFAULT 0,
  reason TEXT NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  finished_at TIMESTAMPTZ NULL
);

CREATE TABLE incident (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  release_id BIGINT NULL REFERENCES release(id) ON DELETE SET NULL,
  severity VARCHAR(16) NOT NULL,
  title VARCHAR(255) NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL,
  log_ref TEXT NOT NULL DEFAULT '',
  metric_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE release_event (
  id BIGSERIAL PRIMARY KEY,
  release_id BIGINT NOT NULL REFERENCES release(id) ON DELETE CASCADE,
  event_type VARCHAR(64) NOT NULL,
  status_before VARCHAR(32) NOT NULL DEFAULT '',
  status_after VARCHAR(32) NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_by BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_log (
  id BIGSERIAL PRIMARY KEY,
  actor_id BIGINT NOT NULL DEFAULT 0,
  actor_role VARCHAR(32) NOT NULL,
  action VARCHAR(64) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id BIGINT NOT NULL DEFAULT 0,
  request_id VARCHAR(128) NOT NULL,
  detail JSONB NOT NULL DEFAULT '{}'::jsonb,
  result VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ai_run (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NULL REFERENCES service(id) ON DELETE SET NULL,
  release_id BIGINT NULL REFERENCES release(id) ON DELETE SET NULL,
  run_type VARCHAR(32) NOT NULL,
  input_ref TEXT NOT NULL DEFAULT '',
  output_ref TEXT NOT NULL DEFAULT '',
  risk_level VARCHAR(16) NOT NULL DEFAULT 'unknown',
  status VARCHAR(32) NOT NULL,
  created_by BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_repo_service_id ON repo(service_id);
CREATE INDEX idx_pipeline_service_id ON pipeline(service_id);
CREATE INDEX idx_build_service_id ON build(service_id);
CREATE INDEX idx_env_service_id ON environment(service_id);
CREATE INDEX idx_domain_service_id ON domain(service_id);
CREATE INDEX idx_certificate_domain_id ON certificate(domain_id);
CREATE INDEX idx_log_source_service_id ON log_source(service_id);
CREATE INDEX idx_metric_panel_service_id ON metric_panel(service_id);
CREATE INDEX idx_alert_rule_service_id ON alert_rule(service_id);
CREATE INDEX idx_release_service_id ON release(service_id);
CREATE INDEX idx_approval_release_id ON approval(release_id);
CREATE INDEX idx_rollback_release_id ON rollback_record(release_id);
CREATE INDEX idx_incident_service_id ON incident(service_id);
CREATE INDEX idx_release_event_release_id ON release_event(release_id);
CREATE INDEX idx_audit_created_at ON audit_log(created_at);
CREATE INDEX idx_ai_run_service_id ON ai_run(service_id);

