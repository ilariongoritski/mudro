# Casino DB — SQL Queries для анализа и отладки

## 📊 Аналитика финансов

### 1. Балансы всех счетов

```sql
SELECT 
    ca.id,
    u.id as user_id,
    u.username,
    ca.code,
    ca.type,
    ca.balance,
    ca.currency,
    ca.updated_at
FROM casino_accounts ca
LEFT JOIN users u ON ca.user_id = u.id
ORDER BY ca.balance DESC;
```

### 2. История транзакций (ledger) по пользователю

```sql
SELECT 
    cle.id,
    ct.kind,
    ca.code,
    cle.direction,
    cle.amount,
    ct.metadata,
    cle.created_at
FROM casino_ledger_entries cle
JOIN casino_transfers ct ON cle.transfer_id = ct.id
JOIN casino_accounts ca ON cle.account_id = ca.id
WHERE ca.user_id = $1  -- подставить user_id
ORDER BY cle.created_at DESC
LIMIT 50;
```

### 3. Баланс с аудитом (проверка целостности ledger)

```sql
-- Должно быть 0 (сумма всех дебитов = сумма всех кредитов)
SELECT 
    SUM(CASE WHEN direction = 'debit' THEN -amount ELSE amount END) as net_balance
FROM casino_ledger_entries;

-- По счетам
SELECT 
    ca.code,
    SUM(CASE WHEN direction = 'debit' THEN -amount ELSE amount END) as calculated_balance,
    ca.balance as stored_balance,
    ABS(ca.balance - SUM(CASE WHEN direction = 'debit' THEN -amount ELSE amount END)) as variance
FROM casino_ledger_entries cle
JOIN casino_accounts ca ON cle.account_id = ca.id
GROUP BY ca.id, ca.code, ca.balance
HAVING ABS(ca.balance - SUM(CASE WHEN direction = 'debit' THEN -amount ELSE amount END)) > 0;
```

### 4. Статистика по типам операций

```sql
SELECT 
    ct.kind,
    COUNT(*) as count,
    SUM(cle.amount) FILTER (WHERE cle.direction = 'credit') as total_credited,
    SUM(cle.amount) FILTER (WHERE cle.direction = 'debit') as total_debited,
    AVG(cle.amount) as avg_amount,
    MAX(cle.created_at) as last_at
FROM casino_transfers ct
LEFT JOIN casino_ledger_entries cle ON ct.id = cle.transfer_id
GROUP BY ct.kind
ORDER BY count DESC;
```

---

## 🎲 Анализ игровых раундов

### 1. Все раунды пользователя

```sql
SELECT 
    id,
    user_id,
    status,
    bet_amount,
    payout_amount,
    multiplier,
    tier_label,
    roll,
    CASE 
        WHEN status = 'resolved' AND payout_amount > bet_amount THEN 'WIN'
        WHEN status = 'resolved' AND payout_amount > 0 THEN 'CASHOUT'
        WHEN status = 'resolved' THEN 'LOST'
        ELSE 'PENDING'
    END as result,
    created_at,
    resolved_at
FROM casino_rounds
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 100;
```

### 2. Статистика игрока (win rate, average bet)

```sql
SELECT 
    user_id,
    COUNT(*) as total_rounds,
    COUNT(*) FILTER (WHERE status = 'resolved') as resolved_rounds,
    COUNT(*) FILTER (WHERE status = 'resolved' AND payout_amount > bet_amount) as wins,
    ROUND(100.0 * COUNT(*) FILTER (WHERE status = 'resolved' AND payout_amount > bet_amount) 
        / NULLIF(COUNT(*) FILTER (WHERE status = 'resolved'), 0), 2) as win_rate_percent,
    ROUND(AVG(bet_amount)::numeric, 2) as avg_bet,
    MAX(bet_amount) as max_bet,
    SUM(bet_amount) as total_wagered,
    SUM(payout_amount) as total_payouts,
    ROUND((SUM(payout_amount) - SUM(bet_amount))::numeric, 2) as net_pnl,
    MIN(created_at) as first_round,
    MAX(created_at) as last_round
FROM casino_rounds
WHERE status = 'resolved'
GROUP BY user_id
ORDER BY total_wagered DESC;
```

