"""Scaffold new app modules."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

from lib.config import ROOT

PROJECT = ROOT.parent
CTRL = "go-svelte-ctrl"
KIND = "go"

_NAME_RE = re.compile(r"^[a-z][a-z0-9_]*$")


def _validate_name(name: str) -> str:
    name = name.strip().lower().replace("-", "_")
    if not _NAME_RE.match(name):
        raise SystemExit(
            "Module name must start with a letter and contain only lowercase "
            "letters, digits, and underscores."
        )
    if name in {"sample", "base", "system", "global"}:
        raise SystemExit(f"Reserved module name: {name}")
    return name


def _write_if_missing(path: Path, content: str) -> bool:
    if path.exists():
        print(f"  skip (exists): {path.relative_to(PROJECT)}")
        return False
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")
    print(f"  created: {path.relative_to(PROJECT)}")
    return True


def _title(name: str) -> str:
    return name.replace("_", " ").title()


def _backend_files(name: str) -> None:
    title = _title(name)
    if KIND == "go":
        base = PROJECT / "backend/internal/modules/apps" / name
        body = (
            "package " + name + "\n\n"
            "import \"net/http\"\n\n"
            "// Routes registers GET <prefix>/" + name + ".\n"
            "func Routes(mux *http.ServeMux, prefix string) {\n"
            "\tmux.HandleFunc(\"GET \"+prefix+\"/" + name + "\", func(w http.ResponseWriter, _ *http.Request) {\n"
            "\t\tw.Header().Set(\"Content-Type\", \"application/json\")\n"
            '\t\t_, _ = w.Write([]byte("{\\"message\\":\\"' + title + ' module\\"}\\n"))\n'
            "\t})\n"
            "}\n"
        )
        _write_if_missing(base / "routes.go", body)
        return
    folder = "routes.ts" if KIND == "elysia" else "router.ts"
    base = PROJECT / "backend/src/modules/apps" / name
    if KIND == "elysia":
        body = f"""import {{ Elysia }} from 'elysia';

export const {name}Routes = new Elysia({{ prefix: '/api/{name}', tags: ['{title}'] }}).get(
\t'/',
\t() => ({{ message: '{title} module' }}),
);
"""
    else:
        body = f"""import {{ createRoute }} from '@hono/zod-openapi';
import {{ MessageSchema }} from '@hono-svelte/contracts';
import {{ createRouter }} from '../../../http.js';

const rootRoute = createRoute({{
\tmethod: 'get',
\tpath: '/',
\ttags: ['{title}'],
\tresponses: {{
\t\t200: {{
\t\t\tdescription: '{title} module',
\t\t\tcontent: {{ 'application/json': {{ schema: MessageSchema }} }},
\t\t}},
\t}},
}});

export const {name}Routes = createRouter().openapi(rootRoute, (c) =>
\tc.json({{ message: '{title} module' }}, 200),
);
"""
    _write_if_missing(base / folder, body)


def _frontend_files(name: str) -> None:
    root = PROJECT / "frontend"
    if not (root / "package.json").is_file() and (root / "web").is_dir():
        root = root / "web"
    api = root / "src/lib/modules/apps" / name / "api.ts"
    page = root / "src/routes/(dashboard)" / name / "+page.svelte"
    _write_if_missing(
        api,
        f"""export function moduleUrl(): string {{
\treturn '/api/v1/{name}';
}}
""",
    )
    title = _title(name)
    _write_if_missing(
        page,
        f"""<script lang="ts">
\timport {{ moduleUrl }} from '$lib/modules/apps/{name}/api';
</script>

<section class="rounded-xl border p-6">
\t<h2 class="text-2xl font-bold">{title}</h2>
\t<p class="mt-4 text-muted-foreground">API <code>{{moduleUrl()}}</code></p>
</section>
""",
    )


def cmd_app_create(args: argparse.Namespace) -> int:
    name = _validate_name(args.name)
    print(f"[{CTRL}] Scaffolding app module: {name}")
    _backend_files(name)
    _frontend_files(name)
    print()
    print("Next steps:")
    print(f"  1. Copy the depth of the sample module.")
    if KIND == "go":
        print(f"  2. Call {name}.Routes(mux, prefix) from backend/internal/httpserver/server.go.")
    elif KIND == "elysia":
        print("  2. .use the new routes from backend/src/app.ts.")
    else:
        print("  2. .route the new router from backend/src/app.ts.")
    print(f"  3. Run: __ctrl__\\{CTRL}.bat test backend")
    return 0


def build_app_subparser(sub: argparse._SubParsersAction) -> None:
    sp = sub.add_parser("app", help="Scaffold application modules")
    actions = sp.add_subparsers(dest="app_action", required=True)
    create_sp = actions.add_parser("create", help="Create a new app module skeleton")
    create_sp.add_argument("name", help="module name (e.g. bookmarks, orders)")
    create_sp.add_argument("--force", action="store_true", help="kept for the shared command surface")
    create_sp.set_defaults(func=cmd_app_create)
