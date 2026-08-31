# ADR 0004: Settings ownership and override semantics

Status: accepted
Date: 2026-08-31

## Context

Phase 1 needs reusable dictionaries and non-secret configuration at both
platform and tenant scope. Tenant invariants prohibit treating a null or magic
tenant ID as an implicit global owner. Reads also need deterministic behavior
when a tenant customizes a platform default.

## Decision

- The `settings` module owns dictionaries and non-secret configuration.
- Global and tenant-owned rows use separate tables. Global tables have no
  `tenant_id`; tenant tables require a non-null, foreign-keyed `tenant_id`.
- A tenant dictionary type with the same stable code replaces the complete
  global dictionary type for that tenant. Dictionary items never merge across
  owners, avoiding ambiguous labels, ordering, and disabled-state inheritance.
- Configuration keys are declared globally with a value type and an explicit
  `tenant_overridable` flag. Global and tenant values are stored separately.
  An eligible tenant value wins per key; otherwise the global value applies.
- Supported configuration value types are `string`, `boolean`, `integer`, and
  `json`. Application validation must match the declared type before storage.
- Settings are presentation and non-security business configuration only.
  Credentials, connection strings, signing material, authorization policy,
  tenant identity, and core domain invariants are forbidden.
- Platform actors manage global definitions/defaults. Tenant actors manage only
  their tenant dictionaries and eligible tenant overrides. Both boundaries
  require explicit permissions/capabilities and transactional audit evidence.
- Desired-state writes use optimistic versions. Effective reads never accept a
  client-supplied tenant as authority.

## Consequences

The schema contains parallel global and tenant dictionary tables, plus global
configuration definitions/defaults and tenant override values. Queries are
slightly more explicit but ownership is mechanically visible. A future need
for partial dictionary-item inheritance or secret configuration requires a new
decision and migration rather than changing the meaning of existing rows.

## Alternatives considered

- Nullable `tenant_id` in shared tables was rejected because ownership and
  uniqueness become implicit and conflict with tenant invariants.
- Magic platform tenant IDs were rejected because they create a fake tenant and
  permit accidental tenant-scoped access.
- Per-item dictionary merging was rejected because deletion, ordering, and
  disabled-state semantics are difficult to explain and audit.
- A generic secret/configuration vault was rejected as outside Phase 1 and an
  inappropriate responsibility for application tables.

## Contract impact

This ADR defines the global/tenant ownership and override semantics required by
the Phase 1 plan decision gate. It does not create an exception to
`AGENTS.md`; it makes the permitted explicit model concrete. Settings APIs,
migrations, authorization identifiers, audit events, and integration tests
must be delivered together.

