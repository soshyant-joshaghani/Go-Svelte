# safe/ — keys, addresses, prod env (local only)

| Pattern | Purpose |
|---------|---------|
| `*-privatekey.pem` | SSH private key |
| `*-address.txt` | VM IP / hostname (first line) |
| `*-env.env` | Production secrets → uploaded as `~/projects/go-svelte/.env` |

| Files | Server id |
|-------|-----------|
| `ar-go-svelte-bamdad-*` | `go-svelte` |

Copy the `*.example` stubs, drop the `.example` suffix, and fill real values.

`*.pem`, `*.env`, `*-address.txt` are gitignored.

Upload env to VM:

```bat
go-svelte-ctrl.bat env
```

That copies `safe/ar-go-svelte-bamdad-env.env` → `~/projects/go-svelte/.env`.
