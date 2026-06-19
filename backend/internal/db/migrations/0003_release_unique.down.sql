-- Remove release environment+version uniqueness constraint
ALTER TABLE "release" DROP CONSTRAINT IF EXISTS uq_release_env_version;
