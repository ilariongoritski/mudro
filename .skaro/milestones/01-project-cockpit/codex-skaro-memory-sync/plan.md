# Implementation Plan: codex-skaro-memory-sync

## Stage 1: Create imported docs structure
**Goal:** Add a stable location for imported summaries and define its rules.  
**Dependencies:** none

### Inputs
- `.codex/todo.md`
- `.codex/done.md`
- `.codex/top10.md`
- `.skaro/docs/`

### Outputs
- `.skaro/docs/README.md`
- `.skaro/docs/imported/README.md`

### Risks
- The imported layer can become a duplicate of `.codex` instead of a summary.

### DoD
- [ ] The imported docs purpose is documented
- [ ] The rules forbid raw-log dumping into Skaro

---

## Stage 2: Import current stable summary
**Goal:** Provide a compact snapshot of current project state and worktree boundaries.  
**Dependencies:** stage 1

### Inputs
- `.codex` memory files
- worktree scope rules from `AGENTS.md`

### Outputs
- `.skaro/docs/imported/2026-03-17-codex-snapshot.md`
- `.skaro/docs/imported/2026-03-17-automation-scope.md`

### Risks
- Important scope boundaries can be oversimplified.

### DoD
- [ ] Current project shape is summarized
- [ ] Scope routing rules are explicit

## Verify
- `powershell -ExecutionPolicy Bypass -File .\scripts\skaro-local.ps1 status`
- Review `.skaro/docs/imported/` for summaries only, with no raw session dumps
