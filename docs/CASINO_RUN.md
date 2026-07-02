# Casino Local Run Guide

## Prerequisites
- Docker Desktop (running)
- Node.js >= 18
- Go 1.24+

## 1. Start Casino DB + API

```powershell
# Casino DB (port 5434) + API (port 8082)
docker compose -f ops/compose/docker-compose.casino.local.yml up -d casino-db
# Apply migrations
$npx psql or use Makefile: make migrate-casino
# Start casino API
docker compose -f ops/compose/docker-compose.casino.local.yml up -d casino-api
```

Or run casino locally:
```powershell
# Set env vars from .env (CASINO_DSN, CASINO_INTERNAL_SECRET, etc.)
go run ./services/casino/cmd/casino
```

## 2. Start Main API (feed-api, port 8080)

Feed-api proxies /api/casino/* to casino service with JWT auth + X-Internal-Secret.

```powershell
# Needs main DB on 5433 + casino on 8081
go run ./services/feed-api/cmd
```

## 3. Start Frontend (port 5173)

```powershell
cd frontend
npm install
npm run dev
```

## 4. Open Casino

- Full page: http://localhost:5173/casino
- TMA (all games): http://localhost:5173/casino/miniapp

## Game Modes

| Game | Tab | Endpoint |
|------|-----|----------|
| Slots | slots | POST /api/casino/spin |
| Roulette (live) | roulette | SSE /api/casino/roulette/stream |
| Roulette (instant) | roulette | POST /api/casino/roulette/instant-spin |
| Blackjack | blackjack | POST /api/casino/blackjack/start |
| Plinko | plinko | POST /api/casino/plinko/drop |
| Bonus | bonus | POST /api/casino/bonus/claim-subscription |
| Faucet | button in TMA | POST /api/casino/faucet/claim |

## Key .env vars

```
CASINO_DSN=postgres://postgres:postgres@localhost:5434/mudro_casino?sslmode=disable
CASINO_MAIN_DSN=postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
CASINO_SERVICE_URL=http://127.0.0.1:8081
CASINO_INTERNAL_SECRET=local-casino-internal-secret
CASINO_FAUCET_AMOUNT=1000
```
