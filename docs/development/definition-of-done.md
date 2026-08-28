# Phase 1 Definition of Done

Phase 1 is complete only when all acceptance statements below are demonstrably true. A feature being present in source code is not sufficient.

## Product outcomes

### Authentication and self-service

- A user can select/discover a tenant, log in with tenant-scoped username or verified email and password, refresh the session, log out, change password, and view/update safe profile fields.
- Disabled users and suspended tenants cannot establish or continue sessions according to policy.
- Refresh replay and all documented invalidation triggers are tested.

### Tenant administration

- A platform actor can provision, view, update, suspend, and reactivate a tenant.
- Provisioning atomically creates the minimum root organization, tenant-administrator role, and secure first-administrator invitation/setup path; retry cannot create duplicates.
- A tenant administrator can manage only users, departments, positions, roles, and policies in its own tenant and cannot grant authority it does not possess.

### Organization

- Department create/read/update and tree navigation work with cycle and cross-tenant prevention.
- The chosen Phase 1 user-department and user-position cardinalities are documented and enforced by schema and APIs.
- Destructive/move operations reject invariant violations with stable errors.

### Authorization and data scope

- Stable resources/actions exist independently of menus.
- Users can be assigned tenant roles; roles can receive functional policies and typed data scopes.
- HTTP operations enforce authorization server-side, while the React shell derives routes/actions from the same permission catalog.
- All, self, department, department-and-descendants, and explicitly designed custom scopes behave according to documented combination rules, or unsupported modes are removed from Phase 1 rather than stubbed.

### System and audit

- Global/tenant ownership is explicit for every dictionary and configuration record.
- Important writes and platform cross-tenant operations create queryable, redacted audit records with actor, tenant, action, resource, resource ID, time, outcome, and trace/request correlation.

## Contract and engineering outcomes

- `api/openapi.yaml` and its allowed `$ref` files describe every Phase 1 HTTP operation.
- Go behavior passes OpenAPI contract tests.
- TypeScript client/hooks regenerate reproducibly with pinned tools and no diff.
- Database migrations apply successfully to an empty supported PostgreSQL instance and follow documented recovery procedure.
- Backend formatting/static analysis, unit tests, PostgreSQL integration tests, authorization tests, tenant-isolation tests, and contract tests pass.
- Frontend formatting, lint, typecheck, production build, critical component tests, and minimal login/admin critical-path E2E tests pass.
- Module dependency and table-ownership checks pass.
- Minimum production baseline in `AGENTS.md` is implemented and documented.
- No production build/import/package input references `_reference/**` or `SpringBlade-boot/**`.

## One-command verification

The repository must provide a root-level, non-interactive `make verify` target that runs all deterministic checks required for a change. It must fail fast enough to be useful locally while CI may additionally run slower suites in separate jobs. Dependencies and build caches must use normal persistent user caches or workspace-local caches, never system temporary directories.

At minimum, `make verify` must cover formatting checks, lint/static analysis, tests, OpenAPI validation, generation cleanliness, migration validation, architecture checks, frontend typecheck, and builds. Any required external service must have a documented reproducible local setup.

## Evidence and exclusions

Every Phase 1 acceptance item must link to an automated test, generated artifact check, or operational verification procedure. Any exception is visible and approved according to `AGENTS.md`; “to be implemented later” does not satisfy Done.

Deferred capabilities listed in `AGENTS.md`, the Constitution, and the Phase 1 plan are excluded unless explicitly promoted through a scoped decision.
