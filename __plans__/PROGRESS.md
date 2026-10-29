# Progress

| Stage | Status | Note |
|-------|--------|------|
| Folder contract | done | frontend, backend, tests, traefik, docs, __plans__, __ctrl__ |
| Module groups | done | apps/sample, base/auth, base/users, system |
| Wire contract | done | Follows CONTRACT.md: Fast routes, snake_case, `detail` errors |
| Python CLI | done | dev, test (incl. `test contract`), app, prod, logs, flatten, remote |
| Redis cache and jobs | done | Soft-degrading cache, Redis list worker (Asynq removed) |
| Backend tests | done | `tests/backend` (own Go module, `replace` to `../../backend`, imports `internal/`): 26 tests, in-memory fakes, no database needed |
| Contract test | done | `tests/contract/contract_test.py`: 0 failed against Postgres 18 and Redis 8 |

Last update: 2026-10-01