### 3. Проверка провабли-фэйр

```sql
-- Раунды, где не хватает данных для верификации
SELECT 
    id,
    user_id,
    server_seed,
    server_seed_hash,
    client_seed,
    nonce,
    round_hash,
    status,
    CASE 
        WHEN server_seed IS NULL THEN 'Missing server_seed'
        WHEN server_seed_hash IS NULL THEN 'Missing server_seed_hash'
        WHEN client_seed IS NULL AND status = 'resolved' THEN 'Missing client_seed (RESOLVED!)'
        WHEN nonce < 0 THEN 'Invalid nonce'
        WHEN round_hash IS NULL AND status = 'resolved' THEN 'Missing round_hash (RESOLVED!)'
        ELSE 'OK'
    END as issue
FROM casino_rounds
WHERE 
    server_seed IS NULL 
    OR server_seed_hash IS NULL
    OR (client_seed IS NULL AND status = 'resolved')
    OR (round_hash IS NULL AND status = 'resolved')
ORDER BY created_at DESC;
```

### 4. Статистика по roll distribution (проверка честности)

```sql
SELECT 
    CASE 
        WHEN roll = 0 THEN '0: МЕГА ДЖЕКПОТ'
        WHEN roll BETWEEN 1 AND 2 THEN '1-2: ДЖЕКПОТ'
        WHEN roll BETWEEN 3 AND 7 THEN '3-7: СУПЕР'
        WHEN roll BETWEEN 8 AND 17 THEN '8-17: КРУТО'
        WHEN roll BETWEEN 18 AND 37 THEN '18-37: ХОРОШО'
        WHEN roll BETWEEN 38 AND 52 THEN '38-52: МЕЛОЧЬ'
        WHEN roll BETWEEN 53 AND 99 THEN '53-99: МИМО'
    END as tier,
    COUNT(*) as count,
    ROUND(100.0 * COUNT(*) / SUM(COUNT(*)) OVER (), 2) as percent,
    ROUND(AVG(multiplier)::numeric, 2) as avg_multiplier,
    ROUND(AVG(COALESCE(payout_amount, 0))::numeric, 2) as avg_payout
FROM casino_rounds
WHERE status = 'resolved'
GROUP BY tier
ORDER BY CASE 
    WHEN tier LIKE '0:%' THEN 1
    WHEN tier LIKE '1-2:%' THEN 2
    WHEN tier LIKE '3-7:%' THEN 3
    WHEN tier LIKE '8-17:%' THEN 4
    WHEN tier LIKE '18-37:%' THEN 5
    WHEN tier LIKE '38-52:%' THEN 6
    ELSE 7
END;
```

---

## 🎯 RTP управление

### 1. RTP профили и их назначения

```sql
SELECT 
    crp.name,
    crp.rtp,
    crp.is_default,
    COUNT(cra.id) as assigned_to_users,
    array_agg(DISTINCT u.username) as sample_users
FROM casino_rtp_profiles crp
LEFT JOIN casino_rtp_assignments cra ON crp.id = cra.rtp_profile_id
LEFT JOIN users u ON cra.user_id = u.id
GROUP BY crp.id, crp.name, crp.rtp, crp.is_default
ORDER BY crp.is_default DESC, crp.rtp DESC;
```

### 2. Активные RTP назначения (с истечением)

```sql
SELECT 
    u.id,
    u.username,
    crp.name as rtp_profile,
    crp.rtp,
    cra.assigned_by,
    cra.reason,
    cra.expires_at,
    CASE 
        WHEN cra.expires_at IS NULL THEN 'Permanent'
        WHEN cra.expires_at < NOW() THEN 'EXPIRED'
        ELSE CONCAT(EXTRACT(DAY FROM cra.expires_at - NOW()), ' days left')
    END as status,
    cra.created_at
FROM casino_rtp_assignments cra
JOIN users u ON cra.user_id = u.id
JOIN casino_rtp_profiles crp ON cra.rtp_profile_id = crp.id
ORDER BY cra.expires_at ASC NULLS LAST;
```

### 3. Истёкшие RTP назначения

```sql
DELETE FROM casino_rtp_assignments
WHERE expires_at < NOW()
RETURNING user_id, rtp_profile_id;
```

