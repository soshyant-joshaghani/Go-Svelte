# Backend (Go / net/http)

Implements the FoxG wire contract (`CONTRACT.md`). Layers: Router -> Service -> Repository -> Postgres.

```text
cmd/api, cmd/worker        the HTTP server and the Redis list worker
internal/core/            config, db, migrate, cache, jobs, security, apierr, httpx, openapi
internal/modules/         apps/sample, base/{auth,users}, system
internal/httpserver/      wiring, CORS (handler tests live in ../tests/backend)
sql/queries/              named SQL (-- name: X), loaded by the repositories
migrations/               plain SQL, applied at startup in name order
```

## Commands

```sh
cd backend
go vet ./...
go test ./...             # no tests here: they live in ../tests/backend (own module, imports internal/)
(cd ../tests/backend && go test ./...)   # handler tests with in-memory fakes
go run ./cmd/api          # listens on APP_HOST:APP_PORT (default 0.0.0.0:8000)
go run ./cmd/worker       # BRPOP foxg:jobs
```

Configuration is read from the environment, then `.env` in `.`, `..`, `../..`. Migrations come from `MIGRATIONS_DIR`, `./migrations` or `../migrations`. Docs: `/docs` (Swagger), `/sdoc` (Scalar), `/api/v1/openapi.json`. Optional extra env: `BCRYPT_COST` (default 12).

Docker: `docker build -f backend/Dockerfile .` from the kit root. The image runs `/app/api`; the worker is `/app/worker`.
