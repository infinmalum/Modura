-- owner: organization
CREATE TABLE modura.departments (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES modura.tenants (id),
    parent_id uuid,
    name text NOT NULL,
    normalized_name text NOT NULL,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT departments_tenant_id_id_unique UNIQUE (tenant_id, id),
    CONSTRAINT departments_parent_fk FOREIGN KEY (tenant_id, parent_id)
        REFERENCES modura.departments (tenant_id, id),
    CONSTRAINT departments_not_self_parent CHECK (parent_id IS NULL OR parent_id <> id),
    CONSTRAINT departments_name_present CHECK (btrim(name) <> ''),
    CONSTRAINT departments_normalized_name CHECK (normalized_name = lower(btrim(normalized_name)) AND normalized_name <> ''),
    CONSTRAINT departments_sibling_name_unique UNIQUE NULLS NOT DISTINCT (tenant_id, parent_id, normalized_name)
);
CREATE UNIQUE INDEX departments_single_root_idx ON modura.departments (tenant_id)
    WHERE parent_id IS NULL;

CREATE TABLE modura.positions (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES modura.tenants (id),
    name text NOT NULL,
    normalized_name text NOT NULL,
    status text NOT NULL CHECK (status IN ('active', 'disabled')),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT positions_tenant_id_id_unique UNIQUE (tenant_id, id),
    CONSTRAINT positions_name_present CHECK (btrim(name) <> ''),
    CONSTRAINT positions_normalized_name CHECK (normalized_name = lower(btrim(normalized_name)) AND normalized_name <> ''),
    CONSTRAINT positions_name_unique UNIQUE (tenant_id, normalized_name)
);

CREATE TABLE modura.user_organization (
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    primary_department_id uuid NOT NULL,
    position_id uuid,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, user_id),
    CONSTRAINT user_organization_user_fk FOREIGN KEY (tenant_id, user_id)
        REFERENCES modura.users (tenant_id, id),
    CONSTRAINT user_organization_department_fk FOREIGN KEY (tenant_id, primary_department_id)
        REFERENCES modura.departments (tenant_id, id),
    CONSTRAINT user_organization_position_fk FOREIGN KEY (tenant_id, position_id)
        REFERENCES modura.positions (tenant_id, id)
);
