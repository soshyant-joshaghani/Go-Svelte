# AGENTS.md — AI development contract

Read this file before architectural changes in Go-Svelte.

Go-Svelte is a product-agnostic foundation: SvelteKit + Go (`net/http`, pgx) + PostgreSQL + Redis. Implement features inside this architecture. Do not reshape it as FastAPI, Hono, Elysia, or Rust.

## Start here

| Question | Answer |
|----------|--------|
| What to read first? | This file → [`__plans__/PROGRESS.md`](__plans__/PROGRESS.md) → [docs/architecture.md](docs/architecture.md) → [CONTRACT.md](../../../CONTRACT.md) |
| Where does a feature go? | `backend/internal/modules/apps/<name>/` and `frontend/src/lib/modules/apps/<name>/` |
| What is the reference? | `backend/internal/modules/apps/sample/` (notes) |
| Where is business logic? | Service |
| Where is database access? | Repository (SQL text in `backend/sql/queries`) |
| How does the UI call the API? | `fetch` in the feature `api.ts`, base URL from `$lib/config/backend` |
| How do migrations work? | SQL files in `backend/migrations`, applied when the API starts |
| How are jobs added? | Enqueue from a service after the write succeeds; handle in the worker (`internal/core/jobs`) |
| How to run it? | `__ctrl__\go-svelte-ctrl.bat dev run all` |

## Plan loop

1. Take the first eligible `pending` stage in `__plans__/PROGRESS.md`, or the stage the user named.
2. Mark it `in_progress`.
3. Implement it.
4. Run `__ctrl__\go-svelte-ctrl.bat test all` for the touched surface.
5. Mark `done` only when tests pass.

Do not gitignore `__plans__/`.

## Backend layers

Route (`net/http` handler) → Service → Repository (interface, pgx) → PostgreSQL

| Layer | Responsibility |
|-------|----------------|
| Route | HTTP only: parse, call a service, shape the response |
| Service | Rules, authorization, cache, enqueue |
| Repository | Queries and writes only |
| Core | Config, database, cache, jobs, security, errors, docs pages |

Errors are `apierr.Error` in `backend/internal/core/apierr`. The response is always `{"detail": "..."}` with the status from [CONTRACT.md](../../../CONTRACT.md). Do not invent a second error envelope. Modules import `core`, never each other, except `sample` -> `base/auth` and `auth` -> `users`; `users` takes the auth guards as a parameter so the two do not import each other.

## Frontend

- Svelte 5 runes. Modules only under `base/` and `apps/`. There is no `components/` folder.
- Tailwind classes on elements. Tokens only in `frontend/src/app.css`.
- Do not hard-code the API origin. Use `apiBaseUrl()` from `$lib/config/backend`.
- The frontend is shared with Fast-Svelte. Do not fork the wire format for this kit.

## Data

- PostgreSQL is the system of record. The schema is Fast's; new tables go in `backend/migrations/NNNN_name.sql` with `IF NOT EXISTS`.
- Redis cache helpers no-op when Redis is down. Do not cache auth.
- Sample notes is the cache example: list and get write the cache, writes invalidate the owner prefix.
- Jobs use the Redis list protocol in the contract. Do not add a second queue library.

## What not to do

- Do not port Python, Hono, or Rust shapes into this kit. Follow the contract.
- Do not turn the kit into a product (CMS, shop, AI app).
- Do not put business rules in routes.
- Do not add a dependency without a reason.

## Definition of done

- Module files in the right places
- Migration if tables changed
- Auth on anything that is not public
- Tests for behavior that can fail (`test backend` runs `go test ./...` in `tests/backend/`, a separate Go module that imports `backend/internal/...`; `test contract` for route changes)
- `__plans__/PROGRESS.md` updated when a stage was in progress
