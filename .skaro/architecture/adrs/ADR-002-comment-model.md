# ADR-002: Canonical Comment Graph With Parent Links And Reactions

**Status:** accepted
**Date:** 2026-03-17

## Context
Telegram discussion data evolved beyond flat comment storage. The project needed a durable internal comment graph and a first-class place for comment reactions.

## Decision
Use a canonical comment model with explicit `parent_comment_id` links and a dedicated `comment_reactions` table. Importers and API code should preserve this graph and keep compatibility bridges additive where needed.

## Alternatives
1. **Keep comments effectively flat:** rejected because reply hierarchies become ambiguous and hard to reconstruct.
2. **Store replies and reactions only in raw JSON:** rejected because it weakens invariants and makes API behavior brittle.

## Consequences
- Positive: reply trees, reactions, and comment-level media become structurally reliable.
- Negative: importers and backfills are more complex than a flat append-only model.
- Risks: unsafe merge or dedupe logic can still damage referential consistency if not guarded.
