# VPS Worker Plane Rollout Summary — 2026-03-23

Этот отчёт агрегирует итоги текущего orchestration thread по `mudro`, `OpenClaw`, `Skaro` и tracked `systemd` rollout.

## Scope

- перевести `mudro-api`, `mudro-bot`, `mudro-agent` на tracked `systemd` units;
- поднять `openclaw.service` и `skaro.service` на VPS;
- зафиксировать live-fixes обратно в репозиторий;
- подготовить следующий шаг: tracked `claudeusageproxy.service` для полного token accounting.

## Control Plane

- `Codex` — канонический исполнитель, diff owner, VPS rollout и финальная проверка;
- внешние `Opus`-прогоны — только reviewer/checklist loop;
- локальные субагенты `Franklin` и `Zeno` — preflight-checklists без прямого редактирования.

## Subagent Inputs

### Franklin

Проверил core-runtime prerequisites для `mudro-api`, `mudro-bot`, `mudro-agent`.

Подтвердил основные обязательные env-поля:

- `JWT_SECRET`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_ALLOWED_USERNAME`
- рабочий non-superuser `DSN`
- валидный `MEDIA_ROOT`

### Zeno

Проверил readiness worker-plane для `OpenClaw` / `Skaro`.

Подтвердил:

- tracked installer требует уже разложенный `/opt/mudro/app`;
- `Skaro` на VPS нужен отдельный Linux launcher, а не PowerShell-only локальный flow;
- для корректного accounting нужен отдельный proxy-service, а не только `MUDRO_CLAUDE_PROXY_URL` в env.

### External Opus reviewer loops

Использовались run-id:

- `20260323-systemd-template`
- `20260323-systemd-worker-plane`
- `20260323-systemd-final-review`
- `20260323-vps-live-rollout`

Практически приняты только те выводы, которые подтвердились локально или на VPS.

## Applied Repository Commits

- `a916100` — tracked `mudro-api` runtime
- `52f0622` — tracked hardened runtimes for `mudro-bot`, `mudro-agent`, `OpenClaw`, `Skaro`
- `370ee06` — repo-side fix for tracked `OpenClaw` rollout

## VPS Rollout Outcome

Core services:

- `mudro-api.service` — active
- `mudro-bot.service` — active
- `mudro-agent-worker.service` — active
- `mudro-agent-planner.timer` — active

Worker plane:

- `openclaw.service` — active
- `skaro.service` — active

## Live Checks

Подтверждённые smoke-checks:

- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/api/front?limit=1`
- `http://127.0.0.1:4700/`
- `http://127.0.0.1:18789/`
- `https://frontend-psi-ten-33.vercel.app/api/front?limit=1`
- `https://frontend-psi-ten-33.vercel.app/media/...`

Итог:

- API жив;
- frontend получает feed;
- media GET на внешнем frontend работает;
- `Skaro` dashboard поднимается;
- tracked root-level `OpenClaw` gateway поднят.

## Drift Fixed During Rollout

- старый VPS runtime использовал неtracked `DSN` override;
- shell wrappers после deploy требовали явный `chmod 755`;
- root-level `openclaw.service` пришлось ослабить относительно избыточного sandbox-hardening, иначе `Node/V8` падал на старте.

Именно это затем было возвращено в репозиторий коммитом `370ee06`.

## Remaining Gaps

1. `claudeusageproxy.service` ещё не был tracked в `main` на момент базового rollout.
2. `Skaro` accounting на VPS пока зависит от наличия proxy по `127.0.0.1:8788`.
3. В локальном дереве был замечен крупный незакоммиченный diff в `README.md`, который выглядит как noisy rewrite и должен быть либо явно утверждён, либо отброшен отдельно от runtime-работ.

## Decision

Следующий безопасный шаг:

1. добавить tracked `claudeusageproxy.service` и его env-контракт;
2. переключить `OpenClaw` и `Skaro` на локальный proxy-base-url для полного token accounting;
3. отдельно разобрать `README.md` вне runtime-ветки.
