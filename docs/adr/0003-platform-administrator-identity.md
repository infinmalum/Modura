# ADR 0003: Distinct platform-administrator identity

Status: accepted
Date: 2026-08-29

## Context

Platform tenant administration crosses tenant boundaries. A tenant role cannot
represent that authority, while API keys and service accounts are explicitly
deferred and are inappropriate for an interactive human administrator.

## Decision

- Platform administrators are global human principals stored separately from
  tenant users. They never have a tenant ID and cannot receive tenant roles.
- Platform login uses a separate endpoint, access-token audience, server-side
  refresh-session table, refresh cookie name, and CSRF cookie name.
- Password hashing, rotation/replay rules, generic authentication failures, and
  security-version invalidation match tenant authentication.
- Platform administrator creation is not public self-registration. An explicit
  local bootstrap/operator command will create the first administrator and must
  reject operation once one exists unless a separately authorized workflow is
  used.
- Cross-tenant operations require the platform actor, explicit target tenant,
  reason, and audit evidence. They are never inferred from a tenant selector.

## Consequences

Platform and tenant sessions cannot be exchanged. Platform endpoints require a
dedicated authorization boundary and do not accept tenant access tokens. Audit
storage and the bootstrap command must be present before platform tenant APIs
are exposed.
