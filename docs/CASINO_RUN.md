# Casino Local Run Guide

## Prerequisites
- Docker Engine / Docker Desktop with Compose
- Node.js and npm
- Go version specified in `go.mod`

## 1. Configure local environment

From the repository root, copy `.env.example` to `.env` and replace the `CHANGE_ME` placeholders with unique local values. In particular, set `POSTGRES_PASSWORD`, `CASINO_POSTGRES_PASSWORD`, `CASINO_INTERNAL_SECRET`, and `JWT_SECRET`. Do not commit the populated `.env` file.

## 2. Start Casino DB + API

```bash
docker compose -f ops/compose/docker-compose.casino.local.yml up -d casino-db
make migrate-casino
docker compose -f ops/compose/docker-compose.casino.local.yml up -d casino-api
```

To run the casino API directly, load the values from `.env` into your shell and run:

```bash
go run ./services/casino/cmd/casino
```

## 3. Start Main API

The feed API proxies `/api/casino/*` to the casino service with JWT auth and `X-Internal-Secret`.

```bash
go run ./services/feed-api/cmd
```

Ensure the main database is available on the configured local port and the casino API is reachable at the configured service URL.

## 4. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

- Full page: `http://localhost:5173/casino`
- TMA (all games): `http://localhost:5173/casino/miniapp`

## Game Modes

| Game | Tab | Endpoint |
| --- | --- | --- |
| Slots | slots | `POST /api/casino/spin` |
| Roulette (live) | roulette | `SSE /api/casino/roulette/stream` |
| Roulette (instant) | roulette | `POST /api/casino/roulette/instant-spin` |
| Blackjack | blackjack | `POST /api/casino/blackjack/start` |
| Plinko | plinko | `POST /api/casino/plinko/drop` |
| Bonus | bonus | `POST /api/casino/bonus/claim-subscription` |
| Faucet | button in TMA | `POST /api/casino/faucet/claim` |

## Relevant environment variables

- `CASINO_DSN` — connection string for the casino database.
- `CASINO_MAIN_DSN` — connection string for the main database.
- `CASINO_SERVICE_URL` — URL of the casino API.
- `CASINO_INTERNAL_SECRET` — unique local secret shared by the feed API and casino API.
- `CASINO_FAUCET_AMOUNT` — faucet amount for local testing.

Use only local development credentials. Never use example values or demo credentials in production.
