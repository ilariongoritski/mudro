-- 012_casino.sql — Casino microservice tables

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Casino accounts (wallets)
CREATE TABLE IF NOT EXISTS casino_accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    type        TEXT NOT NULL DEFAULT 'user' CHECK (type IN ('user','system')),
    code        TEXT UNIQUE NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'МДР',
    balance     NUMERIC(30,10) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_casino_accounts_user ON casino_accounts(user_id);

-- Ledger (append-only double-entry)
CREATE TABLE IF NOT EXISTS casino_transfers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind        TEXT NOT NULL CHECK (kind IN ('bet_stake','bet_payout','deposit','withdrawal','adjustment')),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS casino_ledger_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES casino_transfers(id) ON DELETE CASCADE,
    account_id  UUID NOT NULL REFERENCES casino_accounts(id) ON DELETE RESTRICT,
    direction   TEXT NOT NULL CHECK (direction IN ('debit','credit')),
    amount      NUMERIC(30,10) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_casino_ledger_transfer ON casino_ledger_entries(transfer_id);
CREATE INDEX IF NOT EXISTS idx_casino_ledger_account ON casino_ledger_entries(account_id);

-- Game rounds (provably fair)
CREATE TABLE IF NOT EXISTS casino_rounds (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    server_seed      TEXT NOT NULL,
    server_seed_hash TEXT NOT NULL,
    client_seed      TEXT,
    nonce            INT NOT NULL DEFAULT 0,
    round_hash       TEXT,
    roll             INT,
    bet_amount       NUMERIC(30,10),
    payout_amount    NUMERIC(30,10),
    multiplier       NUMERIC(10,4),
    tier_label       TEXT,
    status           TEXT NOT NULL DEFAULT 'prepared' CHECK (status IN ('prepared','resolved','cancelled')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_casino_rounds_user ON casino_rounds(user_id, status);

-- RTP profiles
CREATE TABLE IF NOT EXISTS casino_rtp_profiles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT UNIQUE NOT NULL,
    rtp        NUMERIC(5,2) NOT NULL,
    paytable   JSONB NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- RTP assignments (user -> profile)
CREATE TABLE IF NOT EXISTS casino_rtp_assignments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rtp_profile_id  UUID NOT NULL REFERENCES casino_rtp_profiles(id) ON DELETE CASCADE,
    assigned_by     TEXT NOT NULL DEFAULT 'system',
    reason          TEXT,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, rtp_profile_id)
);
CREATE INDEX IF NOT EXISTS idx_casino_rtp_assign_user ON casino_rtp_assignments(user_id);

-- Idempotency keys
CREATE TABLE IF NOT EXISTS casino_idempotency (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key          TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'processing' CHECK (status IN ('processing','succeeded','failed')),
    response     JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, key)
);

-- Seed default RTP profiles
INSERT INTO casino_rtp_profiles (name, rtp, paytable, is_default) VALUES
('default', 96, '[{"minRoll":0,"maxRoll":0,"multiplier":25,"label":"МЕГА ДЖЕКПОТ","symbol":"👑👑👑"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔥🔥🔥"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🚀🚀🚀"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"⭐⭐⭐"},{"minRoll":38,"maxRoll":52,"multiplier":0.6,"label":"МЕЛОЧЬ","symbol":"🍀🍀🍀"},{"minRoll":53,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💀"}]'::jsonb, true)
ON CONFLICT (name) DO NOTHING;

INSERT INTO casino_rtp_profiles (name, rtp, paytable, is_default) VALUES
('vip', 97, '[{"minRoll":0,"maxRoll":0,"multiplier":30,"label":"МЕГА ДЖЕКПОТ","symbol":"👑👑👑"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔥🔥🔥"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🚀🚀🚀"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"⭐⭐⭐"},{"minRoll":38,"maxRoll":52,"multiplier":0.333,"label":"МЕЛОЧЬ","symbol":"🍀🍀🍀"},{"minRoll":53,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💀"}]'::jsonb, false)
ON CONFLICT (name) DO NOTHING;

INSERT INTO casino_rtp_profiles (name, rtp, paytable, is_default) VALUES
('shark', 94, '[{"minRoll":0,"maxRoll":0,"multiplier":20,"label":"МЕГА ДЖЕКПОТ","symbol":"👑👑👑"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔥🔥🔥"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🚀🚀🚀"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"⭐⭐⭐"},{"minRoll":38,"maxRoll":57,"multiplier":0.6,"label":"МЕЛОЧЬ","symbol":"🍀🍀🍀"},{"minRoll":58,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💀"}]'::jsonb, false)
ON CONFLICT (name) DO NOTHING;

-- Seed system accounts
INSERT INTO casino_accounts (type, code, currency, balance) VALUES
('system', 'SYSTEM_HOUSE_POOL', 'МДР', 1000000),
('system', 'SYSTEM_SETTLEMENT_POOL', 'МДР', 1000000)
ON CONFLICT (code) DO NOTHING;
