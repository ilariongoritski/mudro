#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

RUN_ID_INPUT="${1:-}"
TASK_INPUT="${2:-}"

RUN_ID="${RUN_ID_INPUT:-$(date +%Y%m%d-%H%M%S)}"
TASK="${TASK_INPUT:-unspecified}"

RUN_DIR=".codex/logs/${RUN_ID}"
LOG_FILE="${RUN_DIR}/index.md"

mkdir -p "$RUN_DIR"

resolve_git_dir() {
  local git_dir=""

  if [[ -f ".git" ]]; then
    git_dir="$(sed -n 's/^gitdir: //p' .git 2>/dev/null | head -n1)"
    if [[ -n "$git_dir" && "$git_dir" =~ ^[A-Za-z]:[\\/].* && -x "$(command -v wslpath 2>/dev/null || true)" ]]; then
      git_dir="$(wslpath -u "$git_dir")"
    fi
  fi

  printf "%s" "$git_dir"
}

GIT_DIR_PATH="$(resolve_git_dir)"
if [[ -n "$GIT_DIR_PATH" ]]; then
  BRANCH_NAME="$(git --git-dir="$GIT_DIR_PATH" --work-tree="$ROOT_DIR" branch --show-current 2>/dev/null || true)"
  COMMIT_SHA="$(git --git-dir="$GIT_DIR_PATH" --work-tree="$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || true)"
else
  BRANCH_NAME="$(git branch --show-current 2>/dev/null || true)"
  COMMIT_SHA="$(git rev-parse --short HEAD 2>/dev/null || true)"
fi

if [[ -z "$BRANCH_NAME" ]]; then
  BRANCH_NAME="unknown"
fi

if [[ -z "$COMMIT_SHA" ]]; then
  COMMIT_SHA="unknown"
fi

if [[ -f "$LOG_FILE" ]]; then
  printf "%s\n" "$LOG_FILE"
  exit 0
fi

cat >"$LOG_FILE" <<EOF
# Orchestration Run ${RUN_ID}

## Context
- ownership_model: Claude drafts, Codex applies
- repository: $(basename "$ROOT_DIR")
- branch: ${BRANCH_NAME}
- commit: ${COMMIT_SHA}
- created_at_utc: $(date -u +%Y-%m-%dT%H:%M:%SZ)
- internal_agent_language: English
- user_facing_language: Russian

## Task
- request: ${TASK}
- success_criteria:

## Claude Draft
- plan:
- risks:
- proposed_diff_summary:

## Codex Apply
- implementation_notes:
- files_touched:
- decisions_made:

## Validation
- targeted_checks:
- results:
- blockers:

## Handoff
- next_actor:
- next_action:
- done_definition:

## Memory Update Checklist
- [ ] .codex/state.md
- [ ] .codex/todo.md (if needed)
- [ ] .codex/done.md (if needed)
- [ ] .codex/time_runtime.json (if needed)
- [ ] .codex/tg_control.jsonl (if needed)
EOF

printf "%s\n" "$LOG_FILE"
