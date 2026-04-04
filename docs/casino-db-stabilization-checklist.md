# Casino DB Stabilization Checklist

Этот чеклист закрывает rollout для уже внесённых изменений по:
- `casino_players` как primary projection в microservice DB
- `casino_participants_v2` как compatibility view
- balance sync queue + reconciler
- `casino_accounts_audit`
- `casino_rtp_tiers` + dual-read fallback
- roulette session cache/janitor

## 1. Preconditions

- [ ] Рабочий shell-контур: WSL или Git Bash через [run-mudro-no-profile.ps1](E:/mudr/mudro11/scripts/windows/run-mudro-no-profile.ps1)
- [ ] Доступны `docker`, `psql`, `go`, `git`
- [ ] Подняты main DB и casino DB
- [ ] Понимаешь, куда смотрят DSN:
  - main DB: `DSN`
  - casino microservice DB: `CASINO_DSN`
  - main wallet lookup from casino service: `CASINO_MAIN_DSN`

## 2. Env Wiring

- [ ] В локальном/прод-контуре задан `CASINO_MAIN_DSN`
- [ ] `CASINO_MAIN_DSN` указывает на main Mudro DB, а не на casino DB
- [ ] `CASINO_DSN` указывает только на casino microservice DB
- [ ] Для Supabase/managed Postgres используется `sslmode=require`, если это внешний контур

Минимальная схема:

```env
DSN=postgres://.../gallery?sslmode=disable
CASINO_DSN=postgres://.../mudro_casino?sslmode=disable
CASINO_MAIN_DSN=${DSN}
```

## 3. Main DB Migrations

- [ ] Применить main casino migrations:
  - `migrations/012_casino.sql`
  - `migrations/017_casino_emoji_v2.sql`
  - `migrations/018_casino_stabilization.sql`
- [ ] Проверить, что создались:
  - `casino_accounts_audit`
  - `casino_rtp_tiers`
  - trigger `tr_casino_accounts_audit`

Команда:

```bash
bash ./scripts/migrate-casino-main.sh
```

## 4. Casino Microservice Migrations

- [ ] Применить casino microservice migrations
- [ ] Проверить, что создались:
  - `casino_participants_v2`
  - `casino_balance_sync_queue`
  - `casino_roulette_sessions`
  - индекс `casino_game_activity_user_game_created_idx`

Команда:

```bash
bash ./scripts/migrate-casino.sh
```

## 5. Code Verification

- [ ] `go test ./services/casino/...`
- [ ] `(опционально) go test ./...` если нужен repo-wide smoke
- [ ] Нет compile/runtime-конфликта по `CASINO_MAIN_DSN`
- [ ] Нет прямых write-path в `casino_participants`

## 6. Runtime Checks

- [ ] casino service стартует
- [ ] `GET http://127.0.0.1:8082/healthz` отвечает `ok`
- [ ] новый пользователь bootstrap'ится в `casino_players`
- [ ] при spin/plinko/roulette-bet создаётся запись в `casino_balance_sync_queue`
- [ ] reconciler переводит queue item в `done`
- [ ] `casino_players.balance` подтягивается к main wallet после reconcile

## 7. Data Checks

- [ ] drift check между main wallet и micro projection
- [ ] view `casino_participants_v2` читается без ошибок
- [ ] RTP читает tiers, а при отсутствии таблицы не падает и идёт через JSON fallback
- [ ] audit rows создаются только при реальном изменении `casino_accounts.balance`

Рекомендуемые SQL-проверки:

```sql
select * from casino_balance_sync_queue order by created_at desc limit 20;
```

```sql
select user_id, balance from casino_players order by updated_at desc limit 20;
```

```sql
select account_id, old_balance, new_balance, reason, changed_by, changed_at
from casino_accounts_audit
order by changed_at desc
limit 20;
```

```sql
select rtp_profile_id, min_roll, max_roll, multiplier, label, symbol
from casino_rtp_tiers
order by rtp_profile_id, min_roll;
```

## 8. Git / Fixup / Deploy

- [ ] `git status --short`
- [ ] при необходимости дофиксить compile/test issues
- [ ] commit
- [ ] push
- [ ] redeploy runtime

## 9. One-Command Launcher

Для Windows:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\casino-db-rollout.ps1 -PrintOnly
```

Repo-wide smoke:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\casino-db-rollout.ps1 -FullSmoke
```

Реальный прогон:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\casino-db-rollout.ps1
```

## 10. Acceptance

- [ ] `casino_players` стал основной projection-моделью
- [ ] legacy `casino_participants` больше не используется как write target
- [ ] balance reconcile идёт в пользу main wallet
- [ ] roulette session cache живёт отдельно от persistent round/bet tables
- [ ] RTP tiers работают как primary read path без жёсткой зависимости на порядок rollout
