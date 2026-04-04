# Implementation Plan: agent-ops-observability

## Stage 1: Capture process model
**Goal:** Turn the current agent contour into a stable operator-facing description.  
**Dependencies:** none

### Inputs
- `AGENTS.md`
- `.codex` memory
- current agent task taxonomy

### Outputs
- future Skaro docs and task notes for the agent contour

### Risks
- Details can become stale if implementation changes faster than docs.

### DoD
- [ ] Safe tasks are listed
- [ ] Review-gate expectations are explicit
- [ ] Ops readers can distinguish planner, worker, reporter, and bot roles

## Verify
- Review the description against current `AGENTS.md` and `.codex` priorities
