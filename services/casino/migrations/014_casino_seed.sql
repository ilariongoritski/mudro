-- Seed casino_config with default values for 5-reel slots
-- Symbols: cherry, lemon, bar, seven, diamond
-- Weights control frequency (higher = more common)
-- Paytable is the multiplier for 5-of-a-kind (4-of-a-kind = 60%, 3-of-a-kind = 30%, 2-of-a-kind = 1x)

insert into casino_config (id, rtp_percent, initial_balance, symbol_weights, paytable, updated_at)
values (
  true,
  95.0,
  500,
  '{"cherry": 30, "lemon": 25, "bar": 20, "seven": 15, "diamond": 10}'::jsonb,
  '{"cherry": 3, "lemon": 5, "bar": 10, "seven": 25, "diamond": 100}'::jsonb,
  now()
)
on conflict (id) do update set
  rtp_percent = excluded.rtp_percent,
  initial_balance = excluded.initial_balance,
  symbol_weights = excluded.symbol_weights,
  paytable = excluded.paytable,
  updated_at = now();
