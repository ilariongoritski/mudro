# Implementation Plan: runtime-runbook-convergence

## Stage 1: Map authority boundaries
**Goal:** Define which runtime knowledge belongs to `.codex`, Skaro, and runbooks.  
**Dependencies:** none

### Inputs
- `.codex/state.md`
- `.codex/done.md`
- `docs/ops-runbook.md`
- `.skaro/docs/`

### Outputs
- updated Skaro process docs and task notes

### Risks
- The split can still be ignored unless it is repeated in operator-facing docs.

### DoD
- [ ] Authority boundaries are explicit
- [ ] No duplicate runtime story is introduced

## Verify
- Review the resulting docs and confirm each runtime fact has a clear canonical home
