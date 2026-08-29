# PostgreSQL integration testing

Identity integration tests use a real PostgreSQL database selected by
`MODURA_TEST_DATABASE_URL`. For safety, the database name must end in `_test`.
The test owns that database's `modura` schema and drops/recreates only that
schema around its run. Never point the variable at a shared or production
database.

The normal `make verify` run skips PostgreSQL integration tests when the
variable is absent. CI and release evidence must set it so the tests execute;
a skipped run is not Phase 1 completion evidence.

For local development, place uncommitted test values in `backend/.env.test`.
This path is ignored by Git. The Go process does not load it automatically;
source it only for the command that runs the tests.

Example for Fish after creating a dedicated database:

```fish
set -gx MODURA_TEST_DATABASE_URL 'postgres://modura:password@127.0.0.1:5432/modura_test?sslmode=disable'
make backend-test
```
