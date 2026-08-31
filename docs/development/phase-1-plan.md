# Modura Phase 1 Plan

Status: in progress
Planning basis: Project Constitution, executable agent contract, and SpringBlade business research

## Implementation status

- Stage 0: complete on 2026-08-28. The backend/admin split, initial OpenAPI contract, generated Gin and React Query bindings, sqlc/migration structure, configuration and graceful HTTP server, workspace-local build caches, locked Admin dependencies, and `make verify` are operational.
- Stage 1: core complete on 2026-08-29, with email delivery explicitly pending.
  Tenant/login boundary ADR, tenant-local identity schema,
  Argon2id credentials, UUIDv7 identifiers, signed access-token validation,
  opaque refresh secrets, atomic PostgreSQL rotation/replay-family revocation,
  login/refresh/logout/logout-all/password-change HTTP delivery, explicit
  process configuration and PostgreSQL wiring, dependency-aware readiness,
  CSRF-protected cookies, and authentication unit/contract tests are
  implemented. Single-use invitation/password-reset issuance and consumption,
  expiry enforcement, credential establishment/replacement, and all-session
  invalidation are implemented; trusted delivery is intentionally left to the
  provisioning/notification workflow rather than exposing secrets over HTTP.
  Tenant-scoped account disable, security-version invalidation, session
  revocation, and locked-account-only unlock operations are implemented as
  application APIs. Their management HTTP exposure waits for Stage 2/3
  administrator provisioning and authorization. Real-PostgreSQL
  tenant-isolation, migration, invitation activation, account disable,
  refresh-rotation, and replay-family-revocation tests passed against the
  dedicated `modura_test` environment on 2026-08-29.
- Email delivery is unavailable in the current environment. Invitation and
  password-reset secrets can be issued and consumed securely, but no user-facing
  request flow is complete until a notification/mail client is selected and
  configured. Tokens MUST NOT be returned from public request endpoints as a
  substitute. This delivery gap is tracked for the Stage 2 provisioning and
  Stage 4 notification/admin workflow and is not represented as working UI.
- Stage 2: in progress. Organization cardinality and provisioning idempotency
  decisions are accepted. Department, position, and single-primary-department
  schema/application foundations are implemented with real-PostgreSQL tests
  for root, sibling uniqueness, cycle, delete, assignment, and cross-tenant
  invariants. Atomic tenant provisioning now creates and activates the tenant,
  root department, invited administrator, primary assignment, reserved
  tenant-administrator role/grant, and invitation token in one transaction.
  Real-PostgreSQL tests cover same-request retry, conflicting key reuse, and
  rollback without partial tenants. A stable resource/action registry and
  server-side reserved tenant-admin check now protect department list/create/
  move/delete HTTP APIs; requests derive tenant scope only from the validated
  actor and writes also require CSRF. Position list/create and desired-state
  user primary-department/optional-position assignment APIs use the same
  boundary, with real-PostgreSQL list and cross-tenant assignment tests. A
  distinct global human platform-administrator model now has one-time
  bootstrap, Argon2id login, separate token audience/principal type,
  server-side refresh rotation/replay revocation, and explicit rejection of
  tenant access tokens. The local, stdin-secret bootstrap command and separate
  platform login/refresh HTTP boundary are implemented with distinct cookies.
  Platform tenant listing plus suspend/reactivate endpoints now require a
  verified platform principal, platform CSRF cookies, and an explicit reason;
  lifecycle state and durable audit evidence commit in one PostgreSQL
  transaction with request correlation. Tenant provisioning now also requires
  a verified platform actor, explicit reason, and request correlation, and its
  successful audit evidence commits in the same idempotent transaction.
  Platform provisioning HTTP delivery and audit coverage for the remaining
  organization writes remain.
- Stages 3–5: pending.

## Goal

Deliver a production-oriented modular-monolith foundation in which a platform actor can provision a tenant, the tenant administrator can manage identity/organization/authorization, and all operations are contract-driven, tenant-isolated, auditable, testable, and usable through a minimal React admin shell.

This plan orders outcomes and decision gates. It is not authorization to expand the Phase 1 scope.

## Stage 0: Contract and project skeleton

Deliverables:

- accept or amend the defaults in `AGENTS.md` through initial ADRs;
- settle the open Phase 1 cardinality and policy questions listed below;
- create Go module and composition-root skeleton, React/Vite shell, `api/openapi.yaml`, migration/sqlc structure, and pinned tooling;
- establish PostgreSQL integration-test environment and workspace/global cache policy;
- implement `make verify` and CI gates for formatting, lint, tests, contract validation, generation cleanliness, migrations, architecture, and builds;
- establish configuration, secret, error, identifier, clock, logging, and telemetry conventions.

