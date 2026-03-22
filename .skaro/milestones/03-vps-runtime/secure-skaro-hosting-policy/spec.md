# Specification: secure-skaro-hosting-policy

## Context
The guide explicitly warns against exposing a future VPS-hosted Skaro UI publicly without access control. That safety rule should be captured inside the cockpit itself.

## User Scenarios
1. **Future hosting decision:** a contributor considers serving Skaro from a server and needs the guardrail in writing.
2. **Security review:** an operator wants to confirm that Skaro is treated as an internal tool, not a public dashboard.

## Functional Requirements
- FR-01: Record that Skaro UI must remain private unless access control is designed and reviewed.
- FR-02: Explicitly forbid exposing `.skaro/secrets.yaml` and imported raw memory through any hosted UI.
- FR-03: Keep the rule visible in Skaro docs and milestone planning.

## Non-Functional Requirements
- NFR-01: The policy must be short and unambiguous.
- NFR-02: The policy must not require runtime implementation in this worktree.

## Boundaries (what is NOT included)
- Building hosted Skaro
- Access-control implementation
- VPS reverse proxy work

## Acceptance Criteria
- [ ] The hosting rule is explicitly documented
- [ ] Sensitive Skaro artifacts are named as non-public
- [ ] The rule is visible in the runtime/process milestone

## Open Questions
- If hosted Skaro ever becomes necessary, should it live behind VPN, SSO, or both?
