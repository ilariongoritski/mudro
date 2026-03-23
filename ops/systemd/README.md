# MUDRO systemd runtime

This directory tracks the `mudro-api` systemd unit and its runtime env contract.

## Goal

Keep the VPS runtime reproducible and prevent `DSN` drift between:

- tracked repo config
- `/etc/systemd/system/mudro-api.service`
- `/etc/mudro/runtime/mudro-api.env`

## Files

- `mudro-api.service` — tracked unit file for `/usr/local/bin/mudro-api`
- `mudro-api.env.example` — tracked example for the runtime env file

## Install on VPS

From the checked-out repo on the server:

```bash
bash ops/scripts/install_mudro_api_systemd.sh
```

This will:

1. install `ops/systemd/mudro-api.service` into `/etc/systemd/system/mudro-api.service`
2. create `/etc/mudro/runtime/mudro-api.env` if it does not exist
3. remove the legacy `10-dsn.conf` override if present
4. reload systemd and restart `mudro-api.service`

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

For the current VPS layout, `MEDIA_ROOT=/root/projects/mudro/data/nu` is valid as long as that path resolves to the real export directory.
