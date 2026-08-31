-- owner: authorization
DELETE FROM modura.role_policies
WHERE resource IN ('settings.dictionaries', 'settings.configurations', 'audit.events');
ALTER TABLE modura.role_policies DROP CONSTRAINT role_policies_resource_valid;
ALTER TABLE modura.role_policies ADD CONSTRAINT role_policies_resource_valid CHECK (resource IN (
    'organization.departments', 'organization.positions',
    'organization.user-organization', 'authorization.roles',
    'authorization.policies', 'authorization.user-roles'
));

-- owner: settings
DROP TABLE IF EXISTS modura.tenant_configuration_values;
DROP TABLE IF EXISTS modura.global_configuration_values;
DROP TABLE IF EXISTS modura.configuration_definitions;
DROP TABLE IF EXISTS modura.tenant_dictionary_items;
DROP TABLE IF EXISTS modura.tenant_dictionary_types;
DROP TABLE IF EXISTS modura.global_dictionary_items;
DROP TABLE IF EXISTS modura.global_dictionary_types;
