# Specification: agent-ops-observability

## Context
The project now includes an agent contour plus reporter/bot operational flows, but their process expectations are distributed across code, `.codex`, and runbooks.

## User Scenarios
1. **Ops review:** an operator wants to know what the safe agent tasks are and where review gates apply.
2. **Process maintenance:** a contributor wants the agent contour described as a stable system, not only as implementation detail.

## Functional Requirements
- FR-01: Describe the current agent/reporter/bot contour at the process level.
- FR-02: Capture review-gate expectations and safe-task boundaries.
- FR-03: Record the observability/recovery expectations at a high level.

## Non-Functional Requirements
- NFR-01: The description must stay process-oriented and avoid sensitive operational details.
- NFR-02: The result must support future task specs and operator docs.

## Boundaries (what is NOT included)
- Bot feature development
- Incident response automation
- Deploy or credentials work

## Acceptance Criteria
- [ ] Safe-task boundaries are documented
- [ ] Review-gate behavior is described
- [ ] Reporter/bot/agent roles are understandable from docs alone

## Open Questions
- Which agent metrics belong in long-lived docs, and which should remain only in ephemeral reports?
