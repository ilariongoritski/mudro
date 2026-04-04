# Casino DB — Quick Reference (Шпаргалка)

## 📌 Одно предложение для каждого компонента

| Компонент | Назначение | Key Insight |
|-----------|-----------|------------|
| `casino_accounts` | Кошельки пользователей | `type: 'user'\|'system'`; UUID PK |
| `casino_transfers` | Основная операция | Связывает дебит/кредит |
| `casino_ledger_entries` | Двойная бухгалтерия | Append-only аудит, `direction: debit\|credit` |
| `casino_rounds` | Provably fair раунды | `server_seed_hash` скрыт до `client_seed` |
| `casino_rtp_profiles` | %RTP по профилям | `default:96%`, `vip:97%`, `shark:98%` |
| `casino_rtp_assignments` | Привязка RTP к юзеру | Опциональный `expires_at` |
| `casino_idempotency` | От дублей операций | `UNIQUE(user_id, key)` |
| `casino_players` | Игроки (microservice) | `level`, `xp_progress` для gamification |
| `casino_game_activity` | Лог всех игр | `game_type: 'slots'\|'roulette'\|'plinko'` |
| `casino_roulette_rounds` | Раунды рулетки | FSM: `betting→locking→spinning→result` |
| `casino_roulette_bets` | Ставки в раунде | 8 типов: `straight, red, black, even, odd, low, high, green` |

---

## 🎯 Три основных фича

### 1️⃣ Двойная бухгалтерия (Double-Entry Ledger)

```sql
-- User ставит $100
INSERT INTO casino_transfers (kind) VALUES ('bet_stake') → T1;
INSERT INTO casino_ledger_entries (..., direction='debit', amount=100) → user_account;
INSERT INTO casino_ledger_entries (..., direction='credit', amount=100) → house_account;
```

✅ Неопровержимый аудит  
✅ Всегда balanced (sum debits = sum credits)  
✅ ACID-гарантирован

---

### 2️⃣ Provably Fair (Честная игра)

```go
// Шаг 1: Server создает seed
server_seed_hash = SHA256(server_seed)

// Шаг 2: Server показывает ТОЛЬКО hash, скрывает seed
// User видит: "server_seed_hash: 0x123abc..."

// Шаг 3: User отправляет client_seed + nonce
HMAC-SHA256(server_seed + client_seed + nonce) → roll (0-99)

// Шаг 4: Server открывает server_seed
User проверяет: hash(server_seed) == server_seed_hash ✅
```

✅ Непредусмотрительно (fair)  
✅ Проверяемо (verifiable)  
✅ Зависит только от `server_seed_hash` (видно до ставки)

---

### 3️⃣ RTP управление (контроль честности казино)

```json
{
  "name": "vip",
  "rtp": 97,  // 97% player wins
  "paytable": [
    {"minRoll": 0, "maxRoll": 0, "multiplier": 30, "label": "👑👑👑 МЕГА"},
    {"minRoll": 1, "maxRoll": 2, "multiplier": 8, "label": "💎💎💎 ДЖЕКПОТ"},
    ...
  ]
}
```

✅ Разные профили для разных юзеров  
✅ Гибкий payable (экспоненты шанса)  
✅ Управляется на уровне БД

---

## 🏗️ Архитектурные решения

### Почему TWO Database?

```
MUDRO MAIN DB                 CASINO MICROSERVICE DB
(одна на всех)                (масштабируется независимо)

migrations/012_casino.sql      services/casino/migrations/
├─ casino_accounts            ├─ casino_players
├─ casino_ledger_entries       ├─ casino_game_activity
├─ casino_rounds               ├─ casino_roulette_rounds
├─ casino_rtp_*                └─ casino_roulette_bets
└─ casino_idempotency

Цель: Независимая масштабируемость live-игр (рулетка)
      Отначальный truth source для финансов (main DB)
```

---

## 🔍 Диагностика в 30 секунд

