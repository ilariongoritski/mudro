# Архитектура БД Казино (mudro11)

## 📊 Обзор систем

Проект содержит **две отдельные casino-системы**:

### 1️⃣ **Основная система (migrations/012_casino.sql)**
Потокоустойчивая финансовая система в главной БД mudro

### 2️⃣ **Live Games (services/casino/)**
Микросервис казино со своей БД (рулетка, слоты, plinko)

---

## 📐 Диаграмма архитектуры

```
┌─────────────────────────────────────────────────────────────┐
│                    DATABASE SCHEMA                           │
└─────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│  MUDRO ГЛАВНАЯ БД (PostgreSQL 15)                           │
│  migrations/012_casino.sql                                   │
└──────────────────────────────────────────────────────────────┘
       │
       ├─── 💳 ФИНАНСОВАЯ СИСТЕМА ────────────────────────────
       │     └─ casino_accounts
       │        ├─ id (UUID PK)
       │        ├─ user_id (FK users)
       │        ├─ type: 'user' | 'system'
       │        ├─ code: UNIQUE (SYSTEM_HOUSE_POOL, SYSTEM_SETTLEMENT_POOL)
       │        ├─ currency: МДР
       │        ├─ balance: NUMERIC(30,10)  ← точность до 10 десятичных
       │        └─ INDEX: idx_casino_accounts_user
       │
       ├─── 📋 ДВОЙНАЯ БУХГАЛТЕРИЯ (Double-Entry Ledger) ────
       │     ├─ casino_transfers
       │     │  ├─ id (UUID PK)
       │     │  ├─ kind: 'bet_stake'|'bet_payout'|'deposit'|'withdrawal'|'adjustment'
       │     │  └─ metadata: JSONB
       │     │
       │     └─ casino_ledger_entries (APPEND-ONLY)
       │        ├─ id (UUID PK)
       │        ├─ transfer_id (FK)
       │        ├─ account_id (FK)
       │        ├─ direction: 'debit' | 'credit'  ← тройной ввод!
       │        ├─ amount: NUMERIC(30,10)
       │        └─ Индексы:
       │           ├─ idx_casino_ledger_transfer
       │           └─ idx_casino_ledger_account
       │
       ├─── 🎲 PROVABLY FAIR СИСТЕМА ───────────────────────
       │     └─ casino_rounds
       │        ├─ id (UUID PK)
       │        ├─ user_id (FK users, CASCADE)
       │        ├─ server_seed: TEXT
       │        ├─ server_seed_hash: TEXT
       │        ├─ client_seed: TEXT (опционально)
       │        ├─ nonce: INT
       │        ├─ round_hash: TEXT
       │        ├─ roll: INT (0-99)
       │        ├─ bet_amount: NUMERIC
       │        ├─ payout_amount: NUMERIC
       │        ├─ multiplier: NUMERIC(10,4)
       │        ├─ tier_label: TEXT (мегаджекпот, джекпот и т.д.)
       │        ├─ status: 'prepared'|'resolved'|'cancelled'
       │        └─ INDEX: idx_casino_rounds_user (user_id, status)
       │
       ├─── 🎯 RTP УПРАВЛЕНИЕ (Управление честностью) ──────
       │     ├─ casino_rtp_profiles
       │     │  ├─ id (UUID PK)
       │     │  ├─ name: UNIQUE ('default', 'vip', 'shark')
       │     │  ├─ rtp: NUMERIC(5,2)  [96%, 97%, 98%]
       │     │  ├─ paytable: JSONB
       │     │  │  └─ Пример:
       │     │  │     {
       │     │  │       "minRoll": 0,
       │     │  │       "maxRoll": 0,
       │     │  │       "multiplier": 25,
       │     │  │       "label": "👑👑👑 МЕГА ДЖЕКПОТ",
       │     │  │       "symbol": "👑👑👑"
       │     │  │     }
       │     │  └─ is_default: BOOLEAN
       │     │
       │     └─ casino_rtp_assignments
       │        ├─ id (UUID PK)
       │        ├─ user_id (FK users, CASCADE)
       │        ├─ rtp_profile_id (FK casino_rtp_profiles, CASCADE)
       │        ├─ assigned_by: TEXT ('system', 'admin')
       │        ├─ reason: TEXT
       │        ├─ expires_at: TIMESTAMPTZ (опционально)
       │        ├─ UNIQUE(user_id, rtp_profile_id)
       │        └─ INDEX: idx_casino_rtp_assign_user
       │
       └─── 🔐 ИДЕМПОТЕНТНОСТЬ (Защита от дублей) ─────────
             └─ casino_idempotency
                ├─ id (UUID PK)
                ├─ user_id (FK)
                ├─ key: TEXT (клиентский ID запроса)
                ├─ request_hash: TEXT
                ├─ status: 'processing'|'succeeded'|'failed'
                ├─ response: JSONB
                └─ UNIQUE(user_id, key)


┌──────────────────────────────────────────────────────────────┐
│  CASINO МИКРОСЕРВИС (services/casino/) - ОТДЕЛЬНАЯ БД       │
└──────────────────────────────────────────────────────────────┘
       │
       ├─── 001_init.sql ──────────────────────────────────
       │     ├─ casino_config
       │     │  ├─ id: BOOLEAN PK (singleton pattern!)
       │     │  ├─ rtp_percent: DOUBLE
       │     │  ├─ initial_balance: BIGINT
       │     │  ├─ symbol_weights: JSONB
       │     │  └─ paytable: JSONB
       │     │
       │     └─ casino_participants (старая таблица)
       │        ├─ user_id: BIGINT PK
       │        ├─ username, email, role
       │        ├─ coins: BIGINT
       │        ├─ spins_count: BIGINT
       │        └─ last_spin_at: TIMESTAMPTZ
       │
       └─── 002_live_roulette.sql ───────────────────────
             │
             ├─ casino_players (новая таблица)
             │  ├─ user_id: BIGINT PK
             │  ├─ username, email, role, display_name
             │  ├─ avatar_url, telegram_username
             │  ├─ balance: BIGINT
             │  ├─ total_wagered, total_won: BIGINT
             │  ├─ games_played, roulette_rounds_played: BIGINT
             │  ├─ level, xp_progress: (gamification)
             │  └─ last_game_at: TIMESTAMPTZ
             │
             ├─ casino_game_activity
             │  ├─ id: BIGSERIAL PK
             │  ├─ user_id (FK casino_players, CASCADE)
             │  ├─ game_type: TEXT ('slots', 'roulette', 'plinko', 'crash')
             │  ├─ game_ref: TEXT (UNIQUE с game_type)
             │  ├─ bet_amount, payout_amount, net_result: BIGINT
             │  ├─ status: TEXT
             │  ├─ metadata: JSONB
             │  └─ INDEX: (user_id, created_at DESC, id DESC)
             │
             ├─ casino_roulette_rounds (фазы раунда)
             │  ├─ id: BIGSERIAL PK
             │  ├─ status: 'betting'|'locking'|'spinning'|'result'
             │  ├─ winning_number: INT (0-36)
             │  ├─ winning_color: 'red'|'black'|'green'
             │  ├─ display_sequence: JSONB (анимация)
             │  ├─ result_sequence: JSONB
             │  ├─ betting_opens_at, betting_closes_at: TIMESTAMPTZ
             │  ├─ spin_started_at, resolved_at: TIMESTAMPTZ
             │  └─ INDEX: (status, created_at DESC, id DESC)
             │
             └─ casino_roulette_bets (ставки в раундах)
                ├─ id: BIGSERIAL PK
                ├─ round_id (FK casino_roulette_rounds, CASCADE)
                ├─ user_id (FK casino_players, CASCADE)
                ├─ bet_type: 'straight'|'red'|'black'|'even'|'odd'|'low'|'high'
                ├─ bet_value: TEXT (номер при straight ставке)
                ├─ stake: BIGINT
                ├─ payout_amount: BIGINT
                ├─ status: 'placed'|'won'|'lost'|'pending'
                └─ Индексы:
                   ├─ (round_id, user_id, bet_type, bet_value) - UNIQUE
                   ├─ (round_id, created_at ASC, id ASC) - для закрытия ставок
                   └─ (user_id, created_at DESC, id DESC) - история ставок
```

