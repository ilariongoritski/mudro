# Claude Usage Proxy on VPS

This document tracks the minimal VPS-side runtime contract for `claudeusageproxy.service`.

## Purpose

`claudeusageproxy.service` provides a tracked local proxy endpoint for Claude-backed accounting used by:

- `Skaro`
- `OpenClaw`

The goal is to keep token accounting and usage ledgers on the VPS in a tracked, reproducible service layout without committing secrets.

## Tracked files

- `ops/systemd/claudeusageproxy.service`
- `ops/systemd/claudeusageproxy.env.example`
- `ops/scripts/install_claudeusageproxy_systemd.sh`
- `scripts/openclaw/claudeusageproxy_linux.sh`

## Runtime files on VPS

- `/etc/openclaw/runtime/claudeusageproxy.env`
- `/var/lib/openclaw/claude-orch/ledger/usage_log.jsonl`
- `/var/lib/openclaw/claude-orch/ledger/token_usage.yaml`
- `/var/lib/openclaw/claude-orch/ledger/role_usage.yaml`

## Install

```bash
cd /opt/mudro/app
sudo bash ops/scripts/install_claudeusageproxy_systemd.sh
```

## Notes

- Secrets stay only in `/etc/openclaw/runtime/claudeusageproxy.env`
- `skaro.service` and `openclaw.service` may point `MUDRO_CLAUDE_PROXY_URL` to `http://127.0.0.1:8788`
- this service is intentionally separate from the core MUDRO stack
