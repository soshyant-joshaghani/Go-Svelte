# Modules

| Group | Role |
|-------|------|
| `apps/sample` | Canonical notes example |
| `base/auth`, `base/users` | Login, session, user records |
| `system` | Health and private dev routes |

Paths: `backend/internal/modules/{apps,base,system}`, `backend/internal/core`, and `backend/cmd/{api,worker}`.

```bat
__ctrl__\go-svelte-ctrl.bat app create myfeature
```

Register the generated `Routes` in `backend/internal/httpserver/server.go`. Then copy the depth of `sample` before adding rules.

1. Copy `backend/internal/modules/apps/sample/` to `backend/internal/modules/apps/<name>/` (routes, service, repository, schemas).
2. Register `<name>.Routes` in `backend/internal/httpserver/server.go`.
3. Add `backend/migrations/NNNN_<name>.sql` when tables change, and the queries to `backend/sql/queries/`.
4. Add `frontend/src/lib/modules/apps/<name>/api.ts` and a route under `frontend/src/routes/(dashboard)/`.
5. Add tests in `tests/backend/<module path>/` using `tests/backend/testkit`.