```bash
# 1. Проверить целостность ledger (должно быть 0)
psql -c "SELECT SUM(CASE WHEN direction='debit' THEN -amount ELSE amount END) FROM casino_ledger_entries;"
# Expected: 0 (or close to 0)

# 2. Проверить баланс каждого аккаунта
psql -c "SELECT code, balance FROM casino_accounts ORDER BY balance DESC LIMIT 5;"

# 3. Найти расхождения между main DB и microservice
psql -c "SELECT cp.user_id, cp.balance::numeric - ca.balance as diff FROM casino_players cp JOIN casino_accounts ca ON cp.user_id = ca.user_id WHERE cp.balance::numeric != ca.balance LIMIT 10;"

# 4. Проверить ошибки в провабли-фэйр
psql -c "SELECT COUNT(*) FROM casino_rounds WHERE status='resolved' AND round_hash IS NULL;"
# Expected: 0
```

---

## 📊 Примеры queries (copy-paste)

### Баланс пользователя с историей (last 10 операций)

```sql
WITH user_account AS (
  SELECT id, balance FROM casino_accounts WHERE user_id = $1 AND type = 'user'
),
history AS (
  SELECT 
    ct.kind, cle.direction, cle.amount,
    cle.created_at, ct.metadata,
    SUM(CASE WHEN cle.direction = 'debit' THEN -cle.amount ELSE cle.amount END)
      OVER (ORDER BY cle.created_at) as running_balance
  FROM casino_ledger_entries cle
  JOIN casino_transfers ct ON cle.transfer_id = ct.id
  JOIN user_account ua ON cle.account_id = ua.id
  ORDER BY cle.created_at DESC
  LIMIT 10
)
SELECT 
  (SELECT balance FROM user_account) as current_balance,
  jsonb_agg(jsonb_build_object(
    'kind', kind,
    'direction', direction,
    'amount', amount,
    'running_balance', running_balance,
    'at', created_at,
    'metadata', metadata
  ) ORDER BY created_at DESC) as history
FROM history;
```

### Статистика игрока за неделю

```sql
SELECT 
  user_id,
  COUNT(*) as rounds,
  COUNT(*) FILTER (WHERE payout_amount > bet_amount) as wins,
  ROUND(100.0 * COUNT(*) FILTER (WHERE payout_amount > bet_amount) / COUNT(*), 2) as win_rate,
  ROUND(AVG(bet_amount), 2) as avg_bet,
  ROUND(SUM(payout_amount - bet_amount), 2) as pnl
FROM casino_rounds
WHERE user_id = $1 
  AND created_at > NOW() - INTERVAL '7 days'
  AND status = 'resolved'
GROUP BY user_id;
```

### Рулетка: последний раунд + мои ставки

```sql
SELECT 
  r.id, r.status,
  r.winning_number, r.winning_color,
  r.betting_opens_at, r.betting_closes_at, r.resolved_at,
  jsonb_agg(jsonb_build_object(
    'id', b.id,
    'bet_type', b.bet_type,
    'bet_value', b.bet_value,
    'stake', b.stake,
    'payout_amount', b.payout_amount,
    'status', b.status
  )) as my_bets
FROM casino_roulette_rounds r
LEFT JOIN casino_roulette_bets b ON r.id = b.round_id AND b.user_id = $1
WHERE r.id = (SELECT MAX(id) FROM casino_roulette_rounds)
GROUP BY r.id;
```

---

## 🚨 Частые проблемы и решения

| Проблема | Причина | Решение |
|----------|---------|---------|
| **Баланс не совпадает** | Микросервис отстаёт | `SELECT * FROM casino_players cp JOIN casino_accounts ca ON ... WHERE cp.balance != ca.balance` → пересинхронизировать |
| **Раунд зависает в `prepared`** | Краш при резолюции | Проверить логи, обновить статус вручную: `UPDATE casino_rounds SET status='resolved' WHERE ...` |
| **Два профиля RTP на юзере** | Дублирование | `DELETE FROM casino_rtp_assignments WHERE id NOT IN (SELECT MIN(id) FROM casino_rtp_assignments GROUP BY user_id, rtp_profile_id)` |
| **Ledger не балансируется** | Ошибка в INSERT | Проверить: `SELECT transfer_id, SUM(amount) FROM casino_ledger_entries GROUP BY transfer_id HAVING COUNT(*) != 2` |
| **Рулетка не принимает ставки** | Раунд закрыт | Проверить `betting_closes_at > NOW()` в `casino_roulette_rounds` |

