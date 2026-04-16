# 🎰 Casino Roadmap 2026

**Дата:** 2026-04-05
**Версия:** 1.0
**Статус:** Active Development

---

## ✅ Phase 1: MVP (Завершена - 2026-04-05)

### Реализованные игры
1. **Blackjack** ✅
   - Hit/Stand action system
   - Dealer AI (hits until 17+)
   - Payouts: 1:1 win, 2.5:1 blackjack, push handling
   - Database: `casino_blackjack_games` table

2. **Roulette** ✅
   - European roulette (37 numbers: 0-36)
   - Bet types: straight, red/black, odd/even, low/high
   - Live multiplayer with WebSocket
   - Round state management: betting → lock → spin → result
   - Database: `casino_roulette_games`, `casino_roulette_bets`

3. **Plinko** ✅
   - 12 rows, 13 slots
   - Risk levels: low (higher left/right hits), medium, high (center variance)
   - Probabilistic path calculation
   - Database: `casino_plinko_games`

4. **Slots** ✅
   - 3-reel machine with emoji symbols
   - 7 tiers with different multipliers
   - Provably fair system (server + client seed)
   - Database: `casino_rounds` table

### Infrastructure
- Background loops: RouletteLoop, BalanceReconciler, RouletteSessionJanitor
- Telegram bonus verification
- Activity feed & social reactions (emoji picker)
- Dual-DB support (CASINO_MAIN_DSN)
- 6 database migrations

---

## 📌 Phase 2: Game Expansion (Q2 2026 - April/May)

### Game Candidates
| Game | Complexity | Priority | Est. Days |
|------|-----------|----------|----------|
| **Dice** | Low | HIGH | 3 |
| **Baccarat** | Medium | HIGH | 5 |
| **Keno** | Medium | MEDIUM | 4 |
| **Video Poker** | Medium | MEDIUM | 6 |
| **Crash** | High | MEDIUM | 7 |

### Dice Game (PRIORITY 1)
**Description:** Simple 1v1 dice roll (player vs house)

**Game Flow:**
1. Player chooses bet amount
2. Player rolls dice (1-6)
3. House rolls dice (1-6)
4. Higher roll wins (2:1 payout for player)
5. Tie = push

**Backend Requirements:**
- New handler: `POST /api/casino/dice/roll`
- Game engine: random roll with deterministic seeding
- Table: `casino_dice_games` (user_id, player_roll, house_roll, winner, payout)
- RTP: 98% (fair odds with house edge)

**Frontend Requirements:**
- DicePanel.tsx component
- Dice animation (rolling → result)
- Bet interface
- Win/loss notification

**Estimated:** 3 days

---

### Baccarat Game (PRIORITY 2)
**Description:** Card game - Player vs Banker

**Game Flow:**
1. Deal 2 cards to Player
2. Deal 2 cards to Banker
3. Evaluate hands (sum mod 10)
4. If either hand is 8 or 9, it's a natural (game ends)
5. Otherwise, third card rules apply
6. Higher hand wins

**Payouts:**
- Player wins: 1:1
- Banker wins: 0.95:1 (5% commission)
- Tie: 8:1

**Backend Requirements:**
- New handlers: `POST /api/casino/baccarat/deal`, `POST /api/casino/baccarat/stand`
- Game engine: card dealing + hand evaluation
- Table: `casino_baccarat_games` (player_hand, banker_hand, winner, payout)
- RTP: 98.76% (Player), 98.94% (Banker)

**Frontend Requirements:**
- BaccaratPanel.tsx component
- Card rendering (Player hand, Banker hand, Tableau)
- Bet interface (Player/Banker/Tie)
- Hand evaluation UI

**Estimated:** 5 days

---

### Keno Game (PRIORITY 3)
**Description:** Lottery-style game (pick 1-20 numbers, 20 drawn)

**Game Flow:**
1. Player picks 1-20 numbers (0-79 grid)
2. Game draws 20 winning numbers
3. Player wins based on how many matches
4. Multiplier increases with matches

**Payouts:**
- Match 1: 0.5:1
- Match 5: 2:1
- Match 10: 20:1
- Match 15: 500:1
- Match 20: 10,000:1

**Backend Requirements:**
- New handler: `POST /api/casino/keno/draw`
- Game engine: number selection + matching logic
- Table: `casino_keno_games` (selected_numbers, winning_numbers, matches, payout)
- RTP: 95% (house keeps 5%)

**Frontend Requirements:**
- KenoPanel.tsx component
- Number grid (80 numbers, selectable)
- Draw animation
- Match highlighting
- Payout calculator

**Estimated:** 4 days

---

### Video Poker (PRIORITY 4)
**Description:** Single-hand poker machine

**Game Flow:**
1. Deal 5 cards
2. Player selects cards to hold
3. Draw replacement cards
4. Evaluate hand (pair, two pair, three of a kind, etc.)
5. Payout based on hand rank

**Payouts:**
- Royal Flush: 250:1
- Straight Flush: 50:1
- Four of a Kind: 25:1
- Full House: 9:1
- Flush: 6:1
- Straight: 4:1
- Three of a Kind: 3:1
- Two Pair: 2:1
- Pair (J+): 1:1

**Backend Requirements:**
- New handlers: `POST /api/casino/videopoker/deal`, `POST /api/casino/videopoker/draw`
- Game engine: hand evaluation + ranking
- Table: `casino_videopoker_games` (initial_hand, final_hand, hand_rank, payout)
- RTP: 99% (very player-favorable)

