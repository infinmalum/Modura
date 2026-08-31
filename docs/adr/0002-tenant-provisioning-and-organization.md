# ADR 0002: Tenant provisioning and organization cardinality

Status: accepted
Date: 2026-08-29

## Context

Tenant provisioning spans identity, organization, and authorization ownership.
Retries must not create partial duplicate tenants, and organization membership
must have unambiguous Phase 1 cardinality.

## Decision

- A platform provisioning request carries a caller-generated UUIDv7 idempotency
  key. The persisted request also stores a canonical request digest. Reusing a
  key with a different digest is rejected.
- Provisioning is one PostgreSQL transaction coordinated through public module
  application operations that share an explicit transaction handle. The
  coordinator does not access module-private repositories.
- A tenant begins in `provisioning` and becomes `active` only after its single
  root department, reserved tenant-administrator role, first invited user,
  role assignment, and invitation token have been created.
- Every tenant has exactly one root department. Other departments have exactly
  one parent. Sibling names are unique after normalization.
- A user has exactly one primary department and zero or one position. Both must
  belong to the user's tenant. Multiple department memberships and concurrent
  positions are deferred until a concrete workflow requires them.
- The raw first-administrator invitation secret is returned only to the trusted
  provisioning coordinator. Until email delivery exists, no public endpoint
  claims that the invitation was delivered.

## Consequences

Database constraints enforce tenant ownership and cardinality while move logic
must still reject cycles transactionally. Provisioning cannot be exposed until
platform authentication and authorization are present; internal application
and integration tests may exercise it beforehand.
