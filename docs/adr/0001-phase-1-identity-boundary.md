# ADR 0001: Phase 1 tenant discovery and identity boundary

Status: accepted
Date: 2026-08-29

## Context

Authentication needs a deterministic tenant realm before credentials are checked, while a client-provided tenant value cannot itself establish authority. Phase 1 also needs tenant-local usernames and email login without introducing a global identity service.

## Decision

- Login accepts a tenant slug as a realm selector. The server resolves it to an active tenant before credential verification and returns the same generic authentication failure for unknown tenants and invalid credentials.
- Authenticated tenant scope comes only from a validated access token and its server-side session/account state. Tenant headers, paths, and request bodies never override it.
- A user belongs to exactly one tenant in Phase 1. Usernames and verified email addresses are normalized with Unicode-aware lower-casing and surrounding-space removal, and are unique within a tenant. The same values may exist in different tenants.
- Email cannot be used to log in until `email_verified_at` is set. Email verification delivery is outside this slice; invitation/reset tokens may establish verification when their consuming workflow explicitly does so.
- Tenant lifecycle states are `provisioning`, `active`, and `suspended`. Deletion remains a later explicit retention workflow; it is not represented as ordinary CRUD.
- Access tokens use HMAC-SHA-256 in Phase 1, with issuer, audience, key identifier, and short lifetime configured by the process. Key rotation accepts a bounded set of verification keys. Moving to asymmetric signing is a security-boundary change requiring an amendment.
- A user has one primary department and zero or one position in Phase 1. Their schema is introduced with organization work in Stage 2.

## Consequences

Tenant slugs are discovery hints only. Every protected operation still receives typed verified actor and tenant values. Local identities remain simple and tenant-isolated, while future cross-tenant human identity linking would require a separate model and migration.
