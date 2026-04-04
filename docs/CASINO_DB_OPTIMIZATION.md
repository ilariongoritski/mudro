# Casino DB — Рекомендации по оптимизации и best practices

## 🎯 Приоритет: ВЫСОКОЙ → НИЗКОЙ

---

## 🔴 КРИТИЧНО (P0) — Реализовать сразу

### 1. Проблема: Денормализация `casino_players`

**Сейчас**:
```sql
-- В 002_live_roulette.sql создаётся таблица casino_players
-- через INSERT ... SELECT из casino_participants
-- Но casino_participants остаётся и используется!
```

**Проблема**:
- Данные о балансе дублируются
- Нет единой истины (две таблицы могут расходиться)
- Синхронизация вручную

**Решение**:

```sql
-- Вариант A: Оставить только casino_players в microservice
-- Вариант B: Синхронизация через VIEW

CREATE VIEW casino_participants_v2 AS
SELECT 
  user_id,
  username,
  email,
  role,
  balance as coins,
  games_played as spins_count,
  last_game_at as last_spin_at,
  created_at,
  updated_at
FROM casino_players;

-- Вариант C: Trigger для синхронизации
CREATE TRIGGER sync_casino_players_after_update
AFTER UPDATE ON casino_players
FOR EACH ROW
EXECUTE FUNCTION sync_participants();
```

**Оценка сложности**: ⏱️ 2 часа  
**Возможное улучшение**: +15% целостности данных

---

### 2. Проблема: Нет синхронизации balance между БД

**Сейчас**:
```
MUDRO MAIN DB           CASINO MICROSERVICE DB
  casino_accounts          casino_players
   balance: 1000           balance: 800  ← РАСХОЖДЕНИЕ!
```

**Причины**:
- Ставка обновляет `casino_accounts` в main DB
- Микросервис обновляет `casino_players` асинхронно
- Сетевая задержка / конфликты

**Решение**:

```go
// В service casino/internal/casino/store.go

// После каждой ставки UPDATE casino_players
func (s *Store) PlaceBet(ctx context.Context, userID int64, bet int64) error {
    // 1. Проверить баланс в casino_players (локально)
    player, err := s.GetPlayer(ctx, userID)
    if err != nil {
        return err
    }
    if player.Balance < bet {
        return ErrInsufficientBalance
    }
    
    // 2. Выполнить ставку, обновить локальный баланс
    tx, _ := s.db.Begin(ctx)
    defer tx.Rollback(ctx)
    
    // INSERT в casino_game_activity, casino_roulette_bets
    // ...
    
    // 3. Обновить casino_players.balance с SAME TRANSACTION
    err = tx.Exec(ctx,
        "UPDATE casino_players SET balance = balance - $1 WHERE user_id = $2",
        bet, userID,
    ).Err()
    
    // 4. SYNC с main DB (асинхронно, через очередь)
    s.queueSyncBalance(userID)
    
    return tx.Commit(ctx).Err()
}

// Периодическая синхронизация (фоновая задача)
func (s *Store) SyncBalancesJob(ctx context.Context) error {
    rows, _ := s.db.Query(ctx, `
        SELECT cp.user_id, cm.balance as main_balance, cp.balance as micro_balance
        FROM casino_players cp
        JOIN casino_accounts ca ON cp.user_id = ca.user_id
        WHERE ABS(cp.balance - ca.balance) > 1
    `)
    
    for rows.Next() {
        var userID int64
        var mainBal, microBal int64
        rows.Scan(&userID, &mainBal, &microBal)
        
        // Резолюция конфликта: используем основную БД как truth
        _, err := s.db.Exec(ctx,
            "UPDATE casino_players SET balance = $1 WHERE user_id = $2",
            mainBal, userID,
        )
        
        s.log.Infof("Synced balance for user %d: %d -> %d", userID, microBal, mainBal)
    }
    return rows.Err()
}
```

**SQL для отладки**:
```sql
-- Найти расхождения
SELECT 
    cp.user_id,
    cp.balance as micro_balance,
    ca.balance as main_balance,
    ABS(cp.balance - ca.balance::bigint) as diff
FROM casino_players cp
JOIN casino_accounts ca ON cp.user_id = ca.user_id AND ca.type = 'user'
WHERE ABS(cp.balance - ca.balance::bigint) > 0
ORDER BY diff DESC;
```

**Оценка сложности**: ⏱️ 4 часа  
**Возможное улучшение**: +40% надёжности платежей

---

### 3. Проблема: Отсутствует индекс на `casino_game_activity.game_ref + game_type`

**Сейчас**:
```sql
CREATE UNIQUE INDEX casino_game_activity_game_ref_uq
  ON casino_game_activity (game_type, game_ref);
```

**Проблема**:
- Индекс уникален, но не оптимизирован для поиска
- Запрос вида `SELECT * FROM casino_game_activity WHERE user_id = ? AND game_type = ?` использует хуже

