[![](./FoxG-Kit.png)](./FoxG-Kit.png)

# Go-Svelte

**Go + net/http + pgx + SvelteKit + PostgreSQL + Redis.**

The scale-tier Go kit of the FoxG family (starter tier: Fast, Elysia, Hono; enterprise tier: Rust, DotNet). It speaks the [FoxG wire contract](../../../CONTRACT.md), so the SvelteKit `frontend/` is the same one Fast-Svelte ships and a project can change backend without touching the UI, the database, or the Redis keys. The server is the Go standard library, SQL stays explicit, and jobs use the contract's Redis list (no queue library).

**Docs:** [AGENTS.md](AGENTS.md) · [ROADMAP.md](ROADMAP.md) · [docs/](docs/) · [plans](__plans__/PROGRESS.md) · [`__ctrl__`](__ctrl__/README.md)

```bat
__ctrl__\go-svelte-ctrl.bat setup-local
__ctrl__\go-svelte-ctrl.bat dev run all
```

Linux and macOS use `__ctrl__/go-svelte-ctrl.sh`. `setup-local` downloads Go modules with `go mod download` and installs the Svelte app with `npm install`. Prerequisites: Go 1.27, Node.js 22, Python 3.12+ for `__ctrl__`, Docker.

| Service | URL |
|---------|-----|
| Dashboard | http://dashboard.localhost |
| Sample notes | http://dashboard.localhost/sample/notes |
| API (Swagger) | http://api.localhost/docs |
| API (Scalar) | http://api.localhost/sdoc |
| Adminer | http://adminer.localhost |
| Traefik | http://localhost:8080 |
| Direct Vite | http://localhost:5000 |
| Direct API | http://localhost:8000/docs |
| Superuser | `admin@example.com` / `Admin@1234` |

## Layout

```text
backend/                          Go API and worker
  cmd/api, cmd/worker             the two processes
  migrations/                     plain SQL, applied when the API starts
  sql/queries/                    named SQL, loaded by the repositories
  internal/modules/apps/sample    notes, the canonical example
  internal/modules/base/          auth and users
  internal/modules/system/        health and private dev routes
  internal/core/                  config, db, cache, jobs, security, errors, docs
  internal/httpserver/            wiring, CORS
frontend/                         SvelteKit (copy of Fast-Svelte's UI)
tests/                            backend/ (Go, own module), contract/, frontend/
traefik/ docs/ __plans__/
__ctrl__/                         Python CLI: dev, test, app, prod, remote
compose.yml compose.dev.yml
```

Layers: Route (`net/http` handler) → Service → Repository (interface, pgx) → PostgreSQL. See [docs/architecture.md](docs/architecture.md).

## Runtime profiles

| Profile | Command | Includes |
|---------|---------|----------|
| Full | `dev run all` | Postgres, Redis, worker, Traefik, Adminer, API, Vite |
| Slim | `dev run all --slim` | Postgres, Traefik, Adminer, API, Vite |

Production always runs the full stack. See [docs/runtime-profiles.md](docs/runtime-profiles.md).

## Adding a feature

1. Copy `backend/internal/modules/apps/sample/` to `backend/internal/modules/apps/<name>/` (routes, service, repository, schemas).
2. Register `<name>.Routes` in `backend/internal/httpserver/server.go`.
3. Add `backend/migrations/NNNN_<name>.sql` when tables change, and the queries to `backend/sql/queries/`.
4. Add `frontend/src/lib/modules/apps/<name>/api.ts` and a route under `frontend/src/routes/(dashboard)/`.
5. Add tests in `tests/backend/<module path>/` using `tests/backend/testkit`.

`__ctrl__\go-svelte-ctrl.bat app create <name>` writes the stubs. Copy the depth of `sample` before adding rules.

## Changing backend

The Svelte `frontend/`, `traefik/`, `tests/frontend`, the module names, the wire contract, and the PostgreSQL schema are shared with [Fast-Svelte](../../../fast-kit/fast-template/fast-svelte/README.md). To move a project from another family:

1. Take the target template (`fast-svelte`, `rust-svelte`, `dotnet-svelte`, and so on).
2. Copy the product's `frontend/src/lib/modules/apps/<name>/` and its routes into it.
3. Rebuild `<name>` under the target's backend path. Keep the routes and JSON from [CONTRACT.md](../../../CONTRACT.md).
4. Point it at the same database. Fast's Alembic tables and this kit's `backend/migrations` are the same schema.

Index: [foxg-kit](../../../README.md).

## Tests

```bat
__ctrl__\go-svelte-ctrl.bat test all
__ctrl__\go-svelte-ctrl.bat test contract
```

`test backend` runs `go test ./...` in `tests/backend/`: handler tests with in-memory fakes for the repositories, the cache, and the job queue (no Postgres, no Redis). `test contract` runs [`tests/contract/contract_test.py`](tests/contract/contract_test.py) against a running API. See [docs/testing.md](docs/testing.md).

## Production

```bat
__ctrl__\go-svelte-ctrl.bat setup
__ctrl__\go-svelte-ctrl.bat clone
__ctrl__\go-svelte-ctrl.bat env
__ctrl__\go-svelte-ctrl.bat start
```

See [docs/deployment.md](docs/deployment.md).

## Environment

Copy `.env.example` to `.env`. Variable names match Fast (`SECRET_KEY`, `POSTGRES_*`, `REDIS_*`, `FIRST_SUPERUSER*`). The API and worker read `.env` from `.`, `..`, or `../..`; real environment variables win. `PUBLIC_API_BASE_URL` is where the frontend calls the API (`/api/v1` through the Vite proxy in dev).

Other families: [Fast-Svelte](../../../fast-kit/fast-template/fast-svelte/README.md), [Elysia-Svelte](../../../elysia-kit/elysia-template/elysia-svelte/README.md), [Hono-Svelte](../../../hono-kit/hono-template/hono-svelte/README.md), [Rust-Svelte](../../../rust-kit/rust-template/rust-svelte/README.md), [DotNet-Svelte](../../../dotnet-kit/dotnet-template/dotnet-svelte/README.md).
