# SpringBlade Business Analysis

Status: research snapshot
Source reviewed: `SpringBlade-boot/`
Purpose: preserve business-domain knowledge before the reference source is removed

## 1. Scope and interpretation

This document records business capabilities, domain relationships, invariants, and workflow ideas observed in the bundled SpringBlade reference application. It deliberately does not preserve its Java package structure, controller/service/mapper layering, framework conventions, annotations, or implementation patterns.

The source is evidence of requirements, not a specification for Modura. Some behavior is mature and worth retaining as a product requirement; some is historical compatibility or an unsafe shortcut and is explicitly marked as such below.

No source code has been copied into this document.

## 2. Product model at a glance

SpringBlade is an administration-platform foundation with five main business areas:

1. Identity and authentication: local accounts, password and captcha login, refresh tokens, external identities, profile/password maintenance, account lock and unlock.
2. Tenant and organization: tenants, hierarchical departments, positions, tenant bootstrap, tenant-scoped users and roles.
3. Authorization: hierarchical roles, role-to-menu grants, UI routes/buttons, API scopes, and row/data scopes.
4. System administration: dictionaries, parameters, regions, notices, OAuth-style clients, navigation groupings, and operational logs.
5. Development tooling: external datasource records and CRUD code-generation recipes.

The central business relationship is:

```text
Tenant
  -> Department tree
  -> Position catalog
  -> User <-> Role
  -> tenant-owned business data

Role
  -> functional permissions/resources
  -> data scopes

Functional permissions
  -> projected as routes, menus, and buttons
```

Modura should retain the concepts but should not make menu records the source of authorization truth.

## 3. Identity and authentication

### 3.1 Local identity

A user carries a tenant identity, account/profile fields, status, and associations with roles, departments, and positions. Accounts are unique within a tenant rather than globally. Users can:

- view and update their own safe profile fields;
- change their password after proving the old password;
- be created, edited, disabled/deleted, unlocked, assigned roles, or have their password reset by an administrator;
- be imported and exported in bulk, with names resolved to organization and role identifiers.

Useful invariant: self-service profile updates must derive the user ID from the authenticated session and must ignore privilege-bearing fields such as account, password, roles, and departments.

Useful invariant: bulk import must derive the tenant from the authenticated operator, not from spreadsheet content. Referenced departments, positions, and roles must resolve within that same tenant.

### 3.2 Login flows

The reference supports multiple grant modes behind one token endpoint:

- username/password;
- username/password plus one-time image captcha;
- refresh token;
- social/external identity.

The durable business idea is that credential verification methods may vary while producing the same authenticated principal. Modura Phase 1 only needs username/email plus password and refresh sessions; captcha and social login can remain deferred.

Login failures are counted independently for the tenant/account pair and the client IP, expire after a cooling period, and can be cleared by a successful login or administrator unlock. This is worth retaining as a security requirement, but the exact thresholds and Redis dependency should remain policy/configuration decisions.

Failure responses intentionally avoid distinguishing “unknown account” from “wrong password.”

### 3.3 External identities

An external identity stores provider, provider subject, profile snapshot, tenant, and an optional local user binding. The observed flow is:

1. Complete provider authorization.
2. Find or create the external identity.
3. If already bound, authenticate the local user.
4. If unbound, return a restricted guest-like state.
5. Registration creates a local account and atomically binds it to the external identity.

Important invariant: the tenant used during registration must come from a trusted external-authentication context or another verified server-side source, never directly from anonymous form input.

This capability is deferred for Modura Phase 1, but the separation of external identity from local user is a sound model.

### 3.4 Behaviors not to inherit

- Do not store password digests with a fast unsalted hash; use Argon2id with upgradeable parameters.
- Do not accept encrypted passwords as a substitute for TLS.
- Do not use stateless refresh JWTs without server-side rotation, reuse detection, and revocation.
- Logout must revoke the relevant session; it cannot be a no-op response.
- Account disable, password reset, password change, and privilege-sensitive identity changes must invalidate affected sessions.
- Do not store role, department, or position IDs as comma-separated columns.

## 4. Tenant lifecycle

### 4.1 Tenant as a security boundary

The tenant is both a product/customer record and the primary data-isolation boundary. A platform administrator can manage tenants; a tenant administrator is restricted to its own tenant. A domain may be resolved to a tenant for login-page discovery, but that resolution is not sufficient proof of access.

