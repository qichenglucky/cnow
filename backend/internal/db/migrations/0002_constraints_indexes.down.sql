-- Reverse 0002

DROP TABLE IF EXISTS idempotency_key;

DROP INDEX IF EXISTS idx_approval_release;
DROP INDEX IF EXISTS idx_incident_service;
DROP INDEX IF EXISTS idx_env_service;
DROP INDEX IF EXISTS idx_build_pipeline;
DROP INDEX IF EXISTS idx_ai_run_service;
DROP INDEX IF EXISTS idx_audit_resource;
DROP INDEX IF EXISTS idx_audit_created_at;
DROP INDEX IF EXISTS idx_audit_action_type;
DROP INDEX IF EXISTS idx_release_event_release;
DROP INDEX IF EXISTS idx_release_env_id;
DROP INDEX IF EXISTS idx_release_service_id;
DROP INDEX IF EXISTS idx_release_created_at;
DROP INDEX IF EXISTS idx_release_status;

ALTER TABLE incident DROP CONSTRAINT IF EXISTS chk_incident_severity;
ALTER TABLE build DROP CONSTRAINT IF EXISTS chk_build_status;
ALTER TABLE approval DROP CONSTRAINT IF EXISTS chk_approval_status;
ALTER TABLE "release" DROP CONSTRAINT IF EXISTS chk_release_status;
ALTER TABLE service DROP CONSTRAINT IF EXISTS chk_service_status;
ALTER TABLE "release" DROP CONSTRAINT IF EXISTS uq_release_service_version;
ALTER TABLE environment DROP CONSTRAINT IF EXISTS uq_env_service_name;
ALTER TABLE service DROP CONSTRAINT IF EXISTS uq_service_name;
