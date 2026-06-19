-- 0003: Add release environment+version uniqueness constraint
ALTER TABLE "release" ADD CONSTRAINT uq_release_env_version UNIQUE (environment_id, version);
