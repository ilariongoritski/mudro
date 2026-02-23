---

## Autonomous recovery (allowed without asking)
Agents may perform **non-destructive** self-healing actions to fix transient local issues and continue the baseline.

### Retry policy (fail-fast with limited retries)
- For each baseline step: up to **2 attempts** (initial + 1 retry).
- If still failing after retry: STOP and ask (include exact error lines + log path).
- Always record retries in `.codex/state.md` and `.codex/logs/<run>/index.md`.

### Allowed self-healing actions (NON-DESTRUCTIVE)
Without asking, an agent may:
- Re-run the failed command once (e.g., `docker version`, `make dbcheck`, `make test`).
- If deps look stale/broken: do a soft restart:
  - `make down`
  - `make up`
  - `docker compose ps`
- If DB connection fails (connection refused / database starting): wait briefly (<= 10s) and retry `make dbcheck` once.
- If tests fail due to flaky environment (timeouts, temporary network pull): retry `make test` once, but never hide failures.

### Not allowed (requires confirmation)
- Any destructive action or potential data loss:
  - drop/truncate/reset DB
  - `docker compose down -v`
  - deleting volumes
  - `rm -rf`
- Any DB schema/migration changes (adding/removing/rewriting migrations)
- Any public JSON API contract changes

### Docker access issues
If the error mentions Docker socket/permissions (`/var/run/docker.sock`):
- Agent may run diagnostics only:
  - `id`, `groups`, `ls -l /var/run/docker.sock`, `docker version`
- Then STOP and ask (this is not fixable purely from inside agent without changing user session/permissions).