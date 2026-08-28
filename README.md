# Modura

Agent-native enterprise application framework for Go, with a separate AI-assisted React administration workspace.

## Layout

- `backend/`: Go modular monolith
- `admin/`: React administration and rapid-development workspace
- `api/`: authoritative OpenAPI contract shared by backend and admin
- `docs/`: architecture, security, development, ADR, and research documents

Read `AGENTS.md` before making changes. The eventual one-command verification entrypoint is `make verify`.
