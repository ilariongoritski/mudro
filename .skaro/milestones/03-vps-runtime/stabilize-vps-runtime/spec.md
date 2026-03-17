# Specification: stabilize-vps-runtime

## Context
The VPS already hosts the public MVP runtime and a substantial ops surface.
Skaro should fit into this as an internal cockpit, not as a public service.

## User Scenarios
1. **Ops overview:** I can reason about how Skaro should live on the VPS later.
2. **Safe hosting:** I do not expose the project cockpit publicly by accident.

## Functional Requirements
- FR-01: Define the safe role of Skaro on VPS.
- FR-02: Keep runtime ownership separate from site/API ownership.

## Non-Functional Requirements
- NFR-01: VPS-hosted Skaro should default to loopback-only access.

## Boundaries (what is NOT included)
- actual VPS rollout in this task

## Acceptance Criteria
- [ ] intended role of Skaro on VPS is documented
- [ ] public site and internal cockpit are clearly separated

