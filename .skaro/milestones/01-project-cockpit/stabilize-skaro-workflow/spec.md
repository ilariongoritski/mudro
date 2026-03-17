# Specification: stabilize-skaro-workflow

## Context
Skaro is already bootstrapped for `mudro`, but it must be easy to use daily.
This requires clear workflow docs, working VS Code tasks, and a predictable
split between `.skaro` and `.codex`.

## User Scenarios
1. **Daily work:** I can open VS Code, run Skaro UI, and know what to update.
2. **Safe usage:** I know which files belong to Skaro and which stay in Codex.

## Functional Requirements
- FR-01: Document the daily workflow for local Skaro usage.
- FR-02: Keep VS Code tasks available for UI, status, and validate.
- FR-03: Ensure `.skaro` is commitable except for secrets and usage files.

## Non-Functional Requirements
- NFR-01: No change should break the current local or VPS runtime.
- NFR-02: The workflow must stay simple and Windows/WSL-compatible.

## Boundaries (what is NOT included)
- VPS-hosted Skaro service rollout
- OpenClaw integration

## Acceptance Criteria
- [ ] guide exists for daily Skaro usage
- [ ] `.skaro` files are trackable in git
- [ ] secrets remain ignored

