#!/usr/bin/env python3
import argparse
import concurrent.futures
import datetime as dt
import json
import os
import pathlib
import sys
import urllib.error
import urllib.request


DEFAULT_ROLES = ["architect", "backend", "frontend", "tester", "devops", "security", "integration", "data"]


def load_text(path: pathlib.Path) -> str:
    return path.read_text(encoding="utf-8")


def trim_text(text: str, limit: int = 4000) -> str:
    value = text.strip()
    if len(value) <= limit:
        return value
    return value[:limit].rstrip() + "\n...[truncated]"


def build_context_pack(repo_root: pathlib.Path) -> str:
    context_files = [
        ("Service Catalog", repo_root / "docs" / "service-catalog.md"),
        ("Services Map", repo_root / "platform" / "agent-control" / "services-map.yaml"),
        ("Microservices Iteration 1", repo_root / "docs" / "microservices-iteration-1.md"),
        ("Microservices Iteration 2", repo_root / "docs" / "microservices-iteration-2.md"),
        ("API Gateway Contract", repo_root / "contracts" / "http" / "api-gateway-v1.yaml"),
        ("BFF Web Contract", repo_root / "contracts" / "http" / "bff-web-v1.yaml"),
    ]

    parts = []
    for label, path in context_files:
        if not path.exists():
            continue
        try:
            parts.append(f"## {label}\n{trim_text(load_text(path))}")
        except Exception:
            continue
    return "\n\n".join(parts)


def repo_head(repo_root: pathlib.Path) -> tuple[str, str]:
    import subprocess

    def run(*args: str) -> str:
        try:
            completed = subprocess.run(
                ["git", *args],
                cwd=str(repo_root),
                capture_output=True,
                text=True,
                check=True,
            )
            return completed.stdout.strip()
        except Exception:
            return "unknown"

    return run("branch", "--show-current"), run("rev-parse", "--short", "HEAD")


def repo_command(repo_root: pathlib.Path, *args: str) -> str:
    import subprocess

    try:
        completed = subprocess.run(
            ["git", *args],
            cwd=str(repo_root),
            capture_output=True,
            text=True,
            check=True,
        )
        return completed.stdout.strip()
    except Exception:
        return ""


def build_worktree_pack(repo_root: pathlib.Path) -> str:
    status = repo_command(repo_root, "status", "--short")
    diff_stat = repo_command(repo_root, "diff", "--stat")
    parts = []
    if status:
        parts.append(f"## Worktree Status\n{trim_text(status, 3000)}")
    if diff_stat:
        parts.append(f"## Worktree Diff Stat\n{trim_text(diff_stat, 3000)}")
    return "\n\n".join(parts)


def build_prompt(repo_root: pathlib.Path, task: str, role_prompt: str) -> str:
    branch, commit = repo_head(repo_root)
    context_pack = build_context_pack(repo_root)
    worktree_pack = build_worktree_pack(repo_root)
    return (
        "Task: review and plan the next safe microservice iteration for MUDRO.\n"
        f"User request: {task}\n"
        f"Repository root: {repo_root}\n"
        f"Branch: {branch}\n"
        f"Commit: {commit}\n"
        "Constraints:\n"
        "- Actual repository is Go + Postgres + React.\n"
        "- Do not assume Node/Express/Prisma unless explicitly present.\n"
        "- Prefer incremental migration over big-bang rewrite.\n"
        "- Preserve backward compatibility with current runtime.\n\n"
        "Repository context pack:\n"
        f"{context_pack}\n\n"
        "Current worktree pack:\n"
        f"{worktree_pack}\n\n"
        f"{role_prompt}\n"
    )


def call_role(proxy_url: str, model: str, role: str, task: str, repo_root: pathlib.Path, run_id: str, role_prompt: str) -> dict:
    payload = {
        "model": model,
        "max_tokens": 4000,
        "messages": [
            {
                "role": "user",
                "content": build_prompt(repo_root, task, role_prompt),
            }
        ],
    }
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        proxy_url.rstrip("/") + "/v1/messages",
        data=data,
        headers={
            "content-type": "application/json",
            "anthropic-version": "2023-06-01",
            "x-mudro-role": role,
            "x-mudro-agent-id": f"role-{role}",
            "x-mudro-run-id": run_id,
            "x-mudro-project": "mudro11",
            "x-mudro-agent-kind": "claude-opus-role",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=360) as resp:
            body = resp.read().decode("utf-8")
            parsed = json.loads(body)
            text = ""
            for block in parsed.get("content", []):
                if block.get("type") == "text":
                    text += block.get("text", "")
            return {
                "role": role,
                "ok": True,
                "status": resp.status,
                "response": parsed,
                "text": text.strip(),
            }
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        return {"role": role, "ok": False, "status": exc.code, "error": raw}
    except Exception as exc:
        return {"role": role, "ok": False, "status": 0, "error": str(exc)}


def main() -> int:
    parser = argparse.ArgumentParser(description="Run parallel Claude role prompts through the local proxy.")
    parser.add_argument("--task", required=True, help="English task for the role matrix")
    parser.add_argument("--repo-root", default=str(pathlib.Path(__file__).resolve().parents[2]), help="Repository root")
    parser.add_argument("--roles", nargs="+", default=DEFAULT_ROLES)
    parser.add_argument("--model", default=os.environ.get("MUDRO_CLAUDE_MODEL", "claude-opus-4.6"))
    parser.add_argument("--proxy-url", default=os.environ.get("MUDRO_CLAUDE_PROXY_URL", "http://127.0.0.1:8788"))
    parser.add_argument("--run-id", default=dt.datetime.now(dt.UTC).strftime("%Y%m%d-%H%M%S-role-matrix"))
    parser.add_argument("--max-workers", type=int, default=4)
    parser.add_argument(
        "--output-dir",
        default=os.environ.get("MUDRO_CLAUDE_RUNS_DIR", r"D:\mudr\_mudro-local\claude-orch\runs"),
    )
    args = parser.parse_args()

    repo_root = pathlib.Path(args.repo_root).resolve()
    roles_dir = repo_root / "ops" / "claude-workers" / "roles"
    run_dir = pathlib.Path(args.output_dir) / args.run_id
    run_dir.mkdir(parents=True, exist_ok=True)

    results = []
    max_workers = max(1, min(len(args.roles), int(args.max_workers)))
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = []
        for role in args.roles:
            role_path = roles_dir / f"{role}.md"
            if not role_path.exists():
                results.append({"role": role, "ok": False, "status": 0, "error": f"missing role file: {role_path}"})
                continue
            futures.append(
                executor.submit(
                    call_role,
                    args.proxy_url,
                    args.model,
                    role,
                    args.task,
                    repo_root,
                    args.run_id,
                    load_text(role_path),
                )
            )
        for future in concurrent.futures.as_completed(futures):
            results.append(future.result())

    summary = {
        "run_id": args.run_id,
        "model": args.model,
        "proxy_url": args.proxy_url,
        "repo_root": str(repo_root),
        "roles": sorted(args.roles),
        "results": sorted(results, key=lambda item: item["role"]),
    }
    (run_dir / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")

    for result in summary["results"]:
        out_path = run_dir / f"{result['role']}.md"
        if result.get("ok"):
            out_path.write_text(result.get("text", ""), encoding="utf-8")
        else:
            out_path.write_text(result.get("error", "unknown error"), encoding="utf-8")

    print(json.dumps({"run_id": args.run_id, "output_dir": str(run_dir), "roles": summary["roles"]}, ensure_ascii=False))
    return 0 if all(item.get("ok") for item in summary["results"]) else 1


if __name__ == "__main__":
    sys.exit(main())