**Frontend Requirements:**
- VideoPokerPanel.tsx component
- Card rendering (hold/discard buttons)
- Hand rank display
- Payout table
- Strategy hints (optional)

**Estimated:** 6 days

---

### Crash Game (PRIORITY 5)
**Description:** Multiplier-based game (bet on the multiplier when it crashes)

**Game Flow:**
1. Round starts with 1.0x multiplier
2. Multiplier increases every 100ms (random acceleration 1.01x - 1.20x)
3. Player must cash out before crash happens
4. If they cash out before crash: win = bet × multiplier
5. If crash happens before cashout: lose bet

**Payouts:**
- Variable (1.1x - 1000x possible)
- House edge via crash probability

**Backend Requirements:**
- New handlers: `POST /api/casino/crash/start`, `POST /api/casino/crash/cashout`
- Game engine: multiplier calculation + crash detection
- Background loop: CrashLoop (100ms tick)
- Table: `casino_crash_games` (start_multiplier, crash_multiplier, cashout_multiplier, payout)
- RTP: 97% (mathematically fair with proper crash rates)

**Frontend Requirements:**
- CrashPanel.tsx component
- Multiplier display (animated growth)
- Real-time price chart (optional)
- Cashout button
- Live players list (show who cashed out, who crashed)
- WebSocket integration

**Estimated:** 7 days

---

## 🎯 Phase 3: Advanced Features (Q3 2026 - June/July)

### Features
- **Tournaments:** Multi-player game competitions with leaderboards
- **Achievements:** Badges for winning streaks, game milestones
- **Multiplayer Games:**
  - Texas Hold'em (full poker)
  - Heads-up Blackjack (player vs real opponent)
  - Card Battle (turn-based)
- **Seasonal Events:** Limited-time games, special bonuses
- **NFT Integration:** Mint winning game results as NFTs (optional)

---

## 📊 Database Schema Additions

### Phase 2 Tables
```sql
-- Dice
CREATE TABLE casino_dice_games (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES casino_players(user_id),
    bet BIGINT NOT NULL,
    player_roll INT NOT NULL,
    house_roll INT NOT NULL,
    winner TEXT, -- 'player', 'house', 'push'
    payout BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Baccarat
CREATE TABLE casino_baccarat_games (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES casino_players(user_id),
    bet BIGINT NOT NULL,
    player_hand JSONB, -- {cards: [...], value: 0-9}
    banker_hand JSONB,
    winner TEXT, -- 'player', 'banker', 'tie'
    payout BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Keno
CREATE TABLE casino_keno_games (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES casino_players(user_id),
    bet BIGINT NOT NULL,
    selected_numbers INT[] NOT NULL,
    winning_numbers INT[] NOT NULL,
    matches INT NOT NULL,
    payout BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Video Poker
CREATE TABLE casino_videopoker_games (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES casino_players(user_id),
    bet BIGINT NOT NULL,
    initial_hand JSONB,
    final_hand JSONB,
    hand_rank TEXT, -- 'royal_flush', 'straight_flush', etc.
    payout BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Crash
CREATE TABLE casino_crash_games (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES casino_players(user_id),
    bet BIGINT NOT NULL,
    crash_at NUMERIC(10,4) NOT NULL,
    cashout_at NUMERIC(10,4),
    payout BIGINT NOT NULL DEFAULT 0,
    status TEXT, -- 'crashed', 'cashed_out'
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 🔄 Migration Strategy

### Per-Game:
1. Add migration file: `services/casino/migrations/00X_<game>_tables.sql`
2. Implement game engine: `services/casino/internal/casino/<game>.go`
3. Add handlers: `services/casino/internal/casino/handlers.go`
4. Implement frontend: `frontend/src/features/casino/<game>/`
5. Update RTK Query: `frontend/src/features/casino/api/casinoApi.ts`
6. Test locally, commit to feature branch, PR to main

---

## 🚀 Deployment Checklist

- [ ] Migration applied to casino DB
- [ ] Backend compiles without errors
- [ ] Frontend builds without errors
- [ ] All game flows tested locally
- [ ] WebSocket tested for real-time updates
- [ ] RTP verified in testing
- [ ] Documentation updated
- [ ] Environment variables configured
- [ ] PR reviewed and approved
- [ ] Merged to main
- [ ] Deployed to staging
- [ ] Smoke tested in staging
- [ ] Deployed to production

---

## 📈 Success Metrics

Track these KPIs:
- **MAU (Monthly Active Users):** Target 500+
- **Game Play Through:** Avg sessions/user/month
- **Win Rate:** Monitor RTP vs expected
- **Retention:** 30-day retention target 40%+
- **Bug Reports:** Target <1 per 1000 plays
- **Load Time:** Game load <2s

---

## 🔗 References

- **MVP commit:** `a8266d1` (2026-04-05)
- **Casino service:** `services/casino/`
- **Frontend:** `frontend/src/features/casino/`
- **DB migrations:** `services/casino/migrations/`
- **Architecture:** `docs/CASINO_DB_ARCHITECTURE.md`

---

## 👥 Team Notes

**Current Focus:** Dice + Baccarat (Easy wins, high engagement)
**Next Sprint:** Keno implementation (parallel with frontend polish)
**Backlog:** Video Poker, Crash, Tournaments

Estimated completion for Phase 2: **End of May 2026**
