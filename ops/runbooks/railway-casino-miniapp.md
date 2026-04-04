# Railway Rollout: Casino Mini App

## Services

Deploy three Railway services from the same repository:

1. `mudro-frontend`
   - root directory: `frontend`
   - dockerfile: `frontend/Dockerfile`
   - public service
2. `mudro-feed-api`
   - root directory: repository root
   - dockerfile: `services/feed-api/Dockerfile`
   - private or public HTTP service
3. `mudro-casino`
   - root directory: repository root
   - dockerfile: `services/casino/Dockerfile`
   - private or public HTTP service

## Required env

### Frontend

- `VITE_API_BASE_URL=https://<feed-api-domain>/api`

### Feed API

- `API_ADDR=:$PORT`
- `DSN=<main mudro postgres dsn>`
- `JWT_SECRET=<secret>`
- `CASINO_SERVICE_URL=https://<casino-domain>`
- `API_ALLOWED_ORIGINS=https://<frontend-domain>`
- `TELEGRAM_BOT_TOKEN=<bot token>`
- optional rate-limit envs if used

### Casino

- `CASINO_ADDR=:$PORT`
- `CASINO_DSN=<supabase dsn with sslmode=require>`
- `CASINO_MAIN_DSN=<main mudro postgres dsn>`
- `CASINO_START_BALANCE=500`
- `CASINO_RTP_BPS=9500`
- `CASINO_MAX_BET=1000`
- `CASINO_ROULETTE_BETTING_MS=15000`
- `CASINO_ROULETTE_LOCK_MS=1500`
- `CASINO_ROULETTE_SPIN_MS=4500`
- `CASINO_ROULETTE_RESULT_MS=3500`
- `TELEGRAM_BOT_TOKEN=<bot token>`
- bonus-specific envs when enabled:
  - `CASINO_BONUS_TELEGRAM_BOT_TOKEN=<bot token used for getChatMember>`
  - `CASINO_BONUS_TELEGRAM_CHANNEL=<channel username or id>`
  - `CASINO_BONUS_FREE_SPINS=10`
  - optional `CASINO_BONUS_TELEGRAM_API_BASE=https://api.telegram.org`

### Optional Railway automation

- `RAILWAY_API_TOKEN=<railway api token>` only if deploys or smoke checks are scripted through Railway CLI/API.

## Healthchecks

- frontend: `/`
- feed-api: `/healthz`
- casino: `/healthz`

## Startup order

1. Ensure `mudro-casino` has valid Supabase envs.
2. Run main wallet / RTP migrations against main Mudro DB:
   - `bash ./scripts/migrate-casino-main.sh`
3. Run casino microservice migrations against Supabase:
   - `bash ./scripts/migrate-casino.sh`
4. Deploy `mudro-casino`
5. Deploy `mudro-feed-api`
6. Deploy `mudro-frontend`

## Smoke checks

- frontend opens `/tma/casino`
- `GET https://<feed-api>/healthz`
- `GET https://<casino>/healthz`
- `GET https://<feed-api>/api/casino/balance` with auth
- `GET https://<feed-api>/api/casino/roulette/state` with auth
- roulette SSE stream stays open through `/api/casino/roulette/stream`

## Notes

- No plaintext secrets are stored in repo.
- Casino stays on Supabase; the core Mudro feed/auth data plane stays on its own database.
- If the frontend is deployed on a separate Railway domain, set `API_ALLOWED_ORIGINS` and `VITE_API_BASE_URL` together.
