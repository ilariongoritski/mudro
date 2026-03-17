# Implementation Plan: skaro-workflow-hardening

## Stage 1: Align local launch flow
**Goal:** Make the launch path deterministic for UI, dashboard, and status checks.  
**Dependencies:** none

### Inputs
- `.vscode/tasks.json`
- `.skaro/config.yaml`
- `.skaro/ops/local-run.md`

### Outputs
- `.vscode/tasks.json`
- `.skaro/config.yaml`
- `.skaro/ops/local-run.md`

### Risks
- VS Code task syntax may drift from actual Skaro CLI expectations.

### DoD
- [ ] Tasks and docs point to the same dashboard port
- [ ] Validate is callable from VS Code without editing JSON manually

---

## Stage 2: Encode safety rules
**Goal:** Prevent local Skaro secret and usage files from creating git noise or leak risk.  
**Dependencies:** stage 1

### Inputs
- `.gitignore`
- `.skaro` local file conventions

### Outputs
- `.gitignore`

### Risks
- Ignoring too much could hide legitimate tracked docs.

### DoD
- [ ] `.skaro/secrets.yaml` is ignored
- [ ] Skaro usage artifacts are ignored without hiding project docs

## Verify
- `powershell -ExecutionPolicy Bypass -File .\scripts\skaro-local.ps1 status`
- Open `http://127.0.0.1:4700/dashboard` after starting `Skaro: UI`
