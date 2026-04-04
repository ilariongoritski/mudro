# Specification: import-project-memory

## Context
`mudro` already has a dense and valuable project memory in `.codex/*`, Telegram
control logs, and run-by-run operational notes. Skaro needs this context in a
structured form to become useful as a cockpit rather than an empty shell.

## User Scenarios
1. **Dashboard context:** I open Skaro and immediately understand current
   product state, hosting state, and active priorities.
2. **History handoff:** I can find the imported snapshots of project memory
   without manually re-reading every `.codex` file first.

## Functional Requirements
- FR-01: Copy current `.codex` memory files into `.skaro/docs/imported/`.
- FR-02: Create readable summary documents for project memory and runtime.
- FR-03: Preserve references to raw `.codex` logs instead of deleting them.

## Non-Functional Requirements
- NFR-01: Import must be repo-native and text-based.
- NFR-02: No secrets may be moved into tracked artifacts.

## Boundaries (what is NOT included)
- Full migration of every raw `.codex/logs/*` file into milestone tasks
- Replacement of `.codex/*` as the operational source of truth

## Acceptance Criteria
- [ ] Imported snapshots exist under `.skaro/docs/imported/`
- [ ] Summary docs exist for memory and runtime
- [ ] Public MVP links are documented in Skaro docs

