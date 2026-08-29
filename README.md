# Modura

Agent-native enterprise application framework for Go, with a separate AI-assisted React administration workspace.

## Layout

- `backend/`: Go modular monolith
- `admin/`: React administration and rapid-development workspace
- `api/`: authoritative OpenAPI contract shared by backend and admin
- `docs/`: architecture, security, development, ADR, and research documents

Read `AGENTS.md` before making changes. The one-command verification entrypoint is `make verify`.

## Backend configuration

The backend fails fast unless `MODURA_DATABASE_URL` and a signing key of at
least 32 bytes in `MODURA_AUTH_SIGNING_KEY` are present. See
`backend/.env.example` for all authentication and HTTP-cookie settings. The
process does not load dotenv files automatically; inject secrets through the
process environment or the deployment secret mechanism.

Real PostgreSQL identity integration tests use `MODURA_TEST_DATABASE_URL` and
require a dedicated database whose name ends in `_test`. See
`docs/development/postgresql-testing.md` for the destructive-safety boundary.
