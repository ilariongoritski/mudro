-- Down migration for 012_casino.sql
BEGIN;

-- Delete seeded data
DELETE FROM casino_accounts WHERE code IN ('SYSTEM_HOUSE_POOL', 'SYSTEM_SETTLEMENT_POOL');
DELETE FROM casino_rtp_profiles WHERE name IN ('default', 'vip', 'shark');

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS casino_idempotency;
DROP INDEX IF EXISTS idx_casino_rtp_assign_user;
DROP TABLE IF EXISTS casino_rtp_assignments;
DROP TABLE IF EXISTS casino_rtp_profiles;
DROP INDEX IF EXISTS idx_casino_rounds_user;
DROP TABLE IF EXISTS casino_rounds;
DROP INDEX IF EXISTS idx_casino_ledger_account;
DROP INDEX IF EXISTS idx_casino_ledger_transfer;
DROP TABLE IF EXISTS casino_ledger_entries;
DROP TABLE IF EXISTS casino_transfers;
DROP INDEX IF EXISTS idx_casino_accounts_user;
DROP TABLE IF EXISTS casino_accounts;

-- Note: not dropping pgcrypto extension as other things may depend on it

COMMIT;
