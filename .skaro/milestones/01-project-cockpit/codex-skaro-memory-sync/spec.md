# Specification: codex-skaro-memory-sync

## Context
The project already has rich `.codex` memory, but Skaro currently has no imported summaries. This makes the cockpit look empty and disconnects task framing from actual project history.

## User Scenarios
1. **Quick re-entry:** a developer opens Skaro and immediately sees the stable summary of project state without reading raw logs.
2. **Process hygiene:** `.codex` remains the canonical history, while Skaro exposes only distilled snapshots.

## Functional Requirements
- FR-01: Create an imported-docs area under `.skaro/docs/imported/`.
- FR-02: Add at least one concise imported snapshot based on current `.codex` memory.
- FR-03: Document the separation of responsibilities between `.codex` and `.skaro`.

## Non-Functional Requirements
- NFR-01: Imported docs must be short, factual, and easy to refresh.
- NFR-02: No raw transcripts or secret-bearing content may be copied.

## Boundaries (what is NOT included)
- Full migration of `.codex` into Skaro
- Rewriting historical logs
- Runtime state export automation

## Acceptance Criteria
- [ ] `.skaro/docs/README.md` explains the doc split
- [ ] `.skaro/docs/imported/` exists and contains current summaries
- [ ] Imported docs clearly reference `.codex` as the source of truth

## Open Questions
- Should imported summaries be refreshed manually after major milestones or by a scripted helper later?
