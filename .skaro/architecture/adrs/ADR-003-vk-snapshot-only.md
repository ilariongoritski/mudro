# ADR-003: VK Remains Snapshot-Only

**Status:** accepted
**Date:** 2026-03-17

## Context
The project now treats Telegram as the actively refreshed source. Continuing to operate VK as a live sync target would add operational noise without corresponding product value.

## Decision
Keep VK as a snapshot-only source until a deliberate decision reopens it. New ongoing freshness work, importer evolution, and operational attention should target Telegram first.

## Alternatives
1. **Maintain VK as a live source in parallel:** rejected because it increases process and runtime noise for limited payoff.
2. **Remove VK entirely from the system:** rejected because the archive still has value for historical browsing and MVP completeness.

## Consequences
- Positive: less operational load, clearer source-of-truth policy, fewer moving parts.
- Negative: VK freshness is intentionally stale by policy.
- Risks: docs and operators can regress into treating VK as live unless the rule stays visible.