---

## 🔐 Идемпотентность и защита от дублей

### 1. Попытки дублирования (повторная отправка одного ключа)

```sql
SELECT 
    user_id,
    key,
    COUNT(*) as attempt_count,
    STRING_AGG(DISTINCT status, ', ') as statuses,
    MIN(created_at) as first_attempt,
    MAX(created_at) as last_attempt,
    EXTRACT(EPOCH FROM MAX(created_at) - MIN(created_at)) as seconds_between
FROM casino_idempotency
GROUP BY user_id, key
HAVING COUNT(*) > 1
ORDER BY attempt_count DESC;
```

### 2. Очистка старых idempotency ключей (старше 24 часов, статус != processing)

```sql
DELETE FROM casino_idempotency
WHERE created_at < NOW() - INTERVAL '24 hours'
  AND status != 'processing'
RETURNING id, user_id;
```

### 3. Зависшие операции (processing > 5 минут)

```sql
SELECT 
    id,
    user_id,
    key,
    status,
    created_at,
    EXTRACT(EPOCH FROM NOW() - created_at) / 60 as minutes_pending
FROM casino_idempotency
WHERE status = 'processing'
  AND created_at < NOW() - INTERVAL '5 minutes'
ORDER BY created_at ASC;
```

---

## 📉 Проблемы и диагностика

### 1. Пользователи с отрицательным балансом (ошибка!)

```sql
SELECT 
    ca.user_id,
    u.username,
    ca.balance,
    SUM(CASE WHEN cle.direction = 'debit' THEN -cle.amount ELSE cle.amount END) as calculated,
    COUNT(cle.id) as ledger_entries
FROM casino_accounts ca
LEFT JOIN users u ON ca.user_id = u.id
LEFT JOIN casino_ledger_entries cle ON ca.id = cle.account_id
WHERE ca.balance < 0
GROUP BY ca.id, ca.user_id, u.username, ca.balance;
```

### 2. Орфанские transfers (без записей в ledger)

```sql
SELECT 
    ct.id,
    ct.kind,
    ct.metadata,
    COUNT(cle.id) as ledger_count,
    ct.created_at
FROM casino_transfers ct
LEFT JOIN casino_ledger_entries cle ON ct.id = cle.transfer_id
WHERE cle.id IS NULL
GROUP BY ct.id;
```

### 3. Раунды без статуса (невозможно)

```sql
SELECT * FROM casino_rounds WHERE status IS NULL;
```

### 4. Несоответствие между bet_amount и ledger для раунда

```sql
WITH round_bets AS (
    SELECT 
        cr.id,
        cr.user_id,
        cr.bet_amount,
        SUM(cle.amount) FILTER (WHERE cle.direction = 'debit') as ledger_debit,
        COUNT(cle.id) as ledger_count
    FROM casino_rounds cr
    LEFT JOIN casino_accounts ca ON cr.user_id = ca.user_id
    LEFT JOIN casino_ledger_entries cle ON ca.id = cle.account_id
    WHERE cr.status = 'resolved'
    GROUP BY cr.id, cr.user_id, cr.bet_amount
)
SELECT *
FROM round_bets
WHERE ABS(COALESCE(bet_amount, 0) - COALESCE(ledger_debit, 0)) > 0.01;
```

---

## ⚡ Оптимизация и performance

### 1. Индексы — использование и stats

```sql
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    CASE 
        WHEN idx_scan = 0 THEN '❌ UNUSED'
        WHEN idx_scan < 100 THEN '⚠️  RARELY USED'
        ELSE '✅ ACTIVE'
    END as status
FROM pg_stat_user_indexes
WHERE tablename LIKE 'casino_%'
ORDER BY idx_scan DESC;
```

### 2. Размер таблиц казино

```sql
SELECT 
    tablename,
    ROUND(pg_total_relation_size(schemaname || '.' || tablename) / 1024.0 / 1024.0, 2) as size_mb,
    n_live_tup as row_count,
    ROUND(n_live_tup * avg_tuple_len / 1024.0, 2) as estimated_data_size_mb
FROM pg_stat_user_tables
WHERE tablename LIKE 'casino_%'
ORDER BY pg_total_relation_size(schemaname || '.' || tablename) DESC;
```

