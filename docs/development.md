# Development

```bat
__ctrl__\go-svelte-ctrl.bat setup-local
__ctrl__\go-svelte-ctrl.bat dev run all
```

| Surface | URL |
|---------|-----|
| Dashboard | http://dashboard.localhost |
| API docs | http://api.localhost/docs |
| Scalar | http://api.localhost/sdoc |
| Direct API | http://localhost:8000/docs |
| Vite | http://localhost:5000 |

The API runs with `go run ./cmd/api` in `backend/`; the worker with `go run ./cmd/worker`. `dev run all --slim` skips Redis and the worker. Only one Traefik stack can bind port 80: stop the other proxy, or run `dev run apps` and use the direct URLs.

For AI-assisted work point the agent at [AGENTS.md](../AGENTS.md) and the `sample` module.
