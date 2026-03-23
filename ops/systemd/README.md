# Systemd runtime templates

Tracked `systemd` templates in this directory cover two runtime groups:

## Core MUDRO stack

- `mudro-api.service`
- `mudro-bot.service`
- `mudro-agent-worker.service`
- `mudro-agent-planner.service`
- `mudro-agent-planner.timer`

Runtime env files on VPS:
- `/etc/mudro/runtime/mudro-api.env`
- `/etc/mudro/runtime/mudro-bot.env`
- `/etc/mudro/runtime/mudro-agent.env`

Installer:

```bash
bash ops/scripts/install_mudro_systemd.sh
```

## Worker plane

- `claudeusageproxy.service`
- `openclaw.service`
- `skaro.service`

Runtime env files on VPS:
- `/etc/openclaw/runtime/claudeusageproxy.env`
- `/etc/openclaw/runtime/openclaw.env`
- `/etc/openclaw/runtime/skaro.env`

Installer:

```bash
bash ops/scripts/install_openclaw_systemd.sh
```

## Notes

1. Keep real secrets only in VPS runtime env files.
2. Do not commit server-local overrides and drop-ins.
3. `OpenClaw` and `Skaro` should use the local `claudeusageproxy` for accounting when the worker plane is deployed on VPS.
