# Local Run Notes

## Tooling location
Skaro local toolchain is intentionally placed on `D:`:

- tools: `D:\mudr\toolchain\uv-tools`
- executables: `D:\mudr\toolchain\uv-bin`
- cache: `D:\mudr\toolchain\uv-cache`
- managed Python installs: `D:\mudr\toolchain\uv-python`

## Recommended launch path
Use the wrapper instead of calling `skaro` directly from PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\skaro-local.ps1 status
powershell -ExecutionPolicy Bypass -File .\scripts\skaro-local.ps1 ui
```

The wrapper forces:
- `UV_*` paths to `D:`
- UTF-8 console output
- `skaro` resolution from `D:\mudr\toolchain\uv-bin`

## VS Code workflow
Recommended local workflow for `mudro`:

1. Open the repo in VS Code.
2. Edit `.skaro/*` files and project code as usual.
3. Run workspace tasks:
   - `Skaro: UI`
   - `Skaro: Open Dashboard`
   - `Skaro: Status`
   - `Skaro: Validate`
4. Keep the dashboard open at `http://127.0.0.1:4700/dashboard`.

Workspace tasks are stored in `.vscode/tasks.json`.

## How this maps to project logs
`Skaro` should complement, not replace, the existing `mudro` working memory.

Keep using:
- `.codex/todo.md`
- `.codex/done.md`
- `.codex/state.md`
- `.codex/logs/*`

Use `Skaro` for:
- constitution
- architecture
- invariants
- devplan
- milestones/tasks
- validation

Use `.codex/*` for:
- session history
- operational notes
- recovery context
- run-by-run evidence

Skaro internal local usage files are separate:
- `.skaro/secrets.yaml`
- `.skaro/token_usage.yaml`
- `.skaro/usage_log.jsonl`

## Agent settings in current setup
Current local role split:
- default: `OpenAI / gpt-5.4`
- architect: `OpenAI / gpt-5.4`
- coder: `OpenAI / gpt-5.3-codex`
- reviewer: `OpenAI / gpt-5.4`

This is configured in `.skaro/config.yaml`.

## LLM credentials
Current project config points Skaro to `OPENAI_API_KEY`.

Options to make LLM phases work locally:
- export `OPENAI_API_KEY` in the shell before launch
- or create `.skaro/secrets.yaml` locally

`secrets.yaml` is gitignored and must never be committed.

## Verify command assumptions
Repo-level verify commands assume:
- backend and `make` commands run through WSL at `/mnt/d/mudr/mudro11`
- frontend build/lint run via `npm.cmd`

Current Go verify excludes scratch-only package `tmp/`, because `tmp/feed_check.go` is not canonical product code.
