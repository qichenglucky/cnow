-- 0004 rollback: Drop pipeline and build tables
-- Note: Only drops if tables were created by this migration (not pre-existing)
DROP TABLE IF EXISTS build;
DROP TABLE IF EXISTS pipeline;
