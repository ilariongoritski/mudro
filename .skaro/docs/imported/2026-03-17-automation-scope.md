# Automation Worktree Scope

**Purpose:** define what this worktree is allowed to improve.

## In scope
- Skaro setup and workflow
- `.codex` memory structure and synchronization rules
- task-spec discipline
- agent contour process
- MCP/devex/process improvements
- repo hygiene and worktree organization

## Out of scope
- routine product bugfixes unrelated to automation/devex
- direct VPS or deploy reconfiguration
- normal runtime/API/DB product work

## Routing rules
- runtime bug, API bug, DB bug: move to `mudro11-bugs`
- nginx, VPS, Vercel, deploy: move to `mudro11-devops`

## Local expectations
- prefer minimal diffs
- prefer repeatable commands
- codify rules in docs and templates
- optimize for lower chaos and faster re-entry into the project