Tenant-aware entities observed in the reference include users, roles, departments, positions, notices, external identities, top navigation, logs, and inherited base records. Modura must explicitly classify every table as tenant-owned or global instead of relying on inheritance or query interceptors.

### 4.2 Tenant provisioning

Creating a tenant triggers an initialization transaction that provisions:

- the tenant record and stable public tenant code;
- a root organization/department;
- a default tenant-administrator role;
- a default executive position;
- an initial administrator user linked to those records.

This is an important business workflow, not generic CRUD. In Modura it should be an explicit `ProvisionTenant` application use case with atomicity, deterministic defaults, audit records, secure initial-credential delivery, and safe retry/idempotency behavior.

The reference initializes a predictable `admin/admin` credential. Modura must never do this. Prefer invitation or one-time credential setup.

### 4.3 Tenant removal

The reference uses logical deletion and cache invalidation but does not model the consequences for owned data. Modura needs an explicit lifecycle such as active, suspended, and scheduled-for-deletion, plus retention and irreversible purge rules. Physical cascading deletion must not be inferred from the reference.

## 5. Organization

### 5.1 Departments

Departments form a tenant-owned tree. Each node records its parent and an ancestor path to support subtree queries. Business rules visible in the source include:

- a top-level department belongs to the current tenant;
- a child inherits its parent’s tenant;
- a node cannot be its own parent;
- ordinary edits cannot move a department across tenants;
- destructive operations should respect descendants and dependent users/data.

The source does not consistently enforce cycle prevention beyond self-parenting or dependency-safe deletion. Modura must prevent arbitrary cycles and define deletion/reparenting behavior explicitly.

### 5.2 Positions

Positions are a tenant-owned catalog with code, name, category, and display order. They describe organizational assignment, not authorization. Position names are used during user import/export.

The intended relationship is many-to-many or an explicit user-position assignment entity if users may hold several positions. It must not be a delimited string on the user record.

### 5.3 User membership

The reference allows a user to carry several department and position IDs, but models them as strings. Modura should decide deliberately whether Phase 1 supports:

- one primary department plus zero or more secondary memberships;
- one or more positions, optionally scoped to a department;
- membership effective dates and primary flags.

If Phase 1 only needs one department and one position, that simpler cardinality should be stated rather than prematurely implementing a generalized model.

## 6. Authorization

### 6.1 Observed authorization dimensions

The reference combines three related grant dimensions:

1. Functional/UI grants: roles are assigned menu and button nodes.
2. API scopes: roles are assigned named API/path scope records.
3. Data scopes: roles are assigned rules describing which rows are visible.

Role grants are replaced as one transaction-like operation: remove the existing associations and insert the selected menu, API-scope, and data-scope associations. This “desired state replaces current state” behavior is useful for an admin UI, but Modura should apply optimistic concurrency or version checks to avoid lost updates.

### 6.2 Roles

Roles are tenant-owned and hierarchical. Users can hold multiple roles. Non-platform administrators cannot create the reserved platform-administrator role. A delegated administrator sees and grants only roles/resources already within its own authority.

Durable invariants for Modura:

- a user may only receive roles from the same tenant;
- a tenant actor cannot create or assign a platform role;
- delegated administration cannot grant more authority than the actor possesses;
- role deletion must handle user assignments and policy bindings explicitly;
- reserved system roles require stable identifiers and protection from rename/delete.

Role hierarchy is not automatically required. Keep it only if there is a concrete inheritance use case; a visual tree alone is insufficient justification.

### 6.3 Resources and UI projection

The reference menu tree contains both navigational nodes and button/action nodes. Role grants determine:

- accessible routes;
- visible navigation nodes;
- visible buttons/actions;
- route-to-authority metadata returned to the frontend.

When a child route is granted, parent nodes are added so the UI can render a connected tree. A separate “top menu” groups/subsets the same navigation tree; the effective navigation is the intersection of role-accessible nodes and the selected top-menu group.

The useful product behavior is permission-aware navigation. The model should be inverted in Modura:

```text
Resource + Action -> Permission/Policy -> authorization decision
                                  -> Route/Menu/Button projection
```

API paths and UI paths must not themselves be the canonical permission identifier. Hiding a button is never authorization enforcement.

Top menus are a UI workspace/navigation grouping feature, not an authorization primitive, and can be deferred unless the first admin shell truly needs multiple workspaces.