---

## 🔗 Отношения между таблицами

```
┌─────────────────────────────────────────────────────────────┐
│                    СВЯЗИ В СИСТЕМЕ                          │
└─────────────────────────────────────────────────────────────┘

users (id: BIGINT)
  ├─ 1──────────→ N casino_accounts (user_id)
  ├─ 1──────────→ N casino_rounds (user_id, CASCADE)
  ├─ 1──────────→ N casino_rtp_assignments (user_id, CASCADE)
  ├─ 1──────────→ N casino_idempotency (user_id, CASCADE)
  │
  └─ МИКРОСЕРВИС:
     ├─ 1──────────→ 1 casino_players (user_id PK)
     ├─ 1──────────→ N casino_game_activity (user_id, CASCADE)
     └─ 1──────────→ N casino_roulette_bets (user_id, CASCADE)

casino_accounts (id: UUID)
  ├─ 1──────────→ N casino_ledger_entries (account_id, RESTRICT)
  │              ↓
  └─────→ casino_transfers (id: UUID)
         └─→ casino_ledger_entries (transfer_id, CASCADE)

casino_rtp_profiles (id: UUID)
  └─ 1──────────→ N casino_rtp_assignments (rtp_profile_id, CASCADE)

casino_roulette_rounds (id: BIGSERIAL)  
  └─ 1──────────→ N casino_roulette_bets (round_id, CASCADE)
  
casino_players (user_id: BIGINT)
  ├─ 1──────────→ N casino_game_activity (user_id, CASCADE)
  └─ 1──────────→ N casino_roulette_bets (user_id, CASCADE)
```

