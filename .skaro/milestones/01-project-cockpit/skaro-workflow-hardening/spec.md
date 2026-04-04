# Specification: skaro-workflow-hardening

## Context
Skaro is initialized in the repo, but the local workflow is still inconsistent: UI port conventions diverge, validation is not exposed in VS Code tasks, and secret-handling rules are not fully encoded in repo hygiene.

## User Scenarios
1. **Daily operator start:** a developer opens the worktree and can reliably launch Skaro UI and dashboard from VS Code without guessing ports or commands.
2. **Safe local setup:** a developer can keep local Skaro secrets and usage files without risking accidental commits.

## Functional Requirements
- FR-01: VS Code tasks must expose `Skaro: UI`, `Skaro: Open Dashboard`, `Skaro: Status`, and `Skaro: Validate`.
- FR-02: Dashboard port and documented URL must be consistent across tasks, config, and local run notes.
- FR-03: Local-only Skaro secret/usage files must be ignored by git.

## Non-Functional Requirements
- NFR-01: The setup must remain Windows + WSL friendly.
- NFR-02: Changes must stay additive and avoid product runtime behavior changes.

## Boundaries (what is NOT included)
- Product runtime bugfixes
- VPS or deploy changes
- Replacing `.codex` with Skaro

## Acceptance Criteria
- [ ] `.vscode/tasks.json` includes a working validate task
- [ ] UI/dashboard references use one port convention
- [ ] `.gitignore` protects local Skaro secret/usage files
- [ ] `.skaro/ops/local-run.md` matches the actual workflow

## Open Questions
- Should Skaro validation later be exposed in CI, or remain local-only for now?