### 6.4 API scopes

The reference attaches named API/path scopes to menu nodes and roles. This captures a real need—roles need server-side capabilities independent of UI visibility—but binds the capability too closely to paths and menus.

Modura should define stable resources and actions (for example `user:read`, represented structurally rather than as scattered strings), map HTTP operations to them at the API boundary, and evaluate them through the authorization service. OpenAPI operations should reference the same canonical permission catalog.

### 6.5 Data scopes

The reference recognizes these row-visibility modes:

- all data;
- own records;
- current department;
- current department and descendants;
- custom values.

It also carries resource, field, class, column, and value expressions intended for generic interception. The business modes are useful; the reflective class/column/value mechanism is framework machinery and must not be copied.

Modura should resolve an explicit scope from actor, tenant, resource, and action, then pass a typed scope to the owning repository. Custom scope must have a defined domain representation, not arbitrary SQL fragments. Multiple-role combination semantics must be specified (normally union of allowed rows, with explicit deny behavior if supported).

## 7. System capabilities

### 7.1 Dictionaries

Dictionaries are hierarchical code/value catalogs. A root describes a dictionary type and children provide ordered values. They support centrally managed labels such as sex, notice category, position category, and scope type.

Retain dictionaries only for operator-managed business vocabularies. Compile-time protocol states and security-sensitive enums should remain code/schema-defined, not mutable dictionary entries.

The ownership model is unclear in the reference. Modura must decide whether a dictionary is global, tenant-owned, or a global definition with tenant overrides.

### 7.2 Configuration parameters

Parameters provide key/value configuration editable from the admin UI. This is useful for non-secret, runtime business settings. Secrets, boot-critical configuration, infrastructure configuration, and authorization policy must not be stored in a generic parameter table.

Configuration needs a scope (global or tenant), data type/schema, validation, audit history, and cache invalidation semantics.

### 7.3 Administrative regions

Regions form a global hierarchy from country/province through village, with ancestor paths and denormalized names for each level. The source prevents deleting a node that still has children and supports lazy tree loading.

This is a reusable reference-data capability, but it is not Phase 1 core unless a concrete business feature requires it. If added later, prefer immutable/versioned dataset import over routine manual CRUD for official region data.

### 7.4 Notices and dashboard

Notices contain tenant, category, title, content, release time, status, and audit metadata. The dashboard exposes recent and personal-notice views, but much of its sample content is static demonstration data.

A real notice capability would additionally need audience/recipient rules, draft/publish/withdraw lifecycle, publisher, read receipts, scheduling, and tenant isolation. The current reference does not fully implement these concepts, so it should not be treated as a complete design.

### 7.5 Operational logs and audit

The reference separates ordinary logs, API request logs, and error logs, recording variations of service, method, URI, IP, user, tenant, parameters, timing, stack trace, and result. It demonstrates the need for searchable operational diagnostics.

Modura should distinguish:

- security/business audit events, which are durable records of who changed what and must be integrity-conscious;
- application logs, which support operations;
- traces and metrics, which support observability.

These have different retention, access, privacy, and failure-handling requirements. Request bodies, credentials, tokens, and personal data must be redacted. An audit write should be coupled to the business transaction through an explicit strategy rather than treated as ordinary logging.

### 7.6 Auth clients

The reference contains client ID/secret, grant types, redirect URIs, scopes, authorities, and token lifetimes—essentially an OAuth client registry. Since Modura Phase 1 explicitly defers becoming a full OAuth authorization server, this module should also be deferred. It must not be confused with browser/session configuration.

### 7.7 Development tooling

Datasource management stores database connection credentials and code generation stores table/package/output-path recipes. Their product purpose is rapid CRUD scaffolding across databases.

They are not Phase 1 enterprise-domain capabilities. A future generator should consume Modura’s own module/OpenAPI conventions, produce predictable reviewable output, avoid server-side arbitrary filesystem paths, and never persist plaintext datasource secrets. Treat the existing feature set as demand evidence only.

## 8. Cross-cutting business patterns worth retaining

