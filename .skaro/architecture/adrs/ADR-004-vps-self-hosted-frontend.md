# ADR-004: Self-Hosted VPS Frontend Is The Canonical Public MVP Surface

**Status:** accepted
**Date:** 2026-03-17

## Context
The public MVP originally depended on Vercel as the main external surface. The project later gained a VPS-hosted `nginx` contour that can serve the frontend directly and proxy API/media locally.

## Decision
Treat the VPS-hosted frontend as the canonical public MVP runtime surface. Auxiliary hosting such as Vercel may still be used for review or convenience, but it is not the primary runtime dependency anymore.

## Alternatives
1. **Keep Vercel as the only public surface:** rejected because it keeps the MVP dependent on an extra runtime layer.
2. **Eliminate auxiliary hosting entirely:** rejected because preview hosting can still be useful for review flows.

## Consequences
- Positive: the public MVP is closer to the actual backend and media runtime and less fragmented operationally.
- Negative: VPS runtime docs and hardening now matter more.
- Risks: operators may accidentally expose internal surfaces if reverse proxy rules and documentation drift.
