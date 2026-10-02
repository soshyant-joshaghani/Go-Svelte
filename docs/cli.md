# CLI

```bat
__ctrl__\go-svelte-ctrl.bat <command>
```

| Command | Effect |
|---------|--------|
| `setup-local` | Install this kit's runtime dependencies |
| `dev run all` | Infra, API, worker, Vite |
| `dev stop all` | Stop host apps and compose |
| `test all` | Backend and frontend tests |
| `test contract` | Wire-contract test against a running API |
| `app create <name>` | Module stub (backend and frontend) |
| `prod start` / `prod stop` | Production compose |
| `logs` | Host and production logs |
| `flatten` / `restore-flat` | Single-root git history |
| `ping`, `clone`, `env`, `start`, `stop`, `status`, `update` | SSH operations from `servers.json` |

Linux and macOS use `go-svelte-ctrl.sh`. Details: [`__ctrl__/README.md`](../__ctrl__/README.md).
