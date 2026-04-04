# Implementation Plan: adr-backfill-key-decisions

## Stage 1: Define ADR storage and format
**Goal:** Create a stable place for decision records inside the Skaro cockpit.  
**Dependencies:** none

### Inputs
- `.skaro/templates/adr-template.md`
- current architecture docs

### Outputs
- `.skaro/architecture/adrs/`

### Risks
- ADR content may drift into operational detail instead of decision framing.

### DoD
- [ ] ADR directory exists
- [ ] ADR numbering and status fields are explicit

---

## Stage 2: Backfill key accepted decisions
**Goal:** Capture already-settled architectural and runtime policy decisions.  
**Dependencies:** stage 1

### Inputs
- `.codex/top10.md`
- `.codex/done.md`
- `README.md`
- `docs/ops-runbook.md`

### Outputs
- `.skaro/architecture/adrs/ADR-001-media-normalization.md`
- `.skaro/architecture/adrs/ADR-002-comment-model.md`
- `.skaro/architecture/adrs/ADR-003-vk-snapshot-only.md`
- `.skaro/architecture/adrs/ADR-004-vps-self-hosted-frontend.md`

### Risks
- Some decisions may need later revision as runtime evolves.

### DoD
- [ ] Each ADR has context, decision, alternatives, and consequences
- [ ] ADRs describe durable choices, not one-off fixes

## Verify
- Review the ADR directory and confirm the four records match existing project docs and memory
