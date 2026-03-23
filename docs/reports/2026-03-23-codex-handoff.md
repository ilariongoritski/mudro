# MUDRO Codex Handoff

Дата: 2026-03-23  
Назначение: sanitzed handoff-репорт для переноса работы в новый проект Codex App без зависимости от текущего чата.

## 1. Контекст проекта

- Проект: `MUDRO`
- Локальный путь: `D:\mudr\mudro11`
- Канонический remote: `origin`
- Базовая ветка: `main`
- Последний гарантированно запушенный коммит по линии `VPS/systemd/OpenClaw`: `370ee06` `Fix tracked OpenClaw systemd rollout`
- Назначение отчёта: передать в новый проект Codex то, что уже доведено до рабочего состояния, что подготовлено локально, и какой узкий следующий шаг нужен без повторного анализа

## 2. Что реально доведено до рабочего состояния

### Core MUDRO runtime на VPS

Через tracked `systemd`-контур уже были подняты и проверены:

- `mudro-api.service`
- `mudro-bot.service`
- `mudro-agent-worker.service`
- `mudro-agent-planner.timer`

Фактический результат:

- `mudro-api` отвечает на `http://127.0.0.1:8080/healthz`
- `mudro-api` отдаёт `api/front`
- `mudro-bot` успешно стартует под tracked unit
- `mudro-agent-worker` активен
- `mudro-agent-planner.timer` активен и planner выполнялся как минимум один раз

### Worker plane на VPS

Через tracked `systemd`-контур уже были подняты и проверены:

- `openclaw.service`
- `skaro.service`

Фактический результат:

- `OpenClaw gateway` отвечал на `http://127.0.0.1:18789/`
- `Skaro` отвечал на `http://127.0.0.1:4700/`

### Frontend / site

- Vercel frontend получает данные с VPS API
- media GET через сайт работает
- HEAD на media через Vercel может возвращать misleading metadata и не считается блокером, если GET отдаёт тело

### Что уже зафиксировано в `main`

В `main` уже находится tracked `systemd`-контур для:

- `mudro-api`
- `mudro-bot`
- `mudro-agent`
- `openclaw`
- `skaro`

## 3. Что подготовлено локально, но ещё не зафиксировано в GitHub

Ниже перечислено как `prepared local diff, not yet committed/applied`.

### Tracked `claudeusageproxy`-контур

Подготовлены файлы:

- `ops/systemd/claudeusageproxy.service`
- `ops/systemd/claudeusageproxy.env.example`
- `scripts/openclaw/claudeusageproxy_systemd.sh`

Назначение:

- ввести tracked `claudeusageproxy.service` на VPS
- перевести `Skaro/OpenClaw` на локальный proxy для централизованного учёта токенов
- хранить реальный upstream Claude key только во внешнем runtime env-файле

### Обновления worker-plane wiring

Подготовлены изменения в:

- `ops/systemd/openclaw.env.example`
- `ops/systemd/skaro.env.example`
- `ops/systemd/openclaw.service`
- `ops/systemd/skaro.service`
- `ops/scripts/install_openclaw_systemd.sh`
- `ops/systemd/README.md`

Назначение:

- перевести `OpenClaw` и `Skaro` на `MUDRO_CLAUDE_PROXY_URL=http://127.0.0.1:8788`
- сделать proxy-accounting воспроизводимым и tracked

### Санитизированные отчёты, уже собранные локально

Подтверждённые локально подготовленные файлы:

- `docs/reports/2026-03-23-vps-rollout-and-orchestration.md`
- `docs/reports/2026-03-23-vps-rollout-and-orchestration.md` следует считать рабочим отчётом по rollout

Возможные альтернативные rollup-файлы, если они есть в дереве, считать только кандидатами на консолидацию, а не каноном:

- `docs/reports/2026-03-23-orchestration-rollup.md`
- `docs/reports/2026-03-23-vps-worker-plane-rollout.md`

## 4. Что сознательно не переносить

Не переносить в новый проект как каноническое состояние:

- raw выгрузки из `D:\mudr\_mudro-local\claude-orch\runs\...`
- runtime ledger/log/state файлы
- любые токены, ключи, пароли, `DSN` с секретами
- англоязычный rewrite `README.md`
- временные или generated артефакты `Skaro/OpenClaw`
- любые ad-hoc server drift artifacts, которые не отражены в tracked repo-файлах

## 5. Решения, которые уже приняты

### README

- англоязычный rewrite `README.md` считать шумом
- его нельзя тащить в `main`
- каноническим остаётся существующий operational README в репозитории

### Worker plane accounting

