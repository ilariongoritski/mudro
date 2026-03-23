# MUDRO tracked systemd runtime

This directory tracks the VPS runtime contracts for:

- `mudro-api.service`
- `mudro-bot.service`
- `mudro-agent-worker.service`
- `mudro-agent-planner.service`
- `mudro-agent-planner.timer`
- `openclaw.service`
- `skaro.service`

## Goal

Keep the server runtime reproducible and prevent drift between:

- tracked repo config
- `/etc/systemd/system/*.service`
- `/etc/mudro/runtime/*.env`
- `/etc/openclaw/runtime/*.env`

## Core MUDRO runtime

Core services run as the dedicated `mudro` user and use:

- app tree: `/opt/mudro/app`
- binaries: `/opt/mudro/bin`
- runtime env: `/etc/mudro/runtime`
- writable state/cache: `/var/lib/mudro`
- logs: `/var/log/mudro`

Install or update the core stack from a checked-out repo on the VPS:

```bash
bash ops/scripts/install_mudro_systemd.sh
```

API-only compatibility path:

```bash
bash ops/scripts/install_mudro_api_systemd.sh
```

## OpenClaw / Skaro worker plane

Worker-plane services run as the dedicated `openclaw` user and use:

- runtime env: `/etc/openclaw/runtime`
- writable state: `/var/lib/openclaw`
- logs: `/var/log/openclaw`

Install tracked root-level units after the core MUDRO app tree is already deployed:

```bash
bash ops/scripts/install_openclaw_systemd.sh
```

## Validation

Run local static validation before pushing:

```bash
make validate-systemd-templates
```

On the VPS:

```bash
systemctl daemon-reload
systemctl status mudro-api mudro-bot mudro-agent-worker mudro-agent-planner.timer
systemctl status openclaw skaro
```

## Notes

- Real secrets stay out of git.
- `mudro-bot` and `mudro-agent` write operational files into `/opt/mudro/app/.codex`, so the installer keeps only that subtree writable.
- `mudro-bot` and `mudro-agent` may need Docker socket access for current control-plane commands. The installer adds `mudro` to the `docker` group if that group exists.

## Media contract

The API serves `/media/*` from `MEDIA_ROOT`.

The database stores media references in:

- `media_assets.original_url`
- `media_assets.preview_url`
- `post_media_links`
- `comment_media_links`

The intended shape is:

- relative paths like `photos/file.jpg` or `stickers/file.webp` live in the database
- API normalizes them to `/media/...`
- `MEDIA_ROOT` points to the host directory that actually contains those files

For the hardened VPS layout, prefer a dedicated media path such as `/var/lib/mudro/media` and set `MEDIA_ROOT` explicitly in `/etc/mudro/runtime/mudro-api.env`.
