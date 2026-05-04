# Secrets Rotation Runbook

## Rules

- Never write real secrets to tracked files.
- Store runtime env only on the VPS, for example under `/etc/mudro/runtime/*.env` or provider secret storage.
- Before editing env files, create a root-only backup outside the repo:
  - `install -d -m 700 /root/secret-backups`
  - `cp -a /etc/mudro/runtime /root/secret-backups/runtime-$(date +%Y%m%d-%H%M%S)`
- After rotation, restart only services that consume the changed secret.
- After validation, revoke the old secret at the provider.

## Inventory

Typical VPS files:
- API: `/etc/mudro/runtime/mudro-api.env`
- Bot: `/etc/mudro/runtime/mudro-bot.env`
- Agent: `/etc/mudro/runtime/mudro-agent.env`
- OpenClaw/Skaro if deployed: `/etc/openclaw/runtime/*.env`

Tracked examples are allowed, but must contain placeholders only:
- `ops/systemd/*.env.example`
- `env/*.env.example`

## Telegram Bot Tokens

Rotate when `TELEGRAM_BOT_TOKEN`, `CASINO_BOT_TOKEN`, `CASINO_BONUS_TELEGRAM_BOT_TOKEN`, or `REPORT_BOT_TOKEN` may be exposed.

Steps:
1. Create a new token in BotFather for the affected bot.
2. Update only the VPS runtime env file that owns the bot token.
3. Restart the affected service:
   - `systemctl restart mudro-bot`
   - `systemctl restart mudro-api` when Telegram WebApp auth uses the changed token.
   - `systemctl restart mudro-casino` if the casino service owns the changed token.
4. Verify:
   - `systemctl status mudro-bot --no-pager`
   - `curl -fsS http://127.0.0.1:8080/healthz`
5. Revoke the old token in BotFather.

## OpenAI / OpenRouter Keys

Rotate when `OPENAI_API_KEY` or `OPENROUTER_API_KEY` may be exposed.

Steps:
1. Create a new provider key with the minimal required scopes and budget limits.
2. Update the runtime env file that consumes the key.
3. Restart consumers:
   - `systemctl restart mudro-bot`
   - restart OpenClaw/Skaro services only if their runtime env uses that key.
4. Verify one low-cost request path or bot command.
5. Revoke the old provider key.

## JWT Secret

Rotate `JWT_SECRET` only with planned auth impact: current JWTs become invalid.

Steps:
1. Announce expected forced logout window.
2. Generate a strong random secret on the VPS:
   - `openssl rand -base64 48`
3. Update `JWT_SECRET` in the API/auth runtime env file.
4. Restart auth consumers:
   - `systemctl restart mudro-api`
   - `systemctl restart mudro-auth-api` if split auth-api is deployed.
5. Verify:
   - `curl -fsS http://127.0.0.1:8080/healthz`
   - login/register flow from the frontend.
6. Confirm old tokens are rejected by using a pre-rotation token against an authenticated endpoint.

## MinIO / S3 Credentials

Rotate when `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, or equivalent S3 credentials may be exposed.

Preferred approach:
1. Create a new non-root service account or access key with bucket-limited permissions.
2. Update application runtime env to use the new access key.
3. Restart services that read/write media.
4. Verify:
   - media upload/read path if enabled;
   - existing `/media/*` URLs still load;
   - `curl -fsS "https://<domain>/api/front?limit=1"` still returns media URLs.
5. Disable the old key.
6. Rotate root credentials separately and keep them outside app env.

## Post-Rotation Audit

Run:

```bash
git status --short
git grep -n "change-me\|OPENAI_API_KEY=\|OPENROUTER_API_KEY=\|TELEGRAM_BOT_TOKEN=\|JWT_SECRET=" -- ':!*.example'
```

Expected:
- no real secret values in tracked files;
- runtime services are healthy;
- old provider keys/tokens are revoked.
