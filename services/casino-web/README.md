# casino-web — Sweet Bonanza Slot (MUDRO Casino)

Next.js 16 + Tailwind + shadcn slot machine, fully integrated with MUDRO casino ecosystem.

## Features

- **Telegram-only login** (primary and only registration method)
- Real-time spin via `casino-api` with shared wallet
- Balance sync with `mudro-casino-db`
- Spin history
- Fairness proof (serverSeedHash + nonce)
- Responsive mobile-first UI

## Tech Stack

- Next.js 16 (App Router, Server Actions ready)
- Bun (runtime) + Node 20 (build)
- Tailwind + shadcn/ui
- Zustand (state)
- Direct fetch to `casino-api`

## Environment Variables

```env
NEXT_PUBLIC_CASINO_API_URL=http://mudro-casino-api-1:8081
NEXT_PUBLIC_AUTH_API_URL=http://mudro-auth-api-1:8080
```

## Running Locally

```bash
bun install
bun dev
```

## Docker

```bash
docker compose -f docker-compose.casino-web.yml up -d
```

## API Integration

- `POST /spin` — real backend spin with JWT
- `GET /wallet/balance` — live balance
- `GET /history` — last 20 spins
- `POST /api/auth/telegram` — Telegram WebApp login

## RTP Configuration

RTP is stored in `casino_config` table (managed by `casino-api`).

Current way to change RTP:

```sql
UPDATE casino_config SET rtp_percent = 96 WHERE id = 1;
```

Admin UI is planned for future versions.

## Statistics Collected

- Per-user: total_bet, total_win, spin_count, win_rate
- Per-game: current RTP, symbol distribution (future)

## Weak Points & Future Improvements

- Add React Query + caching
- Error boundaries
- Fairness proof modal
- Pagination in history
- Rate limiting on frontend

## Deployment

See `docker-compose.casino-web.yml` and nginx location `/casino`.

## License

Part of MUDRO project.
