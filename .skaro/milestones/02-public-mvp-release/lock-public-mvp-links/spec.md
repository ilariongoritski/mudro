# Specification: lock-public-mvp-links

## Context
The project has both a self-hosted VPS frontend and at least one Vercel URL.
Skaro should make the final public handoff link explicit, so the project can be
shown immediately after the final commit and push.

## User Scenarios
1. **External review:** I can send one stable public MVP link.
2. **Release clarity:** I know which Vercel URL is public and which is protected.

## Functional Requirements
- FR-01: Document verified public links.
- FR-02: Mark protected previews as non-final.
- FR-03: Keep one recommended Vercel URL for public review.

## Non-Functional Requirements
- NFR-01: Link selection must match real reachable URLs.

## Boundaries (what is NOT included)
- new Vercel deploy
- new DNS/domain setup

## Acceptance Criteria
- [ ] VPS and Vercel public links are documented
- [ ] protected preview is explicitly marked as unsuitable for final handoff

