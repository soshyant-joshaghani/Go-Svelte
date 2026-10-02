# Architecture

Go-Svelte follows the FoxG folder contract. The Svelte site lives in `frontend/`. It is a copy of Fast-Svelte's UI.

```text
go-svelte/
├── frontend/
├── backend/                 `backend/internal/modules/{apps,base,system}`, `backend/internal/core`, and `backend/cmd/{api,worker}`
├── tests/
├── traefik/
├── docs/
├── __plans__/
└── __ctrl__/                Python CLI
```

Request flow: Route (`net/http` handler) → Service → Repository (interface, pgx) → PostgreSQL. One Go module, two binaries (`cmd/api`, `cmd/worker`). Go keeps `cmd/` and `internal/` because that is how a module hides its packages. The packages under `internal/modules` import `internal/core`; `internal/httpserver` wires them.

Routes speak the [wire contract](../../../../CONTRACT.md): `/api/v1`, `snake_case` JSON, `{"detail": "..."}` errors, form login, JWT HS256. Data is PostgreSQL with the Fast schema. Redis is the cache and the job queue. Both degrade softly: a missing Redis never fails a request.
