# Authorization and Data-Scope Contract

## Stable permissions

Backend resources and actions are canonical identifiers independent of menus,
routes, and buttons. OpenAPI operations declare their required identifier and
CI verifies that it exists in the backend registry. UI capabilities are a
projection of the same registry and never authorize a request.

## Policy model

Tenant users receive roles, and roles receive allow policies consisting of a
resource, action, and typed data scope. Phase 1 has no explicit deny and no role
inheritance. A request is functionally authorized when at least one assigned
role has the requested allow policy in the actor's verified tenant. Casbin
evaluates this tenant-domain RBAC relationship; PostgreSQL remains the
authoritative policy store and application services own all writes.

The reserved `tenant-admin` role receives the complete tenant-management
policy set during provisioning. Reserved-role policies cannot be edited or
removed through delegated tenant APIs.

## Multiple roles and scopes

Allow policies combine by union. This is monotonic: assigning another role can
add authority but cannot subtract authority granted by another role. For a
permitted resource/action pair, scopes combine from broadest to narrowest:

1. `all` makes all other scopes irrelevant;
2. `department-and-descendants` contributes the actor's primary department and
   its descendants;
3. `department` contributes only the actor's primary department;
4. `custom` contributes explicitly stored department IDs;
5. `self` contributes only records owned by or representing the actor.

The result is a typed value passed to the owning repository. It never contains
SQL or an unchecked predicate. A scope that needs the actor's department but
the actor has no primary department contributes nothing. Custom departments
must belong to the same tenant and are stored as foreign-keyed UUIDs.

Department and user-organization policies support all five scope modes.
Tenant-wide catalogs and authorization-administration resources support only
`all`, because they do not have meaningful row ownership or department
membership; narrower values are rejected instead of being treated as all.

## Change propagation and concurrency

Authorization is loaded server-side for each protected request, so committed
role and policy changes affect the next request without waiting for access-token
expiry. Desired-state role grants use an expected version; a stale version is
rejected rather than silently overwriting a concurrent administrator's change.
The grant replacement and its before/after audit evidence commit atomically.