---

## 🎮 Как работает типовая игра (Slots)

```
1. Client: "Я хочу ставить 100 МДР"
   ↓
2. Server: POST /casino/spin {bet: 100}
   ├─ Проверить баланс в casino_accounts
   └─ Создать casino_rounds (status='prepared')
     └─ Генерировать server_seed, server_seed_hash
     └─ Вернуть round_id, server_seed_hash
   ↓
3. Client: Видит server_seed_hash (не видит самого seed)
   ├─ Генерирует client_seed (random)
   └─ POST /casino/resolve {round_id, client_seed}
   ↓
4. Server:
   ├─ Получить casino_rounds (id=round_id)
   ├─ HMAC-SHA256(server_seed + client_seed + nonce) → roll (0-99)
   ├─ Найти multiplier в паytable для roll
   ├─ Вычислить payout = bet * multiplier
   └─ UPDATE casino_rounds (status='resolved', payout_amount=X, multiplier=Y)
   ↓
5. Обновить баланс (через ledger):
   ├─ INSERT casino_transfers (kind='bet_payout')
   ├─ INSERT casino_ledger_entries (direction='debit', amount=bet, ...)
   ├─ INSERT casino_ledger_entries (direction='credit', amount=payout, ...)
   └─ UPDATE casino_accounts SET balance = balance - bet + payout
   ↓
6. Client:
   ├─ Видит multiplier, payout, symbols
   └─ Новый баланс: balance - bet + payout
```

---

## 📈 Как читать эту документацию

```
Быстрый старт?
↓
Прочитай: CASINO_DB_ARCHITECTURE.md (этот файл)
          + диаграмму Mermaid
          + таблица "Три основных фички"

Нужно писать SQL?
↓
→ CASINO_SQL_QUERIES.md

Нужно оптимизировать?
↓
→ CASINO_DB_OPTIMIZATION.md

Нужно добавить миграцию?
↓
→ CASINO_DB_MIGRATIONS.md

Нужно дебажить финансовую ошибку?
↓
Диагностика в 30 секунд + таблица "Частые проблемы"
```

---

## 🔑 Ключевые метрики для мониторинга

```sql
-- Каждый час проверять (alert если != 0):
SELECT SUM(CASE WHEN direction='debit' THEN -amount ELSE amount END)
FROM casino_ledger_entries;

-- Каждый день:
SELECT DATE(created_at), COUNT(*), SUM(payout_amount - bet_amount) as pnl
FROM casino_rounds WHERE status='resolved'
GROUP BY DATE(created_at);

-- Синхронизация между БД (должна быть < 100):
SELECT COUNT(*) 
FROM casino_players cp
JOIN casino_accounts ca ON cp.user_id = ca.user_id  
WHERE cp.balance::numeric != ca.balance;
```

---

## 🔐 Безопасность: правила

✅ **DO:**
- Всегда使用 transaction для нескольких операций ledger
- Проверять оба направления дебит/кредит перед коммитом
- Сохранять server_seed_hash ДО открытия ставок
- Использовать idempotency key для все бет-запросов

❌ **DON'T:**
- Никогда не обновлять balance напрямую (только через ledger)
- Никогда не доверять client_seed (может быть faked)
- Никогда не удалять ledger entries (только INSERT/append)
- Никогда не менять RTP mid-round

---

## 📞 Контакты смежных систем

- **Frontend**: `/src/features/casino/`
- **Backend Main**: `internal/casino/` (domain, usecase, repository)
- **Backend Microservice**: `services/casino/internal/casino/` (handlers, store)
- **Contracts**: `contracts/http/` (WebSocket для live-рулетки)

