# Specification: adr-backfill-key-decisions

## Context
Several durable architectural and operational decisions already exist, but they are fragmented across chats, `.codex`, README/runbooks, and implementation history. Without ADRs, future work risks relitigating settled choices.

## User Scenarios
1. **Architecture review:** a developer needs a crisp explanation of why the project chose a given model or runtime direction.
2. **Onboarding:** a new contributor should see durable decisions without trawling chat history.

## Functional Requirements
- FR-01: Create an ADR location under `.skaro/architecture/adrs/`.
- FR-02: Backfill ADRs for media normalization, comment model, VK snapshot-only policy, and VPS self-hosted frontend.
- FR-03: Keep ADRs concise, decision-oriented, and focused on durable facts.

## Non-Functional Requirements
- NFR-01: ADRs must not become implementation diaries.
- NFR-02: ADRs must remain additive and consistent with current code/docs.

## Boundaries (what is NOT included)
- New architecture redesign
- Product bugfix implementation
- Runtime rollout work

## Acceptance Criteria
- [ ] `.skaro/architecture/adrs/` exists
- [ ] Four key ADRs are present and marked with status/date
- [ ] ADRs reflect already-made decisions rather than speculative future ideas

## Open Questions
- Should future ADR numbering stay global across the repo or Skaro-only inside `.skaro/architecture/adrs/`?
