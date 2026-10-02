# Tests

```bat
__ctrl__\go-svelte-ctrl.bat test all
__ctrl__\go-svelte-ctrl.bat test backend
__ctrl__\go-svelte-ctrl.bat test frontend
__ctrl__\go-svelte-ctrl.bat test contract
```

`test backend` runs `go test ./...` in `tests/backend/`. The Go handler tests live there, like the other kits' `tests/backend`, in folders that mirror the module paths: `apps/sample`, `base/auth`, `base/users`, `system`, `core/{config,jobs,security}`. `testkit/` holds the shared harness (in-memory fakes for the repositories, the cache, and the job queue, plus `testkit.New`), so the tests need no database and no Redis.

`tests/backend` is its own Go module (`.../tests/backend`) with `replace <backend module> => ../../backend`. Go enforces `internal/` by import path: a module whose path sits under the backend module's path may import `<backend module>/internal/...`, so the tests reach the real packages without exporting anything from `backend/`. The Docker image never copies `tests/`.

`test contract` runs `tests/contract/contract_test.py` (the file from `foxg-kit/contract/`) against a running API, by default `http://localhost:8000`, with `--local --jobs`.

`tests/frontend` holds the Vitest suite shared with Fast-Svelte. `test frontend` runs it and `svelte-check`.