- для `Skaro/OpenClaw` нужен tracked `claudeusageproxy.service`
- учёт токенов должен быть локальным и воспроизводимым на VPS
- proxy должен жить как отдельный tracked `systemd`-unit, а не как ручной непроверяемый runtime

### Работа с секретами

Секреты должны жить только вне репозитория:

- `/etc/mudro/runtime/*.env`
- `/etc/openclaw/runtime/*.env`

Ни один реальный ключ, пароль или токен не должен попадать:

- в `docs/reports/*`
- в tracked `.env.example`
- в rollout-скрипты

### Архитектурное разделение

- `OpenClaw` и `Skaro` остаются отдельным worker-plane
- они не должны смешиваться с core MUDRO runtime
- core runtime и worker-plane должны оставаться отдельно tracked и отдельно конфигурируемыми

## 6. Подтверждённые live-интерфейсы и runtime-контракты

### Подтверждённые live endpoints

- `mudro-api`: `http://127.0.0.1:8080/healthz`
- `Skaro`: `http://127.0.0.1:4700/`
- `OpenClaw gateway`: `http://127.0.0.1:18789/`

### Runtime env contracts

Уже используются:

- `/etc/mudro/runtime/mudro-api.env`
- `/etc/mudro/runtime/mudro-bot.env`
- `/etc/mudro/runtime/mudro-agent.env`
- `/etc/openclaw/runtime/openclaw.env`
- `/etc/openclaw/runtime/skaro.env`

Целевой новый контракт:

- `/etc/openclaw/runtime/claudeusageproxy.env`

### Целевой ledger path для полного accounting

- `/var/lib/openclaw/claude-orch/ledger/usage_log.jsonl`
- `/var/lib/openclaw/claude-orch/ledger/token_usage.yaml`
- `/var/lib/openclaw/claude-orch/ledger/role_usage.yaml`

## 7. Основной технический блокер текущего чата

Текущий тред упёрся в хостовую ошибку Codex shell:

```text
Internal Windows PowerShell error. Loading managed Windows PowerShell failed with error 8009001d.
```

Из-за этого в текущем чате не были завершены:

- `git status / commit / push`
- локальная валидация подготовленного proxy-diff
- VPS apply для `claudeusageproxy.service`
- финальный live smoke-check accounting-слоя

Важно: это блокер среды текущего треда, а не блокер кода репозитория.

## 8. Что подтвердили субагенты и проверки

### Уже подтверждённые выводы

- tracked `mudro` stack на VPS уже был успешно применён
- `openclaw.service` и `skaro.service` уже были подняты и проверены
- следующий узкий шаг — не новый рефакторинг, а именно tracked `claudeusageproxy.service`
- шумовой rewrite `README.md` не должен попадать в `main`

### Как трактовать субагентные артефакты

- использовать только санитизированные выводы и решения
- raw сообщения субагентов не считать каноном
- архитектурной точкой истины остаются `main` и фактическое VPS runtime state

## 9. Следующий узкий план для нового проекта

Новый проект должен стартовать не с широкого переосмысления архитектуры, а с этого узкого списка:

1. восстановить рабочий shell
2. проверить рабочее дерево в `D:\mudr\mudro11`
3. провалидировать tracked `claudeusageproxy` diff
4. закоммитить и запушить его в `main`
5. применить `claudeusageproxy.service` на VPS
6. перезапустить `claudeusageproxy`, `openclaw`, `skaro`
7. проверить ledger-файлы и `healthz`
8. только после этого запускать новый полный `Opus` code review

## 10. Validation checklist для нового проекта

### Локально

```bash
git -C D:\mudr\mudro11 status -sb
make validate-systemd-templates
```

### На VPS до proxy rollout

```bash
systemctl is-active mudro-api mudro-bot mudro-agent-worker mudro-agent-planner.timer openclaw skaro
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:4700/
```

### На VPS после proxy rollout

```bash
systemctl is-active claudeusageproxy openclaw skaro
curl http://127.0.0.1:8788/healthz
```

Проверить наличие и обновление:

- `/var/lib/openclaw/claude-orch/ledger/usage_log.jsonl`
- `/var/lib/openclaw/claude-orch/ledger/token_usage.yaml`
- `/var/lib/openclaw/claude-orch/ledger/role_usage.yaml`

## 11. Базовые допущения

- Формат этого файла: handoff-репорт, а не полный хронологический журнал
- Отчёт санитизирован: без секретов, токенов и runtime-dumps
- Каноническая точка истины: `main` + VPS runtime state
- Все prepared-but-uncommitted изменения описываются явно как локально подготовленные, но не зафиксированные
- До восстановления shell не следует притворяться, что `commit/push` и VPS apply уже завершены