---

## 🎮 Поддерживаемые игры

### Основная БД (mudro)
- **Slots (Слоты)** — provably fair
- Все связано с `casino_rounds`

### Микросервис (services/casino)
- **Roulette (Рулетка)** — live-раунды, многопользовательские ставки
- **Plinko (Плинко)** — логируется в `casino_game_activity`
- **Crash** — логируется в `casino_game_activity`  
- **Coin Flip** — логируется в `casino_game_activity`
- **Slots (Слоты)** — legacy в `casino_spins`, но мигрирует в `casino_players`

---

## 💰 Финансовая модель

### Двойная бухгалтерия (Double-Entry Ledger)

```sql
-- Пример: User ставит 100 МДР

-- 1. Создать Transfer
INSERT INTO casino_transfers (kind, metadata)
VALUES ('bet_stake', '{"game":"slots", ...}')
RETURNING id;

-- 2. Дебет со счета игрока
INSERT INTO casino_ledger_entries (transfer_id, account_id, direction, amount)
VALUES (transfer_id, user_account_id, 'debit', 100);

-- 3. Кредит на системный счет (house pool)
INSERT INTO casino_ledger_entries (transfer_id, account_id, direction, amount)
VALUES (transfer_id, house_account_id, 'credit', 100);

-- 4. UI видит:
SELECT balance FROM casino_accounts WHERE user_id = $1;
-- результат: -100
```

### Система счетов

| Тип | Код | Назначение |
|-----|-----|-----------|
| `user` | N/A | Личный счет игрока |
| `system` | `SYSTEM_HOUSE_POOL` | Казино (дом) |
| `system` | `SYSTEM_SETTLEMENT_POOL` | Расчетный счет |

---

## 🎯 RTP (Return to Player) управление

По умолчанию создаются 3 профиля:

| Профиль | RTP | Описание |
|---------|-----|---------|
| `default` | 96% | Стандартные игроки |
| `vip` | 97% | VIP-члены |
| `shark` | 98% | Топ-игроки (риск) |

```json
{
  "minRoll": 0,
  "maxRoll": 0,
  "multiplier": 25,
  "label": "👑👑👑 МЕГА ДЖЕКПОТ",
  "symbol": "👑👑👑"
}
```

Таблица выигрышей привязана к rolls 0-99:
- 0: 25x (мега джекпот)
- 1-2: 8x (джекпот)  
- 3-7: 3x (супер)
- 8-17: 1.5x (круто)
- 18-37: 0.8x (хорошо)
- 38-52: 0.6x (мелочь)
- 53-99: 0x (мимо)

---

## ⚙️ Критические индексы

