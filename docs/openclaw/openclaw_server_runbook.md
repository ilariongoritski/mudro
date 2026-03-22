# OpenClaw Server Runbook (VPS)

Source baseline:
- official docs: `docs.openclaw.ai`
- repo scripts in `scripts/openclaw/`

## What is automated in the repo

- Root bootstrap: `scripts/openclaw/server_bootstrap_root.sh`
- User install: `scripts/openclaw/openclaw_install_user.sh`
- Gateway user service: `scripts/openclaw/openclaw_gateway_user_service.sh`
- Post-install checks: `scripts/openclaw/openclaw_post_install_checks.sh`

## Recommended server flow

As `root`:

```bash
cd /root/projects/mudro || cd /root
bash scripts/openclaw/server_bootstrap_root.sh
```

As the OpenClaw user:

```bash
bash scripts/openclaw/openclaw_install_user.sh
openclaw doctor --generate-gateway-token --non-interactive --yes
bash scripts/openclaw/openclaw_gateway_user_service.sh
bash scripts/openclaw/openclaw_post_install_checks.sh
```

## Remote access

If the gateway is local-only on the VPS, use SSH port forwarding from your laptop:

```bash
ssh -N -L 18789:127.0.0.1:18789 <user>@<server-ip>
```

Then open:

- `http://127.0.0.1:18789/`

## Notes

- Keep tokens in local secrets, not in tracked repo files.
- Prefer the user service path over ad-hoc background processes.
- If `systemd --user` is unavailable, the gateway helper uses a `nohup` fallback.