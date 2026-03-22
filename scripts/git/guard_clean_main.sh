#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
branch="$(git -C "$repo_root" rev-parse --abbrev-ref HEAD 2>/dev/null || true)"

if [[ "$branch" != "main" && "$branch" != "master" ]]; then
  exit 0
fi

if git -C "$repo_root" status --porcelain --untracked-files=all | grep -q .; then
  printf 'Refusing to continue: worktree is dirty on protected branch %s.\n' "$branch" >&2
  printf 'Commit, stash, or discard changes before retrying.\n' >&2
  git -C "$repo_root" status --short >&2
  exit 1
fi

exit 0
