"""go-svelte-ctrl test."""

from __future__ import annotations

import argparse
import sys

from utils.dev.cli import _setup_local, _tool_argv, PROJECT

CTRL = "go-svelte-ctrl"
BRAND = "go-svelte"


def _run(argv_name: str, args: list[str], cwd) -> int:
    argv = _tool_argv(argv_name, *args)
    if not argv:
        return 1
    print(f"[{BRAND}] {' '.join(argv)}")
    import subprocess
    proc = subprocess.run(argv, cwd=cwd)
    return proc.returncode


def _run_backend() -> int:
    print(f"[{BRAND}] Backend tests (go test ./... in tests/backend)")
    return _run("go", ["test", "./..."], PROJECT / "tests" / "backend")


def _run_frontend() -> int:
    front = PROJECT / "frontend"
    if not (front / "package.json").is_file():
        print(f"[{BRAND}] No frontend package. Skipping frontend tests.")
        return 0
    print(f"[{BRAND}] Frontend tests (vitest)")
    code = _run("npm", ["test"], PROJECT)
    if code != 0:
        return code
    print(f"[{BRAND}] Frontend check (svelte-check)")
    return _run("npm", ["run", "check"], front)


def _run_contract(args: argparse.Namespace) -> int:
    import subprocess

    script = PROJECT / "tests" / "contract" / "contract_test.py"
    if not script.is_file():
        print(f"error: missing {script}", file=sys.stderr)
        return 1
    cmd = [sys.executable, str(script), "--base", args.base, "--local", "--jobs"]
    print("[contract] " + " ".join(cmd))
    return subprocess.run(cmd, cwd=PROJECT).returncode



def cmd_test(args: argparse.Namespace) -> int:
    if args.target == "contract":
        return _run_contract(args)
    code = _setup_local(force_install=False)
    if code != 0:
        return code
    target = args.target
    if target in ("backend", "all"):
        code = _run_backend()
        if code != 0:
            print("Backend tests failed.", file=sys.stderr)
            return code
    if target in ("frontend", "all"):
        code = _run_frontend()
        if code != 0:
            print("Frontend tests failed.", file=sys.stderr)
            return code
    if target == "all":
        print("\nAll tests passed.")
    return 0


def build_test_subparser(sub: argparse._SubParsersAction) -> None:
    sp = sub.add_parser("test", help="Run backend and frontend tests")
    sp.add_argument("target", nargs="?", default="all", choices=("all", "backend", "frontend", "contract"))
    sp.add_argument("--base", default="http://localhost:8000", help="API origin for `test contract`")
    sp.set_defaults(func=cmd_test)