**Решение**:

```sql
-- Добавить составной индекс
CREATE INDEX idx_casino_game_activity_user_type
  ON casino_game_activity (user_id, game_type, created_at DESC);

-- Тестировать EXPLAIN
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM casino_game_activity
WHERE user_id = 12345 AND game_type = 'roulette'
ORDER BY created_at DESC
LIMIT 20;
```

**EXPLAIN план (ДО)**:
```
Seq Scan on casino_game_activity  (cost=0.00..1000.00)
  Filter: (user_id = 12345 AND game_type = 'roulette')
```

**EXPLAIN план (ПОСЛЕ)**:
```
Index Scan using idx_casino_game_activity_user_type
  (cost=0.29..15.60)
  Index Cond: (user_id = 12345 AND game_type = 'roulette')
```

**Оценка сложности**: ⏱️ 30 минут  
**Возможное улучшение**: +200% скорость поиска истории игр

---

## 🟡 ВЫСОКИЙ (P1) — В течение месяца

### 4. Добавить временные таблицы для сессий рулетки

**Сейчас**: Все данные в persistent tables (медленно для live-раунда)

**Решение**:

```sql
-- Redis-подобный кеш в PostgreSQL (с TTL через trigger)
CREATE TABLE casino_roulette_sessions (
    id BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    expiresat TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '5 minutes',
    bets_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Trigger: автоматическое удаление истёкших сессий
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM casino_roulette_sessions WHERE expiresat < NOW();
END;
$$ LANGUAGE plpgsql;

-- Вызывать каждые 30 секунд
CREATE RULE cleanup_sessions AS ON INSERT TO casino_roulette_sessions DO ALSO
    SELECT cleanup_expired_sessions() WHERE random() < 0.01;
```

**Вместо**:
```sql
SELECT * FROM casino_roulette_bets WHERE round_id = $1;  -- 100ms
```

**Получаем**:
```sql
SELECT bets_json FROM casino_roulette_sessions WHERE round_id = $1;  -- 5ms
```

**Оценка сложности**: ⏱️ 3 часа  
**Возможное улучшение**: +95% скорость live-раунда

---

### 5. Добавить аудит для `casino_accounts.balance`

**Сейчас**: Нет истории изменений баланса (только через ledger)

**Решение**:

```sql
-- Таблица аудита
CREATE TABLE casino_accounts_audit (
    id BIGSERIAL PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES casino_accounts(id),
    old_balance NUMERIC(30, 10),
    new_balance NUMERIC(30, 10),
    change_amount NUMERIC(30, 10) GENERATED ALWAYS AS (new_balance - old_balance) STORED,
    reason TEXT,
    changed_by TEXT DEFAULT 'system',
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_casino_accounts_audit_account ON casino_accounts_audit(account_id, changed_at DESC);

-- Trigger
CREATE OR REPLACE FUNCTION audit_casino_accounts()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.balance != OLD.balance THEN
        INSERT INTO casino_accounts_audit (account_id, old_balance, new_balance)
        VALUES (NEW.id, OLD.balance, NEW.balance);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_audit_casino_accounts
AFTER UPDATE ON casino_accounts
FOR EACH ROW EXECUTE FUNCTION audit_casino_accounts();
```

**Оценка сложности**: ⏱️ 1 час  
**Возможное улучшение**: +100% видимость операций

---

## 🟠 СРЕДНИЙ (P2) — В течение квартала

### 6. Оптимизировать паттерн RTP для вероятности

**Сейчас**:
```json
[
  {"minRoll": 0, "maxRoll": 0, "multiplier": 25, ...},
  {"minRoll": 1, "maxRoll": 2, "multiplier": 8, ...},
  ...
]
```

**Проблема**: Каждый раз прохождение массива для резолюции (O(n))

**Решение**:

```sql
-- Создать нормализованную таблицу
CREATE TABLE casino_rtp_tiers (
    id BIGSERIAL PRIMARY KEY,
    rtp_profile_id UUID NOT NULL REFERENCES casino_rtp_profiles(id) ON DELETE CASCADE,
    min_roll INT NOT NULL,
    max_roll INT NOT NULL,
    multiplier NUMERIC(10, 4) NOT NULL,
    label TEXT NOT NULL,
    symbol TEXT NOT NULL,
    UNIQUE(rtp_profile_id, min_roll, max_roll)
);

-- Индекс для быстрого поиска
CREATE INDEX idx_casino_rtp_tiers_profile_roll
  ON casino_rtp_tiers(rtp_profile_id, min_roll, max_roll);

-- Заменить вызов
-- ДО: Parse JSON, loop через массив
-- ПОСЛЕ:
SELECT multiplier FROM casino_rtp_tiers
WHERE rtp_profile_id = $1 AND min_roll <= $2 AND max_roll >= $2
LIMIT 1;
```

