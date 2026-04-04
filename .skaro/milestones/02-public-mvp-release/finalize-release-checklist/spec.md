# Specification: finalize-release-checklist

## Context
Before the final push, `mudro` needs a compact checklist that covers Skaro docs,
public MVP links, and the minimum validation for a showable release surface.

## User Scenarios
1. **Final push:** I can finish with one final commit and know what to verify.
2. **Public handoff:** I know which URL to share after the push.

## Functional Requirements
- FR-01: Document a short release checklist for the current MVP.
- FR-02: Include Skaro, frontend, backend, and public URL checks.

## Non-Functional Requirements
- NFR-01: Checklist must be short and realistic.

## Boundaries (what is NOT included)
- CI redesign
- production domain/TLS rollout

## Acceptance Criteria
- [ ] checklist exists in Skaro docs or task plan
- [ ] public MVP link is part of the checklist

