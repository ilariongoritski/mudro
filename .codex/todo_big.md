# TODO BIG (стратегические цели)

Горизонты:
- H1: 1-2 недели
- H2: 1-2 месяца
- H3: 1-2 квартала

## Epic A: Стабильный прод-контур (H1)
- [ ] A1 | Единый `docker-compose.prod.yml` (`api/db/agent/reporter/reverse-proxy`)
- [ ] A2 | Healthchecks + restart-policy + минимальные лимиты ресурсов
- [ ] A3 | Runbook: cold start, rollback, incident response
- [ ] A4 | Авто-бэкап Postgres (daily + retention 7/30)
- [ ] A5 | Базовый hardening VPS (fail2ban, UFW, ssh hardening)

## Epic B: Автономный агент v1 (H1-H2)
- [ ] B1 | Очередь задач `agent_queue` (сделано частично)
- [ ] B2 | Planner: автогенерация задач из TODO/find/health-fail
- [ ] B3 | Worker: safe-only pipeline (patch + test + log)
- [ ] B4 | Review Gate: обязательный approve для risky-задач
- [ ] B5 | Reporter: digest выполнений/падений/стоимости
- [ ] B6 | SLO агента: success-rate, regression-rate, MTTR

## Epic C: Качество и предсказуемость поставки (H2)
- [ ] C1 | Nightly pipeline (`make health`, `go test ./...`, security checks)
- [ ] C2 | Стандартизованный формат изменений и авто-лог в `.codex/logs/*`
- [ ] C3 | Тесты на критичные сценарии bot/api/agent
- [ ] C4 | Миграционный контроль: dry-run + preflight checks

## Epic D: Продуктовый backend до frontend-ready (H2)
- [ ] D1 | Зафиксировать контракт `/api/front` v1 + версионирование
- [ ] D2 | Ввести auth/roles (MVP)
- [ ] D3 | Улучшить feed-ранжирование/поиск/фильтры
- [ ] D4 | Подготовить модель feature flags

## Epic E: Контур роста (H3)
- [ ] E1 | Reverse proxy + TLS + rate-limit + basic WAF rules
- [ ] E2 | Redis для очередей/кеша состояний агента
- [ ] E3 | Объектное хранилище (MinIO/S3) для медиа/бэкапов
- [ ] E4 | Multi-tenant readiness (данные/изоляция/лимиты)

## Epic F: Экономика и эффективность (H2-H3)
- [ ] F1 | Учет стоимости LLM-токенов на задачу/день/неделю
- [ ] F2 | KPI: lead time idea->commit, доля автоуспехов, MTTR
- [ ] F3 | Лимиты бюджета на автозадачи (timebox/tokenbox/day)

## Definition of Done для крупных задач
- [ ] Есть команда запуска/проверки
- [ ] Есть запись в `.codex/logs/<run>/index.md`
- [ ] Есть обновление `.codex/state.md`
- [ ] Есть заметка в `.codex/done.md`
- [ ] Не нарушены правила безопасности AGENTS.md
