# Casino Smoke Checklist

Минимальный прогон казино перед показом или релизом.

## 1. Автотесты

```bash
go test ./services/casino/internal/casino
cd frontend && npm run lint && npm run test && npm run build
```

Если есть изолированная мигрированная casino DB:

```bash
set MUDRO_CASINO_INTEGRATION_TEST_DSN=postgres://...
go test -run Integration -v ./services/casino/internal/casino
```

## 2. Обязательные ручные сценарии

1. Slots:
   - открыть баланс
   - сделать spin
   - проверить обновление истории

2. Roulette:
   - instant spin на `red`
   - instant spin на `green`
   - live round: ставка, lock, result, settlement

3. Plinko:
   - drop на `low`, `medium`, `high`
   - проверить path, payout и баланс

4. Blackjack:
   - start
   - `hit`
   - `stand`
   - проверить resolved state и payout

5. Bonus:
   - открыть bonus state
   - проверить сценарий без Telegram init data
   - проверить claim с валидной Telegram WebApp сессией

## 3. Денежный контур

Проверить после каждого сценария:

1. Баланс пользователя в casino UI.
2. Запись в `casino_game_activity`.
3. Связанные переводы в `casino_transfers` и `casino_ledger_entries`.
4. Отсутствие рассинхрона между `casino_players.balance` и account-проводками.

## 4. Wallet sync

Если включён `CASINO_MAIN_DSN`, отдельно проверить:

1. После ставки появляется задача в `casino_balance_sync_queue`.
2. Reconciler вычищает очередь без `failed`.
3. `GetBalanceDrift` не показывает дрейф после settle.
4. При искусственном рассинхроне projection восстанавливается.

## 5. Что не показывать

- `crash`
- `coinflip`
- bonus claim без живой Telegram-конфигурации
- admin/config flow без подготовленного admin-пользователя
