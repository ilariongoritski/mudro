# CASINO ROADMAP 2026

## Phase 1: Core Slot & Roulette (DONE)
- [x] Base Hub/Handler architecture.
- [x] HMAC-SHA512 Provably Fair engine.
- [x] Basic Roulette with SSE streaming.
- [x] Multi-multiplier Plinko implementation.

## Phase 2: TMA Integration & Stability (CURRENT)
- [x] Plinko UI/State synchronization fix.
- [x] Backend context propagation for handler lifecycle.
- [x] Fix Handlers unit tests.
- [/] Roulette Streaming Hub (In progress: currently uses per-client polling).
- [ ] Blackjack TMA UI rollout.

## Phase 3: Social & Economy (NEXT)
- [ ] Global Win Feed (shared via bff-web).
- [ ] Daily Faucet for TG users.
- [ ] Tournament brackets for Plinko high-rollers.

## Hotfix Queue
- [x] RTP & MaxBet constraints in `.env`.
- [x] Handlers test context fix.
- [ ] SSE channel leak protection (audit current ticker implementation).