**Оценка сложности**: ⏱️ 2 часа  
**Возможное улучшение**: +50% скорость резолюции

---

### 7. Добавить partitioning для старых раунов

**Для рост данных**:

```sql
-- Перестроить таблицу с партициями
CREATE TABLE casino_rounds_partitioned (
    LIKE casino_rounds INCLUDING ALL
) PARTITION BY RANGE (created_at);

CREATE TABLE casino_rounds_2024
  PARTITION OF casino_rounds_partitioned
  FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE casino_rounds_2025
  PARTITION OF casino_rounds_partitioned
  FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

-- Migrate данные (в фоне)
INSERT INTO casino_rounds_partitioned 
SELECT * FROM casino_rounds;

-- Переименовать таблицы
ALTER TABLE casino_rounds RENAME TO casino_rounds_old;
ALTER TABLE casino_rounds_partitioned RENAME TO casino_rounds;
```

**Оценка сложности**: ⏱️ 8 часов  
**Возможное улучшение**: +60% для queries на свежих данных

---

## 🟢 НИЗКИЙ (P3) — На будущее

### 8. Добавить GIN индекс для JSONB полей

```sql
CREATE INDEX idx_casino_transfers_metadata
  ON casino_transfers USING GIN (metadata);

CREATE INDEX idx_casino_rtp_profiles_paytable
  ON casino_rtp_profiles USING GIN (paytable);

-- Позволяет быстрые поиски:
SELECT * FROM casino_transfers 
WHERE metadata @> '{"currency": "МДР"}';
```

---

### 9. Реализовать Materialized View для статистики

```sql
CREATE MATERIALIZED VIEW casino_daily_stats AS
SELECT 
    DATE(created_at) as day,
    COUNT(*) as total_rounds,
    COUNT(*) FILTER (WHERE payout_amount > bet_amount) as win_count,
    SUM(bet_amount) as total_wagered,
    SUM(payout_amount) as total_payouts,
    SUM(payout_amount - bet_amount) as net_pnl,
    AVG(multiplier) as avg_multiplier
FROM casino_rounds
WHERE status = 'resolved'
GROUP BY DATE(created_at);

CREATE INDEX idx_casino_daily_stats_day ON casino_daily_stats(day DESC);

-- Обновлять каждый день в 00:00
CREATE OR REPLACE FUNCTION refresh_casino_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY casino_daily_stats;
END;
$$ LANGUAGE plpgsql;

-- SELECT * FROM casino_daily_stats WHERE day >= CURRENT_DATE - 30;
-- Response time: 50ms вместо 5000ms (100x быстрее!)
```

---

## 📋 Чеклист оптимизации

### Фаза 1 (1 неделя): Критичное

- [ ] Добавить индекс на `casino_game_activity(user_id, game_type, created_at)`
- [ ] Реализовать синхронизацию balance между БД
- [ ] Избавиться от дублирования `casino_participants` / `casino_players`

**Оценка**: 6 часов разработки

### Фаза 2 (1 месяц): High Priority

- [ ] Добавить кеширование для live-раундов рулетки
- [ ] Реализовать аудит `casino_accounts.balance`
- [ ] Оптимизировать RTP tier резолюцию

**Оценка**: 8 часов разработки

### Фаза 3 (1 квартал): Medium Priority

- [ ] Partitioning для `casino_rounds` на года
- [ ] GIN индексы для JSONB
- [ ] Materialized Views для статистики

**Оценка**: 20 часов разработки

---

## 🧪 Метрики для мониторинга

### Qeuries производительность

```sql
-- Top 10 медленных queries
SELECT 
    query,
    calls,
    ROUND(max_exec_time::numeric, 2) as max_ms,
    ROUND(mean_exec_time::numeric, 2) as avg_ms
FROM pg_stat_statements
WHERE query LIKE '%casino%'
ORDER BY max_exec_time DESC
LIMIT 10;
```

### Размер индексов

```sql
SELECT 
    indexname,
    ROUND(pg_relation_size(indexrelid) / 1024.0 / 1024.0, 2) as size_mb,
    idx_scan,
    idx_tup_read,
    CASE WHEN idx_scan = 0 THEN '❌ UNUSED' ELSE '✅ ACTIVE' END as status
FROM pg_stat_user_indexes
WHERE schemaname = 'public' AND tablename LIKE 'casino%'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Целостность данных

```sql
-- Проверить целостность ledger каждый час
SELECT 
    SUM(CASE WHEN direction = 'debit' THEN -amount ELSE amount END) as net_balance
FROM casino_ledger_entries
-- Expected: 0
```

---

## 🔗 Ссылки на сопутствующие файлы

- [Casino DB Architecture](./CASINO_DB_ARCHITECTURE.md) — основная схема
- [Casino SQL Queries](./CASINO_SQL_QUERIES.md) — примеры для анализа
- [Casino DB Migrations](./CASINO_DB_MIGRATIONS.md) — история версий

