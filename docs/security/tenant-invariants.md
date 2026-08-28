# Tenant Invariants

## Model

`TenantID` is a distinct UUIDv7-based domain type, not a free-form string. The platform/global context is represented explicitly; it is never encoded as an empty, zero, default, or magic tenant ID.

Every table must be classified in schema documentation as either:

- tenant-owned: every row belongs to exactly one tenant and has a non-null `tenant_id`; or
- global: rows are owned by the platform and have no implied tenant ownership.

A table must not sometimes use a null tenant to mean global and sometimes use a value to mean tenant-owned unless an accepted ADR defines override semantics and constraints.

## Tenant resolution

For an authenticated request, tenant scope is derived from the authenticated server-side session/token claims and revalidated against active user membership/session state when required. A hostname or tenant selector may help choose a login realm. A client header, URL field, query parameter, or request body value is only a requested target and is never trusted by itself.

Middleware may place the verified actor and tenant scope in `context.Context`. Application APIs must still make tenant-sensitive behavior visible through typed actor/scope inputs. Repositories for tenant-owned data require `TenantID` or a typed tenant scope as a parameter. A global/unscoped repository method is forbidden outside narrowly reviewed platform administration and migration code.

## Database invariants

- Tenant-owned rows have `tenant_id uuid NOT NULL` with a valid tenant reference where lifecycle/partitioning permits it.
- Natural/business uniqueness includes `tenant_id` by default, such as `(tenant_id, normalized_username)`.
- Foreign-key or equivalent application relationships between tenant-owned records must guarantee same-tenant ownership. Use composite constraints where appropriate or verify inside the transaction when PostgreSQL cannot express it cleanly.
- Queries name tenant predicates explicitly. Global ORM/interceptor injection is not the correctness mechanism.
- Inserts derive tenant from verified scope. Updates/deletes match both primary key and tenant. “Load by ID then trust it” is insufficient unless same-tenant ownership was verified.

## Actor boundaries

A tenant administrator can act only within its authenticated tenant and cannot assign platform roles or grant authority it does not possess.

A platform administrator is a distinct platform principal/capability, not a specially named tenant role. Cross-tenant operations require:

- explicit platform authorization;
- an explicit target tenant;
- a business/operational reason where the operation is sensitive;
- an audit event containing platform actor, target tenant, action, resource, outcome, and correlation ID.

Impersonation, if ever added, requires a separate ADR and must preserve both real actor and effective actor.

## Background work

Tenant jobs carry an immutable envelope containing job ID, verified tenant ID, initiating actor or system principal, correlation ID, and payload version. Workers validate that the tenant is in an allowed lifecycle state before doing work. They never infer tenant from payload-owned records or process a tenant-owned job in a global context.

Platform jobs must be explicitly classified as global. Fan-out across tenants produces independently observable tenant-scoped work items.

## Audit and lifecycle

Every important tenant-owned write records the tenant. Tenant creation, suspension, reactivation, deletion scheduling, and platform cross-tenant access are always audited.

Suspension blocks new interactive sessions and tenant business writes according to documented policy. Deletion is not ordinary CRUD: retention, export, dependency handling, session revocation, and final purge require an explicit workflow before production.

## Required tests

Integration tests must prove at least:

- identical business keys can exist in different tenants but not twice in one tenant;
- list/detail/update/delete cannot observe or mutate another tenant's records, including guessed IDs;
- cross-tenant role, department, position, parent, and policy assignments fail atomically;
- tenant values supplied by clients cannot override authenticated scope;
- tenant jobs retain scope across serialization and execution;
- platform operations require the dedicated capability and emit audit evidence;
- suspended/deleted tenant behavior matches lifecycle policy.

These tests use real PostgreSQL constraints and queries, not repository mocks.
