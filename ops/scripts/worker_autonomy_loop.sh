#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

RUN_ID="$(date +%Y%m%d-%H%M)"
RUN_DIR=".codex/logs/${RUN_ID}"
LOG_FILE="${RUN_DIR}/index.md"
STATE_FILE=".codex/state.md"

mkdir -p "$RUN_DIR"

log() {
  printf "%s\n" "$*" | tee -a "$LOG_FILE"
}

run_step() {
  local title="$1"
  local cmd="$2"
  local attempts=0
  local max_attempts=2

  log ""
  log "### ${title}"
  while (( attempts < max_attempts )); do
    attempts=$((attempts + 1))
    log "- попытка ${attempts}/${max_attempts}: ${cmd}"
    if bash -lc "$cmd" >>"$LOG_FILE" 2>&1; then
      log "- статус: ok"
      return 0
    fi
    log "- статус: ошибка"
  done

  return 1
}

cat >"$LOG_FILE" <<LOGHDR
# Прогон ${RUN_ID}

## 1) Окружение
- pwd: $(pwd)
- branch: $(git branch --show-current)
- commit: $(git rev-parse --short HEAD)
- date: $(date -Iseconds)

## 2) Команды и ключевой вывод
LOGHDR

failed_step=""
failed_cmd=""

if ! run_step "Шаг 1: поднять сервисы" "make up"; then
  failed_step="make up"
  failed_cmd="make up"
fi

if [[ -z "$failed_step" ]] && ! run_step "Шаг 2: docker compose ps" "docker compose ps"; then
  failed_step="docker compose ps"
  failed_cmd="docker compose ps"
fi

if [[ -z "$failed_step" ]] && ! run_step "Шаг 3: dbcheck" "make dbcheck"; then
  failed_step="make dbcheck"
  failed_cmd="make dbcheck"
fi

if [[ -z "$failed_step" ]] && ! run_step "Шаг 4: migrate" "make migrate"; then
  failed_step="make migrate"
  failed_cmd="make migrate"
fi

if [[ -z "$failed_step" ]] && ! run_step "Шаг 5: tables" "make tables"; then
  failed_step="make tables"
  failed_cmd="make tables"
fi

if [[ -z "$failed_step" ]] && ! run_step "Шаг 6: test" "make test"; then
  failed_step="make test"
  failed_cmd="make test"
fi

if [[ -z "$failed_step" ]] && ! run_step "Шаг 7: sanity count" "psql \"${DSN:-postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable}\" -X -c \"select count(*) from posts;\""; then
  failed_step="sanity"
  failed_cmd="psql \"${DSN:-postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable}\" -X -c \"select count(*) from posts;\""
fi

log ""
log "## 3) Ретраи и решения"
if [[ -n "$failed_step" ]]; then
  log "- Прогон остановлен: шаг ${failed_step} упал после 2 попыток."
  log "- Требуется вопрос человеку с точной ошибкой и ссылкой на лог: ${RUN_DIR}/index.md."
else
  log "- Все шаги health loop завершились успешно."
fi

log ""
log "## 4) Изменения в репе (diff кратко)"
log "- Этот скрипт не меняет код проекта; только пишет лог прогона."

status_time="$(date -Iseconds)"
{
  echo ""
  echo "- Дата/время: ${status_time}"
  echo "- Что запускал: scripts/worker_autonomy_loop.sh (полный health loop с авто-логом и ретраями)"
  if [[ -n "$failed_step" ]]; then
    echo "- Что прошло: создан лог прогона в ${RUN_DIR}/index.md; успешные шаги до падения зафиксированы в логе."
    echo "- Что упало (ошибка 5–15 строк): см. ${RUN_DIR}/index.md (последний упавший шаг: ${failed_cmd})."
    echo "- Что починил (если было): выполнены только безопасные повторы команд (до 2 попыток на шаг)."
    echo "- Следующий шаг: задать человеку вопрос с точной командой/ошибкой и приложить путь ${RUN_DIR}/index.md."
  else
    echo "- Что прошло: health loop пройден полностью; лог в ${RUN_DIR}/index.md."
    echo "- Что упало (ошибка 5–15 строк): не падало"
    echo "- Что починил (если было): не требовалось"
    echo "- Следующий шаг: перейти к задачам из .codex/todo.md"
  fi
} >> "$STATE_FILE"

if [[ -n "$failed_step" ]]; then
  exit 1
fi
