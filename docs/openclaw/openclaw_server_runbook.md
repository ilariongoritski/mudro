# OpenClaw Server Runbook (VPS)

Source baseline:
- `D:/тело/клов0603.pdf` (OCR snapshot saved to `output/claw0603_ocr.txt`)
- official docs: `docs.openclaw.ai` and `github.com/openclaw/openclaw`

## 1) What is automated in repo

- Root bootstrap script: `scripts/openclaw/server_bootstrap_root.sh`
- User install script: `scripts/openclaw/openclaw_install_user.sh`
- Post-install checks: `scripts/openclaw/openclaw_post_install_checks.sh`

## 2) What still needs your manual input

- Telegram bot token from `@BotFather`
- OAuth login confirmations in local browser during `openclaw onboard`
- Provider/mode selection in onboarding wizard (OpenAI Codex, model, channel)

## 3) Exact execution order on server

As `root`:

```bash
cd /root/projects/mudro || cd /root
bash /path/to/repo/scripts/openclaw/server_bootstrap_root.sh
```

As non-root user (`openclaw` or your custom user):

```bash
bash /path/to/repo/scripts/openclaw/openclaw_install_user.sh
openclaw onboard --install-daemon
```

After onboarding:

```bash
bash /path/to/repo/scripts/openclaw/openclaw_post_install_checks.sh
```

## 4) Telegram pairing (from PDF flow)

When bot replies with pairing code, approve on server:

```bash
openclaw pairing approve telegram <code>
```

## 5) Web UI access for remote VPS

Use local SSH tunnel from your laptop:

```bash
ssh -N -L 18789:127.0.0.1:18789 <user>@<server-ip>
```

Then open:
- `http://127.0.0.1:18789/`

## 6) Important provider warnings (from PDF + docs)

- OpenAI Codex via ChatGPT OAuth: supported workflow in article.
- Anthropic subscription OAuth in third-party tools: high ban risk.
- Google Antigravity OAuth: high ban risk.
- For production stability, prefer official API keys/providers allowed by provider ToS.

## 7) VPN note

You already validated Mullvad in container (`US, Los Angeles` egress).
This runbook installs OpenClaw on host (not containerized gateway).  
If you want strict VPN-only egress for OpenClaw itself, we need separate network design
(gateway in container or host-wide tunnel policy).