### 3. Медленные запросы (если включен pg_stat_statements)

```sql
SELECT 
    query,
    calls,
    ROUND(mean_exec_time::numeric, 2) as avg_ms,
    ROUND(max_exec_time::numeric, 2) as max_ms,
    ROUND(total_exec_time::numeric, 2) as total_ms,
    rows
FROM pg_stat_statements
WHERE query LIKE '%casino%'
ORDER BY total_exec_time DESC
LIMIT 10;
```

---

## 📱 Микросервис Casino (services/casino/)

### 1. Лучшие игроки по total_wagered

```sql
SELECT 
    user_id,
    username,
    display_name,
    balance,
    total_wagered,
    total_won,
    games_played,
    roulette_rounds_played,
    level,
    ROUND(100.0 * total_won / NULLIF(total_wagered, 0), 2) as roi_percent,
    last_game_at
FROM casino_players
ORDER BY total_wagered DESC
LIMIT 20;
```

### 2. История игр по типам

```sql
SELECT 
    game_type,
    COUNT(*) as count,
    SUM(bet_amount) as total_bet,
    SUM(payout_amount) as total_payout,
    SUM(net_result) as net_pnl,
    ROUND(100.0 * SUM(net_result) / NULLIF(SUM(bet_amount), 0), 2) as rtp_percent,
    MAX(created_at) as last_played
FROM casino_game_activity
GROUP BY game_type
ORDER BY count DESC;
```

### 3. Активные раунды рулетки

```sql
SELECT 
    id,
    status,
    winning_number,
    winning_color,
    betting_opens_at,
    betting_closes_at,
    spin_started_at,
    resolved_at,
    (SELECT COUNT(*) FROM casino_roulette_bets WHERE round_id = cr.id) as bet_count,
    (SELECT SUM(stake) FROM casino_roulette_bets WHERE round_id = cr.id) as total_stake
FROM casino_roulette_rounds cr
WHERE status IN ('betting', 'spinning')
ORDER BY betting_opens_at DESC;
```

---

## 🧹 Техническое обслуживание

### 1. Перестройка индексов (при фрагментации)

```sql
REINDEX INDEX CONCURRENTLY idx_casino_accounts_user;
REINDEX INDEX CONCURRENTLY idx_casino_rounds_user;
REINDEX INDEX CONCURRENTLY idx_casino_ledger_transfer;
REINDEX INDEX CONCURRENTLY idx_casino_ledger_account;
REINDEX INDEX CONCURRENTLY idx_casino_rtp_assign_user;
```

### 2. Анализ таблиц (обновление статистики)

```sql
ANALYZE casino_accounts;
ANALYZE casino_transfers;
ANALYZE casino_ledger_entries;
ANALYZE casino_rounds;
ANALYZE casino_rtp_profiles;
ANALYZE casino_rtp_assignments;
ANALYZE casino_idempotency;
```

### 3. Вакум (очистка мёртвых строк)

```sql
VACUUM ANALYZE casino_accounts;
VACUUM ANALYZE casino_ledger_entries;
VACUUM ANALYZE casino_rounds;
```

---

## 📝 Примеры операций

### Создание нового аккаунта с валидацией

```sql
BEGIN;

INSERT INTO casino_accounts (user_id, type, code, currency, balance)
VALUES ($1, 'user', 'USER_' || $1, 'МДР', 0)
RETURNING id, user_id, balance;

-- Проверка целостности
SELECT SUM(CASE WHEN direction = 'debit' THEN -amount ELSE amount END)
FROM casino_ledger_entries
WHERE account_id = (SELECT id FROM casino_accounts WHERE user_id = $1);

COMMIT;
```

### Ставка (bet) с идемпотентностью

```sql
BEGIN;

-- 1. Проверить idempotency
INSERT INTO casino_idempotency (user_id, key, request_hash, status)
VALUES ($1, $2, $3, 'processing')
ON CONFLICT (user_id, key) DO UPDATE SET status = 'processing'
RETURNING id, status;

-- 2. Подготовить раунд
INSERT INTO casino_rounds (user_id, server_seed, server_seed_hash, status)
VALUES ($1, $4, $5, 'prepared')
RETURNING id;

COMMIT;
```

