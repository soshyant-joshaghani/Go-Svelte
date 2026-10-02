# Testing

`test backend` runs `go test ./...` in `tests/backend/`.

The handler tests live in `tests/backend/`, split by module path (`apps/sample`, `base/auth`, `base/users`, `system`, `core/config`, `core/jobs`, `core/security`) with the shared harness in `tests/backend/testkit`. `tests/backend` is a separate Go module whose path (`.../tests/backend`) is under the backend module's path and which `replace`s it with `../../backend`; Go allows that to import `internal/...`. They inject in-memory repositories, cache, and job queue, so they need neither Postgres nor Redis. They cover login and bearer errors, the superuser routes, notes isolation and caching, the local-only private routes, CORS, config, bcrypt, and the job retry protocol.

`test contract` runs `tests/contract/contract_test.py` (a copy of the file in `foxg-kit/contract/`) against a running API: `__ctrl__\go-svelte-ctrl.bat test contract --base http://localhost:8000`. It needs Postgres and, for the job check, Redis. `test frontend` runs Vitest (`tests/frontend`, shared with Fast-Svelte) and `svelte-check`.
