# Settings and Audit Access

## Settings boundary

Dictionaries provide stable codes and administrator-controlled display values
for forms and projections. Configuration stores only explicitly declared,
non-secret presentation or operational values. Neither facility is a generic
database for domain aggregates, authorization, credentials, endpoints, or
code-executing expressions.

Tenant identity always comes from the verified actor. Effective dictionary
reads select the tenant-owned type as a whole when its code exists; otherwise
they select the global type. Effective configuration reads select an allowed
tenant override per key and otherwise the global default.

All writes validate normalized keys/codes, value types, length limits, unique
item codes, and optimistic versions. Important changes and their before/after
states are audited in the same PostgreSQL transaction.

## Audit query boundary

Tenant audit queries require the stable `audit.events/read` permission and are
always constrained to the actor tenant. Platform audit queries require the
distinct platform principal and an explicit target tenant when filtering tenant
events. Results are newest-first, cursor/page bounded, and never expose
credentials, tokens, session identifiers, password material, or raw secret
configuration.

Audit before/after JSON is produced only by trusted application services. The
query projection applies a defense-in-depth recursive key redaction policy for
names containing password, secret, token, credential, authorization, cookie,
or session material before returning data. Audit records are immutable through
the application. Phase 1 documents retention requirements but does not add an
automatic purge job; production retention and backup policy are a Stage 5
release gate.

