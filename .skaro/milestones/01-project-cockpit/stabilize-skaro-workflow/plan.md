# Implementation Plan: stabilize-skaro-workflow

## Stage 1: Fix tracking rules
**Goal:** Make Skaro commitable without leaking secrets.
**Dependencies:** none

### Inputs
- `.gitignore`
- `.skaro/*`

### Outputs
- tracked `.skaro` project artifacts
- ignored `.skaro/secrets.yaml`

### Risks
- accidental over-ignore can hide useful files from git

### DoD
- [ ] `.skaro` project docs are trackable
- [ ] `secrets.yaml` remains ignored

---

## Stage 2: Document daily workflow
**Goal:** Make Skaro usable as a daily project cockpit.
**Dependencies:** stage 1

### Inputs
- local launcher
- VS Code tasks
- current role config

### Outputs
- `.skaro/docs/skaro-workflow-guide.md`
- `.skaro/ops/local-run.md`

### Risks
- guide can fall behind if workflow changes

### DoD
- [ ] guide covers local startup, editing, validation, and public links
- [ ] guide clearly separates Skaro from Codex memory