1. Tenant/account uniqueness, not global username uniqueness.
2. Tenant provisioning as one explicit application transaction.
3. Trust tenant identity only after server-side resolution and authorization.
4. Validate that every assigned role, department, position, parent, and related record belongs to the same tenant.
5. Derive self-service identity from the authenticated principal.
6. Derive bulk-import tenant from the operator context.
7. Make administrator grants desired-state replacements and audit before/after values.
8. Preserve parent navigation nodes when projecting authorized leaf routes.
9. Keep external identity separate from local account and bind atomically.
10. Lock or throttle authentication by both account and network signal without revealing account existence.
11. Invalidate cached authorization/configuration state immediately after relevant writes.
12. Reject deletion or movement that would violate tree and relationship invariants.

## 9. Structural and business risks discovered

These are not merely Java-style concerns; they affect correctness and should inform Modura’s design:

- Referential relationships are frequently delimited strings and lack database foreign keys/unique constraints.
- Several tables have only primary-key constraints; tenant-scoped uniqueness is enforced inconsistently in application code.
- Menu, route, button, API path, and permission concepts are coupled, making backend enforcement and UI presentation drift-prone.
- Tenant filtering is inconsistent and often depends on each query remembering to add it.
- Platform-administrator bypass is broad and should be modeled as explicit cross-tenant authorization with audit.
- Role grant targets are tenant-checked, but all granted menu/scope IDs are not consistently ownership/authority checked.
- Department moves update the moved node’s ancestor path but do not visibly update descendants atomically.
- Tenant deletion does not define suspension, retention, dependent-data handling, or session invalidation.
- Password reset uses a shared default password and does not force secure setup.
- Refresh tokens are self-contained and logout is effectively unimplemented.
- Social identity lookup should include tenant/provider/subject uniqueness and must prevent one provider identity being ambiguously reused across tenants.
- Generic condition binding and generic update endpoints create mass-assignment and authorization risks.
- Logs can capture sensitive request parameters without a domain-aware redaction policy.
- Mutable global dictionaries/parameters have unclear tenant ownership and change governance.
- Cache invalidation is broad and ad hoc, indicating missing ownership/versioning semantics.

## 10. Recommended Modura Phase 1 extraction

### Include

- Tenant lifecycle: create, view, update, suspend; provision initial organization and administrator securely.
- User identity: local account, profile, status, password lifecycle, server-side refresh sessions.
- Organization: department tree and a deliberately simple Phase 1 position/membership model.
- Authorization: user-role assignment, stable resource/action catalog, role policies, explicit data scopes.
- Admin-shell projection: derive routes and actions from permissions without making UI artifacts authoritative.
- Dictionary and non-secret configuration with explicit global/tenant ownership.
- Durable business audit for important writes.
- Tenant isolation, authorization, contract, and provisioning integration tests.

### Defer

- Social/external login and guest registration.
- Captcha unless threat data shows it is needed.
- Full OAuth authorization server and client registry.
- Top-menu/multi-workspace navigation.
- Notices and read receipts.
- Administrative-region datasets.
- Datasource administration and CRUD code generation.
- Report designer integration.
- General-purpose operational log viewer beyond the minimum observability baseline.

## 11. Open design decisions for Modura

The reference cannot safely answer these; they must be decided explicitly in the Constitution/AGENTS and architecture documents:

1. Is a login identifier unique per tenant, and how does the user select/discover the tenant?
2. Does Phase 1 allow email login, and must email uniqueness be tenant-scoped?
3. Does a user have one primary department or multiple memberships?
4. Are positions single or multiple, and are they scoped to departments?
5. Is role hierarchy a real requirement or should roles remain flat?
6. How are permissions registered and kept consistent with OpenAPI operations and frontend projections?
7. How do multiple role policies and data scopes combine, including custom scopes and denies?
8. Which dictionaries/configurations are global, tenant-defined, or tenant-overridable?
9. What are tenant suspension, deletion, retention, and restoration semantics?
10. How is tenant provisioning retried safely, and how is the first administrator invited?
11. Which identity or authorization changes revoke all sessions versus selected sessions?
12. Which audit events must be transactionally guaranteed?

## 12. Provenance

This analysis was derived from reading the bundled Apache-2.0 SpringBlade reference repository, primarily its module controllers, entities, service behavior, mapper queries, authentication flows, and database creation/seed script. It is a high-level factual and critical summary written independently for requirements research. No implementation code, comments, templates, SQL definitions, or seed data were copied.

After `SpringBlade-boot/` is removed, this document is the retained research artifact. It must not be interpreted as permission to reproduce SpringBlade code or as a compatibility contract.
