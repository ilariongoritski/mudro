# Casino DB — миграции и эволюция схемы

## 📋 Хронология версий

### Версия 012 (главная БД)
**Файл**: `migrations/012_casino.sql`  
**Статус**: ✅ Production  
**Дата**: early 2024  

Содержит:
- `casino_accounts` — кошельки (NUMERIC для точности)
- `casino_transfers` + `casino_ledger_entries` — double-entry ledger
- `casino_rounds` — provably fair раунды
- `casino_rtp_profiles` + `casino_rtp_assignments` — управление честностью
- `casino_idempotency` — защита от дублей

**Индексы**: 5 штук, все активные

---

### Версия 001 (casino microservice)
**Файл**: `services/casino/migrations/001_init.sql`  
**Статус**: ✅ Legacy  

Содержит:
- `casino_config` — singleton конфиг (рTP, начальный баланс)
- `casino_participants` — старая таблица игроков
- `casino_spins` — история спинов слотов (BIGINT)

**Особенность**: Синглтон паттерн для конфига
```sql
CREATE TABLE casino_config (
  id boolean primary key default true,  -- только одна строка!
  ...
)
```

---

### Версия 002 (live roulette)
**Файл**: `services/casino/migrations/002_live_roulette.sql`  
**Статус**: ✅ Current  
**Дата**: 2024-2025  

**Добавляет**:
- `casino_players` — новая таблица игроков (денормализация из `casino_participants`)
- `casino_game_activity` — лог всех игр (slots, roulette, plinko, crash)
- `casino_roulette_rounds` — раунды рулетки с фазами
- `casino_roulette_bets` — ставки в раундах рулетки

**Миграция данных**:
```sql
-- Миграция участников в новую таблицу игроков
INSERT INTO casino_players (user_id, username, ..., level, xp_progress)
SELECT 
  p.user_id,
  p.username,
  ...,
  greatest(1, 1 + (coalesce(ss.total_wagered, 0) / 1000)) as level,
  coalesce(ss.total_wagered, 0) % 1000 as xp_progress
FROM casino_participants p
LEFT JOIN spin_stats ss ON ss.user_id = p.user_id
ON CONFLICT (user_id) DO UPDATE SET ...
```

---

## 🔄 Соотношение таблиц по версиям

| Таблица | 012 Main | 001 Legacy | 002 Current | Назначение |
|---------|----------|-----------|-------------|-----------|
| `casino_config` | ❌ | ✅ | ✅ | Конфиг казино (синглтон) |
| `casino_participants` | ❌ | ✅ | 📦 (deprecated) | Игроки V1 |
| `casino_spins` | ❌ | ✅ | ✅ | История слотов |
| `casino_accounts` | ✅ | ❌ | ❌ | Кошельки (main DB) |
| `casino_transfers` | ✅ | ❌ | ❌ | Транзакции (main DB) |
| `casino_ledger_entries` | ✅ | ❌ | ❌ | Бухгалтерия (main DB) |
| `casino_rounds` | ✅ | ❌ | ❌ | Раунды (main DB) |
| `casino_rtp_*` | ✅ | ❌ | ❌ | RTP управление (main DB) |
| `casino_idempotency` | ✅ | ❌ | ❌ | Дубли (main DB) |
| `casino_players` | ❌ | ❌ | ✅ | Игроки V2 с gamification |
| `casino_game_activity` | ❌ | ❌ | ✅ | Лог всех игр |
| `casino_roulette_rounds` | ❌ | ❌ | ✅ | Раунды рулетки |
| `casino_roulette_bets` | ❌ | ❌ | ✅ | Ставки рулетки |

---

## 🎯 Двойная архитектура

### Почему два места?

```
┌─────────────────────────────────────────────────────┐
│  ТРЕБОВАНИЕ: Казино должно быть НЕЗАВИСИМЫМ        │
│  микросервисом и отдельной БД                       │
└─────────────────────────────────────────────────────┘

✅ Решение: Гибридная архитектура

  MUDRO MAIN DB (migrations/012_casino.sql)
  ├─ Авторитетный источник для финансов
  ├─ Двойная бухгалтерия (неопровержимый аудит)
  ├─ Управление честностью (RTP)
  └─ Защита от дублей (idempotency)

  CASINO MICROSERVICE DB (services/casino/)
  ├─ Независимая БД для live-игр
  ├─ Быстрое масштабирование (BIGINT вместо UUID)
  ├─ Собственные индексы для рулетки
  └─ Локальное состояние раунда
```

### Синхронизация между БД

❓ **На данный момент явной синхронизации нет в миграциях**

Предполагается:
1. **Основная БД** — truth source для финансов
2. **Микросервис** — синхронизирует через:
   - API вызовы к основной БД
   - Периодическое обновление `casino_players.balance`
   - Event streaming (возможно, в будущем)

---

## 📊 Тип данных: NUMERIC vs BIGINT

### NUMERIC (в 012 main)

```sql
-- Точность: NUMERIC(30, 10) = 30 цифр, 10 после запятой
balance NUMERIC(30,10) NOT NULL DEFAULT 0

-- 999999999999999999999.9999999999
-- Точность: микротранзакции, криптовалюта
-- Скорость: медленнее на ~3x
-- Индексирование: полная поддержка
```

### BIGINT (в microservice)

```sql
-- Целые числа: -9223372036854775808 до 9223372036854775807
balance BIGINT NOT NULL DEFAULT 500

-- Точность: только целые числа (нет центов)
-- Скорость: быстро, нативный тип
-- Индексирование: отлично
-- Риск: переполнение при balance > 10^18
```

**Выбор**:
- `NUMERIC` в 012: нужна точность для бухгалтерии
- `BIGINT` в microservice: масштабируемость важнее

