# CASINO ROADMAP 2026

## Phase 1: Core Slot & Roulette (DONE)
- [x] Base Hub/Handler architecture.
- [x] HMAC-SHA512 Provably Fair engine.
- [x] Basic Roulette with SSE streaming.
- [x] Multi-multiplier Plinko implementation.

## Phase 2: TMA Integration & Stability (DONE)
- [x] Plinko UI/State synchronization fix.
- [x] Backend context propagation for handler lifecycle.
- [x] Fix Handlers unit tests.
- [x] Roulette Streaming Hub (SSE via shared broadcast hub).
- [x] Blackjack TMA UI rollout.

## Phase 3: Social & Economy (PARTIAL)
- [x] Global Win Feed (live-feed + top-wins endpoints, CasinoLiveSidebar).
- [x] Daily Faucet for TG users (POST /faucet/claim, 24h cooldown).
- [ ] Tournament brackets for Plinko high-rollers.

## Hotfix Queue
- [x] RTP & MaxBet constraints in .env.
- [x] Handlers test context fix.
- [x] SSE channel leak protection (subscriber cap 500, buffered channels).
- [x] Roulette wheel responsive CSS (no overflow on mobile).
