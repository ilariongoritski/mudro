# Specification: runtime-runbook-convergence

## Context
Operational truth is split between `.codex`, `docs/ops-runbook.md`, and newer Skaro notes. This creates avoidable confusion during recovery and handoff.

## User Scenarios
1. **Recovery pass:** an operator wants one coherent explanation of the runtime process and where to look next.
2. **Doc maintenance:** a contributor wants to update one contour without silently contradicting another.

## Functional Requirements
- FR-01: Identify overlapping runtime docs and their intended authority.
- FR-02: State where Skaro should summarize versus where `.codex` and runbooks remain canonical.
- FR-03: Keep runtime documentation aligned with current worktree boundaries.

## Non-Functional Requirements
- NFR-01: The result must reduce ambiguity rather than add another competing source.
- NFR-02: No runtime commands should be changed as part of this documentation task.

## Boundaries (what is NOT included)
- VPS rollout work
- service reconfiguration
- firewall or proxy changes

## Acceptance Criteria
- [ ] The authority split between `.codex`, Skaro, and runbooks is explicit
- [ ] No infrastructure changes are introduced
- [ ] Runtime docs become easier to navigate

## Open Questions
- Should there be a single index doc that links all runtime references, or is a split-by-authority model enough?