### Основная БД
```sql
-- Поиск счетов по пользователю
CREATE INDEX idx_casino_accounts_user ON casino_accounts(user_id);

-- Быстрый поиск ставок в истории
CREATE INDEX idx_casino_rounds_user ON casino_rounds(user_id, status);

-- Отслеживание финансов  
CREATE INDEX idx_casino_ledger_transfer ON casino_ledger_entries(transfer_id);
CREATE INDEX idx_casino_ledger_account ON casino_ledger_entries(account_id);

-- RTP назначения
CREATE INDEX idx_casino_rtp_assign_user ON casino_rtp_assignments(user_id);
```

### Микросервис
```sql
-- История спинов
CREATE INDEX casino_spins_user_created_idx
  ON casino_spins (user_id, created_at DESC, id DESC);

-- История игр
CREATE INDEX casino_game_activity_user_created_idx
  ON casino_game_activity (user_id, created_at DESC, id DESC);

-- Раунды рулетки
CREATE INDEX casino_roulette_rounds_status_created_idx
  ON casino_roulette_rounds (status, created_at DESC, id DESC);

-- Ставки в раунде
CREATE INDEX casino_roulette_bets_round_created_idx
  ON casino_roulette_bets (round_id, created_at ASC, id ASC);
CREATE INDEX casino_roulette_bets_user_created_idx
  ON casino_roulette_bets (user_id, created_at DESC, id DESC);
```

---

## 🔐 Безопасность

### Защита от дублей (Idempotency)
```
user_id + key → UNIQUE constraint
гарантирует, что один запрос не обработается дважды
```

### Целостность данных
- `casino_ledger_entries.account_id`: RESTRICT (нельзя удалить счет с историей)
- Все игровые данные: CASCADE (удаление пользователя → удаление истории)

### Финансовая точность
- `NUMERIC(30,10)` — точность до 10 десятичных знаков (для микротранзакций)
- `BIGINT` микросервис (целые цифры, масштабируемость)

---

## 📊 Структура RTP профиля в JSONB

```json
[
  {
    "minRoll": 0,
    "maxRoll": 0,
    "multiplier": 25,
    "label": "👑👑👑 МЕГА ДЖЕКПОТ",
    "symbol": "👑👑👑"
  },
  {
    "minRoll": 1,
    "maxRoll": 2,
    "multiplier": 8,
    "label": "💎💎💎 ДЖЕКПОТ",
    "symbol": "💎💎💎"
  },
  ...
]
```

---

## 🎲 Алгоритм Provably Fair

1. **server_seed** → хешируется (`server_seed_hash`) и сохраняется до раунда
2. **Раунд подготовлен** → игрок видит `server_seed_hash` (не видит самого seed)
3. **Игрок ставит** → отправляет `client_seed` и `nonce`
4. **Резолюция**: 
   ```
   HMAC-SHA256(server_seed + client_seed + nonce) → 0-99999
   roll = (результат % 100) // 0-99
   ```
5. **Верификация** → клиент может проверить: `server_seed_hash == hash(server_seed)`

---

## 📈 Метрики (Gamification)

В микросервисе для каждого игрока:

```
level = (total_wagered / 1000) + 1
xp_progress = total_wagered % 1000

Пример: total_wagered = 5500
→ level = 6, xp_progress = 500 (50% к уровню 7)
```

---

## 🔄 Миграции версий

| Версия | Файл | Назначение |
|--------|------|-----------|
| 012 | `migrations/012_casino.sql` | Основная финансовая + provably fair |
| 001 | `services/casino/migrations/001_init.sql` | Legacy slots + config |
| 002 | `services/casino/migrations/002_live_roulette.sql` | Рулетка + обновление players |

---

## ✅ Оптимизации и замечания

✅ **Сильные стороны:**
- Двойная бухгалтерия = неопровержимый аудит финансов
- Provably fair встроена в основную БД
- Отдельный микросервис для масштабирования live-игр
- Идемпотентность защищает от дублей
- RTP управление на уровне БД

⚠️ **Точки внимания:**
- `casino_players` дублирует данные из `casino_participants` (миграция 002)
- `casino_rtp_profiles.paytable` в JSONB (можно индексировать через GIN)
- Нет автоматических триггеров пересчета `balance` (вручную через ledger)
- Микросервис использует `BIGINT` вместо `NUMERIC` (возможно переполнение на 10^18)

