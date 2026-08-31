-- owner: authorization
DROP TABLE IF EXISTS modura.user_role_versions;
DROP TABLE IF EXISTS modura.role_policy_departments;
DROP TABLE IF EXISTS modura.role_policies;
ALTER TABLE modura.roles DROP CONSTRAINT IF EXISTS roles_version_positive;
ALTER TABLE modura.roles DROP COLUMN IF EXISTS version;
