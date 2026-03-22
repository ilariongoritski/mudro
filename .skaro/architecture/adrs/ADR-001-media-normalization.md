# ADR-001: Media Normalization As The Canonical Direction

**Status:** accepted
**Date:** 2026-03-17

## Context
Media handling in `mudro` started as inline JSON embedded in posts and comments. This was convenient initially, but it made indexing, dedupe, referential integrity, and future extensions harder to reason about.

## Decision
Treat normalized media tables (`media_assets`, `post_media_links`, `comment_media_links`) as the canonical direction for the data model. Legacy JSON payloads may remain as a compatibility layer during migration and read fallback, but new evolution should target the normalized model first.

## Alternatives
1. **Stay on JSON-only media forever:** rejected because it keeps media opaque and weakens integrity checks.
2. **Perform a hard cutover with no fallback:** rejected because it increases rollout risk across local and VPS contours.

## Consequences
- Positive: media becomes queryable, extensible, and safer to backfill or dedupe.
- Negative: the project temporarily carries both normalized and compatibility layers.
- Risks: fallback code may drift from canonical reads if not kept under explicit checks.
