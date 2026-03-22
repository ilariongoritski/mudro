# Specification: release-review-checklists

## Context
Public MVP review currently depends too much on memory and chat context. The project needs lightweight checklists for external review and release readiness.

## User Scenarios
1. **Pre-release pass:** a contributor wants a short list of checks before calling the MVP reviewable.
2. **External acceptance:** a reviewer wants to know what should be checked visually and functionally.

## Functional Requirements
- FR-01: Define a concise review checklist for public MVP surface checks.
- FR-02: Distinguish operator checks from reviewer-visible checks.
- FR-03: Keep the checklist compatible with both VPS-hosted and auxiliary preview surfaces.

## Non-Functional Requirements
- NFR-01: The checklist must fit a small operational pass, not a heavy QA program.
- NFR-02: Items must be observable and concrete.

## Boundaries (what is NOT included)
- Automated UI test implementation
- Browser automation authoring
- Production incident response policy

## Acceptance Criteria
- [ ] Checklist categories are clear
- [ ] User-visible checks are separated from internal checks
- [ ] The checklist avoids product-spec sprawl

## Open Questions
- Which parts of the release checklist should eventually become automated smoke tests?
