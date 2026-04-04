# Specification: prepare-skaro-vps-host

## Context
A future always-on Skaro dashboard on VPS is useful only if it is safe and fits
the current `mudro` deployment style.

## User Scenarios
1. **Remote access:** I can reach Skaro from anywhere through a safe path.
2. **Predictable hosting:** I know what service and access model to use later.

## Functional Requirements
- FR-01: Define a future VPS hosting pattern for Skaro.
- FR-02: Prefer `systemd` plus loopback-only binding and SSH tunnel access.

## Non-Functional Requirements
- NFR-01: No public unauthenticated Skaro UI by default.

## Boundaries (what is NOT included)
- implementation of the VPS service in this local step

## Acceptance Criteria
- [ ] VPS hosting approach is documented
- [ ] SSH tunnel or equivalent internal access path is explicit

