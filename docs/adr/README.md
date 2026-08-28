# Architecture Decision Records

ADRs record durable choices that future contributors and agents must understand. Create one when a change:

- departs from a MUST/SHOULD architectural default;
- introduces or replaces a foundational dependency;
- changes module, tenant, authentication, authorization, data, or deployment boundaries;
- creates an independently deployed service;
- makes a difficult-to-reverse operational or data-model choice.

Do not create ADRs for routine implementation details.

## Lifecycle

Use status `Proposed`, `Accepted`, `Superseded`, or `Rejected`. Only an `Accepted` ADR can authorize a deviation. Never rewrite an accepted decision to hide history; supersede it with a new ADR.

Name files `NNNN-short-title.md`, beginning with `0001`. Each ADR must contain:

```markdown
# NNNN: Title

- Status: Proposed
- Date: YYYY-MM-DD
- Owners: names or team

## Context

## Decision

## Consequences

## Alternatives considered

## Contract impact

List affected AGENTS.md rules, exceptions, review/expiry conditions, and required follow-up.
```

The change that accepts an ADR must update affected authoritative documentation. An ADR cannot silently override legal, licensing, or user instructions.
