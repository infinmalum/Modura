# Phase 1 Authentication Contract

## Included flows

Phase 1 supports:

- login with tenant-scoped username or verified email plus password;
- short-lived access tokens;
- opaque rotating refresh tokens backed by server-side sessions;
- logout current session and logout all sessions;
- password change/reset and account disable;
- administrator unlock after abuse controls trigger.

Tenant discovery/selection happens before credential verification. Responses must not disclose whether a tenant, username, or email exists.

## Passwords

Passwords are hashed with Argon2id using a well-reviewed library and a self-describing encoded format containing algorithm version, parameters, salt, and hash. Parameters must be benchmarked for the deployment class and documented in configuration; use current OWASP-aligned values at implementation time rather than freezing guessed constants here.

Each password uses a cryptographically random salt. Plaintext exists only for verification/setup processing, is never logged or persisted, and is transmitted only over TLS. Successful login may transparently rehash when stored parameters are obsolete.

Password policy should prioritize length and breached/common-password rejection over composition rules. Reset uses a single-use, short-lived server-side token or an invitation flow; shared default passwords are forbidden.

Token issuance and consumption are implemented, but Modura currently has no
configured mail client. Public recovery/invitation request workflows MUST remain
unavailable until a delivery adapter is selected; they MUST NOT expose raw
tokens in HTTP responses as a development substitute. Stage 2 provisioning and
the later notification/admin workflow own that integration.

## Access tokens

Access tokens are short-lived signed tokens containing only stable identifiers and minimum authorization/session context: issuer, audience, subject/user ID, tenant ID, session ID, issued/expiry times, and token ID. Sensitive profile data and complete mutable permission lists do not belong in the token.

The API validates signature, algorithm, issuer, audience, expiry, tenant/session binding, and account/session security state as required. Key rotation and clock skew are explicit configuration. Authorization still occurs for every protected operation.

The browser keeps the access token in memory, not local storage. TLS is mandatory outside local development.

## Refresh sessions

The refresh token is an opaque, high-entropy secret delivered in a `Secure`, `HttpOnly`, appropriately scoped `SameSite` cookie. Only a hash of the secret is stored server-side. The session stores user, tenant, token family/session identifiers, expiry, creation/last-use data, revocation state/reason, and minimal client metadata needed for security review.

Every successful refresh atomically consumes the presented token and issues a new token in the same family. Reuse of a consumed token revokes the entire family and emits a security audit event. Concurrent refresh behavior must be deterministic and tested.

Cookie-authenticated refresh/logout endpoints require an explicit CSRF defense compatible with the chosen SameSite policy. Refresh tokens never appear in URLs or application logs.

## Invalidation

The current refresh session is revoked on logout. All sessions are revoked on logout-all, account disable, tenant suspension where applicable, password reset, confirmed credential compromise, or administrative security revocation. Password change revokes other sessions and rotates the current session unless product policy explicitly requires all-session logout.

Role/policy changes must become effective within a documented bounded interval. Prefer server-side authorization lookup/versioning rather than long-lived permission claims.

## Abuse controls and audit

Authentication applies configurable throttling to tenant/account and network/device signals. Successful and failed attempts, lock/unlock, refresh reuse, password/reset operations, and session revocation produce privacy-conscious security events. Responses remain generic, while internal events contain correlation data without passwords or tokens.

## Deferred

The following are not Phase 1 requirements: MFA, SAML, enterprise SSO, passkeys, service accounts, social login, guest registration, authorization-code grants, and a full OAuth authorization server/client registry. Adding one requires threat modeling and an ADR where it changes the authentication boundary.

## Required tests

Tests must cover successful and failed login without account enumeration, tenant isolation, password rehash, access-token validation, refresh rotation, refresh replay/family revocation, logout, all invalidation triggers, CSRF behavior, expiry/clock boundaries, and redaction of secrets.