---

## 🔐 Безопасность: CASCADE vs RESTRICT

### CASCADE (удаление пользователя удаляет игры)

```sql
-- Если удалить пользователя:
user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

-- Автоматически удалятся:
-- - casino_rounds WHERE user_id = $1
-- - casino_rtp_assignments WHERE user_id = $1
-- - casino_idempotency WHERE user_id = $1
```

**Последствия**:
✅ Простое удаление учётной записи  
❌ Историческая информация теряется  

### RESTRICT (нельзя удалить счет с историей)

```sql
-- Если есть записи в ledger:
account_id UUID NOT NULL REFERENCES casino_accounts(id) ON DELETE RESTRICT,

-- Попытка удалить casino_accounts вызовет ошибку:
-- ERROR: update or delete on table "casino_accounts" violates 
-- foreign key constraint "casino_ledger_entries_account_id_fkey"
```

**Последствия**:
✅ История неудаляема  
❌ Нельзя полностью удалить счет  

---

## 🚀 Типовые операции

### 1. Новый пользователь регистрируется

```sql
-- В 012 (main DB):
INSERT INTO casino_accounts (user_id, type, code, balance)
VALUES (NEW_USER_ID, 'user', 'USER_' || NEW_USER_ID, 0);

-- В microservice (async, через API):
INSERT INTO casino_players (user_id, username, balance, level)
SELECT id, username, 0, 1 FROM users WHERE id = NEW_USER_ID;
```

### 2. Игрок ставит деньги

```sql
-- В 012 (двойная бухгалтерия):

1. INSERT INTO casino_transfers (kind) VALUES ('bet_stake')
   → transfer_id

2. INSERT INTO casino_ledger_entries (transfer_id, account_id, direction, amount)
   VALUES (transfer_id, user_account, 'debit', 100)
   
3. INSERT INTO casino_ledger_entries (transfer_id, account_id, direction, amount)
   VALUES (transfer_id, house_account, 'credit', 100)

4. UPDATE casino_accounts SET balance = balance - 100 WHERE user_id = $1
```

### 3. Раунд разрешается (win/loss)

```sql
-- В 012:
UPDATE casino_rounds 
SET status = 'resolved',
    roll = computed_roll,
    payout_amount = computed_payout,
    multiplier = computed_multiplier,
    resolved_at = NOW()
WHERE id = $1;

-- Затем аналогично ставке: INSERT в ledger

-- В microservice (async):
UPDATE casino_players 
SET balance = balance + net_pnl,
    total_wagered = total_wagered + bet,
    total_won = total_won + payout,
    games_played = games_played + 1
WHERE user_id = $1;

INSERT INTO casino_game_activity (user_id, game_type, ...)
VALUES ($1, 'slots', ...);
```

---

## 📈 Прогноз миграций (future)

### V003: Ускорение (оптимизация индексов)

```sql
-- Добавить индекс для курсор-пагинации
CREATE INDEX idx_casino_rounds_cursor
  ON casino_rounds(published_at DESC, id DESC)
  
-- GIN для JSONB (paytable)
CREATE INDEX idx_casino_rtp_profiles_paytable
  ON casino_rtp_profiles USING GIN (paytable)
```

### V004: Кеш слоёв

```sql
-- Кеш популярных раундов (в Redis / обновляемо)
CREATE MATERIALIZED VIEW casino_rounds_cache AS
SELECT 
  user_id,
  COUNT(*) as recent_rounds,
  AVG(multiplier) as avg_multiplier
FROM casino_rounds
WHERE created_at > NOW() - INTERVAL '1 day'
GROUP BY user_id;

CREATE INDEX mv_cache_user ON casino_rounds_cache(user_id);
```

### V005: Партиционирование (для масштабирования)

```sql
-- Для таблиц с историей > 1GB
CREATE TABLE casino_rounds_2024 PARTITION OF casino_rounds
  FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
  
CREATE TABLE casino_rounds_2025 PARTITION OF casino_rounds
  FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
```

---

## 🧪 Тестирование схемы

### Проверка целостности после миграции

```sql
-- 1. Все accounts должны иметь ledger
SELECT COUNT(*) FROM casino_accounts ca
WHERE NOT EXISTS (
  SELECT 1 FROM casino_ledger_entries cle 
  WHERE cle.account_id = ca.id
);
-- Expected: 0 (или только новые счета)

-- 2. Сумма дебитов = сумма кредитов
SELECT ABS(
  COALESCE(SUM(amount) FILTER (WHERE direction = 'debit'), 0) - 
  COALESCE(SUM(amount) FILTER (WHERE direction = 'credit'), 0)
) FROM casino_ledger_entries;
-- Expected: 0

-- 3. Провабли-фэйр целостность
SELECT COUNT(*) FROM casino_rounds 
WHERE status = 'resolved' 
  AND (server_seed IS NULL OR client_seed IS NULL OR round_hash IS NULL);
-- Expected: 0
```

---

## 📝 Чеклист обновления схемы

При добавлении новой миграции:

- [ ] Прочитать `000X.up.sql` и `000X.down.sql`
- [ ] Проверить индексы: `CREATE INDEX CONCURRENTLY`
- [ ] Проверить FK: CASCADE vs RESTRICT vs SET NULL
- [ ] Проверить типы: NUMERIC/BIGINT/TEXT
- [ ] Добавить initial data (INSERT с ON CONFLICT)
- [ ] Тестировать rollback: `psql < 000X.down.sql`
- [ ] Проверить давление на индексы
- [ ] Документировать в `docs/CASINO_DB_ARCHITECTURE.md`
- [ ] Обновить это руководство

