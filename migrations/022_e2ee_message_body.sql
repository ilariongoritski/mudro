-- Add E2EE columns to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS encrypted_body TEXT;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS nonce TEXT;

-- Index for future encrypted body search (if needed via trigram on encrypted?) - usually not useful for search
-- But good for audit
COMMENT ON COLUMN messages.encrypted_body IS 'Base64 encoded XSalsa20-Poly1305 cyphertext';
COMMENT ON COLUMN messages.nonce IS 'Base64 encoded 24-byte nonce';
