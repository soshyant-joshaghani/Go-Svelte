# Frontend

SvelteKit dashboard (Svelte 5, TypeScript, Tailwind 4, shadcn-svelte, adapter-node).

## Layout

```
src/
├── lib/config/backend.ts           # API_BASE_URL (/api/v1)
├── lib/modules/base/               # kit/platform (auth, shell, stores)
│   └── ui/                         # shadcn-svelte primitives
├── lib/modules/apps/<name>/api.ts  # per-app HTTP clients
└── routes/                         # SvelteKit pages (+page.svelte)
```

### Frontend modules (mandatory)

Under the frontend modules root there are **only**:

- `base/` — kit/platform (auth, users, shell, i18n, stores) + design primitives at `base/ui/`
- `apps/<domain>/` — product domains (API clients + UI), mirroring `backend/src/modules/apps/<domain>/`

There is **no** project `components/` folder as the app UI home. Modules are the component home.
Do not add `global/`, `shell/`, `layout/`, or a top-level `modules/ui/` peer of `base`/`apps`.
Where shadcn (or equivalent) is used: `ui` → `…/modules/base/ui`, `components` alias → `…/modules/base`.

## Canonical example

- API client: `src/lib/modules/apps/sample/api.ts`
- UI: `src/routes/sample/notes/+page.svelte`

Inspect these before building new features.

## Dev

From repo root:

```bat
npm run dev
```

Or via CLI: `__ctrl__\go-svelte-ctrl.bat dev run all`

Vite proxies `/api` → `API_PROXY_TARGET` (default `http://localhost:8000`).

In Docker dev, Traefik serves http://dashboard.localhost → Vite :5000.

## Production

Multi-stage `Dockerfile` builds with `npm run build -w frontend` and runs `node build` on port 5000.

Set `PUBLIC_API_BASE_URL` at build time (see `compose.yml`).

## Ecosystem

Install npm packages in this workspace for 3D (Three.js, Babylon.js), charts, UI libraries, etc.

Full docs: [docs/modules.md](../docs/modules.md) · [docs/architecture.md](../docs/architecture.md)
