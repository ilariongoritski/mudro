#!/usr/bin/env bash
set -euo pipefail

fail=0

check_list() {
  local label="$1"
  shift

  local output
  output="$("$@")"

  local listed_down
  listed_down="$(printf '%s\n' "${output}" | grep -E '\.down\.sql$' || true)"
  if [ -n "${listed_down}" ]; then
    echo "${label} up-list contains rollback migrations:" >&2
    echo "${listed_down}" >&2
    fail=1
    return
  fi

  echo "${label} up-list: ok"
  printf '%s\n' "${output}"
}

main_list() {
  if [ -n "${MAIN_UP_MIGRATIONS:-}" ]; then
    # MAIN_UP_MIGRATIONS comes from Makefile's UP_MIGRATIONS, so this verifies
    # the exact list used by migrate-all without recursively invoking make.
    printf '%s\n' ${MAIN_UP_MIGRATIONS}
    return
  fi

  find "${MIGRATIONS_DIR:-migrations}" -maxdepth 1 -type f -name "*.sql" ! -name "*.down.sql" -print | sort
}

casino_list() {
  "${BASH:-bash}" ./scripts/migrate-casino.sh --list
}

check_list "main" main_list
check_list "casino" casino_list

exit "${fail}"
