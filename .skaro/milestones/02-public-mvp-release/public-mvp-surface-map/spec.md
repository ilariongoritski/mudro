# Specification: public-mvp-surface-map

## Context
The project now has multiple public-facing surfaces: VPS-hosted frontend, auxiliary Vercel URL, API/media paths, and review screenshots. Without one stable map, review and acceptance are harder than they should be.

## User Scenarios
1. **External review:** a reviewer needs the correct public link and a clear note about what is canonical.
2. **Internal routing:** a contributor needs to know which surface is primary and which is auxiliary.

## Functional Requirements
- FR-01: Document the canonical public MVP surface and auxiliary review surfaces.
- FR-02: Record which paths matter for review (`/`, `/api`, `/media`, `/healthz`).
- FR-03: Distinguish runtime ownership from review convenience.

## Non-Functional Requirements
- NFR-01: The map must stay short and review-oriented.
- NFR-02: No secrets or operator-only access details may be exposed.

## Boundaries (what is NOT included)
- Deploy changes
- Infrastructure hardening
- Product UI implementation work

## Acceptance Criteria
- [ ] Public surfaces are listed with clear roles
- [ ] Canonical vs auxiliary surfaces are explicitly separated
- [ ] The task avoids deploy instructions beyond high-level ownership

## Open Questions
- When domain/TLS is ready, should the canonical surface record move from IP-based to domain-based references immediately?
