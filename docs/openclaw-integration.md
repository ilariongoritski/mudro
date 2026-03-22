# OpenClaw integration for MUDRO

Goal: use Claude Opus as the planning and review layer, while Codex applies changes locally and OpenClaw runs the VPS worker plane.

## Runtime layout

- This chat: control plane
- Claude Opus: planner / reviewer / draft generator
- Local Codex: applies diffs, runs tests, owns repo truth
- Local Skaro: dashboard and usage accounting
- VPS OpenClaw: remote worker plane

## Local proxy and usage

- Local Claude calls go through `D:\mudr\_mudro-local\skaro\claude.env`
- Proxy usage is tracked in `D:\mudr\_mudro-local\skaro\usage_log.jsonl`
- Token summary is tracked in `D:\mudr\_mudro-local\skaro\token_usage.yaml`
- The Claude usage proxy is loopback-only by default.
- If you ever bind it beyond `127.0.0.1`, set `MUDRO_CLAUDE_PROXY_TOKEN` and require it on non-loopback requests.

## VPS gateway

- User install script: `scripts/openclaw/openclaw_install_user.sh`
- Gateway service helper: `scripts/openclaw/openclaw_gateway_user_service.sh`
- Post-install checks: `scripts/openclaw/openclaw_post_install_checks.sh`
- Root bootstrap: `scripts/openclaw/server_bootstrap_root.sh`

## Notes

- Keep secrets out of the repo.
- Keep auxiliary files under `D:\mudr\_mudro-local`.
- Use explicit UTF-8 for command output when text looks garbled.
- If `systemd --user` is unavailable on a host, the gateway helper falls back to a `nohup` launch.
