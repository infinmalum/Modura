# Module Boundary Contract

## Purpose

Modura is a modular monolith first. Modules are cohesive business ownership boundaries inside one process; they are not miniature network services.

## Package shape

Business modules live at `backend/internal/modules/<module>`. Each module exposes a deliberately small application-level API to the composition root and permitted consumers. Domain, SQL, generated queries, and transport adapters remain private to the owner.

HTTP adapters live with their owning module under `backend/internal/modules/<module>/transport/http`. They may depend on Gin, generated contract types, and small application APIs, while domain/application code remains transport-independent. `backend/internal/api/handler` only composes these adapters into the single generated `ServerInterface`; it must not accumulate module business handling. Shared cookie, CSRF, bearer, and problem-response mechanics live under `backend/internal/api/transport` and contain no business decisions.

`backend/internal/platform` contains technical capabilities such as database connection setup, telemetry, clocks, and identifier generation. It must not become a business-domain dumping ground.

## Ownership rules

1. Every database table and migration has exactly one owning module, recorded in the migration header or a schema ownership manifest once tooling exists.
2. Only the owner writes its tables.
3. A module does not import another module's repository, SQL package, private domain representation, or transport handler.
4. Cross-module commands and authoritative reads go through the owner's public application API.
5. A public application API accepts domain-relevant explicit inputs, `context.Context`, verified actor/tenant information where required, and returns stable results/errors without transport types.
6. Dependency cycles are forbidden, including cycles hidden through shared packages.
7. Shared business concepts are not moved to `shared` merely to break a cycle. Revisit ownership or use an event/workflow coordinator.

## Cross-module reads and joins

A module may read another module's data only through its application API by default.

A direct cross-module read model or SQL join is allowed only when all conditions hold:

- it is read-only;
- the query serves an explicit reporting/list projection rather than a domain decision;
- table ownership remains clear;
- tenant and authorization rules for every source are enforced;
- the owner approves the dependency in an ADR or documented read-model contract;
- the consumer tolerates the owner changing its private schema only through coordinated migration.

Never use a cross-module join to modify another owner's data or bypass its invariants. Prefer a dedicated projection owned by the consuming/reporting module when the read becomes performance-critical or independently deployable.

## Transactions and workflows

A single-module use case owns its transaction. A workflow spanning modules is coordinated by an application workflow component and calls public application APIs.

If atomicity across several modules is genuinely mandatory while they share PostgreSQL, the coordinator may supply a transaction scope through an explicit, documented contract. It must not reach into private repositories. If atomicity is not mandatory, use events with idempotent consumers and visible retry/failure state.

## Events

Use an event when the producer owns a completed fact and does not require an immediate consumer result—for example, propagating an audit or projection update. Do not use events to disguise a synchronous query or avoid defining ownership.

Events must have a stable name/version, event ID, occurrence time, actor/tenant correlation where applicable, idempotency semantics, and an owner. In-process delivery is the default. Durable outbox/inbox infrastructure is added only when delivery guarantees require it.

## Extraction to a service

Splitting a module requires an accepted ADR demonstrating at least one present reason: independent scaling, resource isolation, fault containment, team/lifecycle ownership, or external service value. The ADR must cover data ownership, consistency, failure modes, observability, authentication, authorization, and migration. gRPC is preferred for internal synchronous calls after extraction.

## Automated enforcement target

CI should verify:

- imports do not cross private module paths;
- module dependencies are acyclic;
- migrations declare a valid owner;
- generated SQL belongs to its owner;
- forbidden reference-source imports/build inputs are absent.

SQL ownership, read-model exceptions, workflow semantics, and quality of public APIs still require review.