Exit criteria: an empty vertical-slice skeleton builds; an empty database migrates; generated artifacts reproduce; one command verifies the repository; no business feature is falsely represented as complete.

## Stage 1: Tenant and local identity foundation

Deliverables:

- tenant aggregate and lifecycle states needed for Phase 1;
- explicit tenant resolution and typed verified actor/scope propagation;
- local user/credential model with tenant-scoped normalized login identifiers;
- Argon2id password handling and secure invitation/reset primitives;
- access-token validation and rotating server-side refresh sessions;
- login, refresh, logout, logout-all, password change/reset, disable and session invalidation;
- tenant isolation and authentication security tests from the start.

Exit criteria: two tenants can use the same username without leakage; all refresh/session attacks in the authentication contract are covered; no endpoint trusts a client tenant selector by itself.

## Stage 2: Tenant provisioning and organization

Deliverables:

- idempotent, atomic tenant provisioning workflow;
- root department, tenant-administrator role, and first-administrator invitation/setup;
- department tree with cycle, move, delete, and cross-tenant invariants;
- Phase 1 position catalog and user membership model;
- platform tenant administration and tenant-administrator organization/user management APIs;
- audit events for lifecycle and management writes.

Exit criteria: provisioning is retry-safe; partial tenants cannot become active; organization relationships are enforced by PostgreSQL and integration tests.

## Stage 3: Authorization and data scope

Deliverables:

- stable resource/action registry owned independently of UI routes;
- tenant roles, user-role assignments, reserved platform capabilities, and delegated-administration checks;
- policy evaluation through Casbin or an ADR-approved equivalent;
- typed data-scope resolution and repository integration;
- explicit combination behavior for users with multiple roles/scopes;
- OpenAPI-to-permission mapping and negative authorization tests;
- desired-state role grant API with lost-update protection and before/after audit.

Exit criteria: server-side authorization and tenant checks cover every protected operation; UI cannot expand authority; all supported data scopes pass real-PostgreSQL tests.

## Stage 4: System configuration, audit, and admin shell

Deliverables:

- dictionary and non-secret configuration with explicit global/tenant ownership, validation, and audit;
- durable audit query surface with redaction and retention/access policy;
- React admin shell for login, navigation, users, departments, positions, roles, policies, tenant administration where authorized, dictionary/configuration, and audit review;
- generated client/hooks as the only routine API access layer;
- permission-derived routes/actions and centralized authentication/error behavior;
- critical component and minimal E2E coverage.

Exit criteria: all required Phase 1 operator workflows can be completed through the UI and pass automated critical-path tests.

## Stage 5: Production hardening and release gate

Deliverables:

- graceful lifecycle, health probes, timeouts, body limits, CORS/CSRF, configuration/secrets, redaction, and observability baseline;
- migration recovery procedure, backup/restore requirements, dependency/license/vulnerability checks;
- concurrency, pagination, failure-path, and security review;
- clean-room confirmation that removed reference source is not a dependency and provenance records are sufficient;
- final evidence mapping against every Definition of Done item.

Exit criteria: `make verify` and CI pass from a clean checkout; an empty database can reach the current schema; release/deployment and recovery procedures are exercised; no Phase 1 acceptance item lacks evidence.

## Decisions required before relevant implementation

Resolve these with concise ADRs or authoritative security/domain documentation:

1. Tenant discovery: hostname, explicit selector, invitation context, or supported combination.
2. Email login: verification flow, normalization, and tenant-scoped uniqueness.
3. User membership: one primary department versus multiple memberships; position cardinality and department scope.
4. Role hierarchy: omit by default unless a concrete inheritance workflow is accepted.
5. Permission registry ownership and generation/check relationship with OpenAPI and frontend.
6. Multiple-role/data-scope combination semantics and whether explicit deny exists.
7. Dictionary/configuration ownership and tenant override semantics.
8. Tenant suspension/deletion/retention/restoration rules.
9. First-administrator invitation and provisioning idempotency key.
10. Audit delivery/transactional reliability and retention boundaries.
11. Access-token signing/key rotation and bounded authorization-change propagation.
12. Exact supported custom data-scope representation; arbitrary SQL is excluded.

## Explicitly deferred

- microservice extraction and broad gRPC surface;
- MFA, SAML, enterprise SSO, passkeys, social login, service accounts, and OAuth authorization server/client registry;
- top-menu workspaces, notices, region datasets, report designer, datasource management, and CRUD generator;
- Redis unless a measured requirement justifies it;
- BPM, complex reporting, IoT, big data, messaging middleware, sharding, distributed transactions, Kubernetes operators, AI gateway, and RAG platform;
- multi-region/high-scale operational automation beyond the minimum production baseline.

Deferred items must not receive placeholder APIs, tables, dependencies, or abstractions.
