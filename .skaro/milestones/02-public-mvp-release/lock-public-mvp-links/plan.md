# Implementation Plan: lock-public-mvp-links

## Stage 1: Verify current links
**Goal:** Confirm which links actually work.
**Dependencies:** none

### Inputs
- current Vercel URLs
- VPS public URL

### Outputs
- verified link status

### Risks
- preview URLs can change later

### DoD
- [ ] public Vercel URL responds
- [ ] VPS URL responds

---

## Stage 2: Record final recommendation
**Goal:** Preserve the public handoff choice in Skaro docs.
**Dependencies:** stage 1

### Inputs
- verification results

### Outputs
- `.skaro/docs/public-links.md`

### Risks
- release URL can change after later deploys

### DoD
- [ ] one Vercel URL is explicitly recommended
- [ ] VPS URL is explicitly documented

