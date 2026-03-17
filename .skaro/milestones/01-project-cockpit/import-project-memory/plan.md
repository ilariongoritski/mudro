# Implementation Plan: import-project-memory

## Stage 1: Import snapshots
**Goal:** Preserve current `.codex` state inside Skaro.
**Dependencies:** none

### Inputs
- `.codex/todo.md`
- `.codex/done.md`
- `.codex/top10.md`
- `.codex/memory.json`
- `.codex/time_runtime.json`
- `.codex/tg_control.jsonl`

### Outputs
- `.skaro/docs/imported/codex-todo.md`
- `.skaro/docs/imported/codex-done.md`
- `.skaro/docs/imported/codex-top10.md`
- `.skaro/docs/imported/codex-memory.json`
- `.skaro/docs/imported/codex-time-runtime.json`
- `.skaro/docs/imported/codex-tg-control.jsonl`
- `.skaro/docs/imported/codex-log-directories.txt`

### Risks
- imported data can become stale if not refreshed later

### DoD
- [ ] all imported files exist
- [ ] secret files are not copied

---

## Stage 2: Build summaries
**Goal:** Turn raw memory into usable Skaro-facing context.
**Dependencies:** stage 1

### Inputs
- imported snapshots
- current public hosting facts

### Outputs
- `.skaro/docs/project-memory-summary.md`
- `.skaro/docs/chat-and-runtime-summary.md`
- `.skaro/docs/public-links.md`

### Risks
- summaries can drift from raw memory over time

### DoD
- [ ] summaries reflect current project state
- [ ] raw imported sources are referenced

