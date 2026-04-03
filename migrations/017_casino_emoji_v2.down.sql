-- Down migration for 017_casino_emoji_v2.sql
BEGIN;

-- Drop the added column
ALTER TABLE casino_rounds DROP COLUMN IF EXISTS tier_symbol;

-- Revert paytable to original emoji symbols
UPDATE casino_rtp_profiles
SET paytable = '[{"minRoll":0,"maxRoll":0,"multiplier":25,"label":"МЕГА ДЖЕКПОТ","symbol":"👑👑👑"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔥🔥🔥"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🚀🚀🚀"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"⭐⭐⭐"},{"minRoll":38,"maxRoll":52,"multiplier":0.6,"label":"МЕЛОЧЬ","symbol":"🍀🍀🍀"},{"minRoll":53,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💀"}]'::jsonb
WHERE name = 'default';

UPDATE casino_rtp_profiles
SET paytable = '[{"minRoll":0,"maxRoll":0,"multiplier":30,"label":"МЕГА ДЖЕКПОТ","symbol":"👑👑👑"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔥🔥🔥"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🚀🚀🚀"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"⭐⭐⭐"},{"minRoll":38,"maxRoll":52,"multiplier":0.333,"label":"МЕЛОЧЬ","symbol":"🍀🍀🍀"},{"minRoll":53,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💀"}]'::jsonb
WHERE name = 'vip';

UPDATE casino_rtp_profiles
SET paytable = '[{"minRoll":0,"maxRoll":0,"multiplier":20,"label":"МЕГА ДЖЕКПОТ","symbol":"👑👑👑"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔥🔥🔥"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🚀🚀🚀"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"⭐⭐⭐"},{"minRoll":38,"maxRoll":57,"multiplier":0.6,"label":"МЕЛОЧЬ","symbol":"🍀🍀🍀"},{"minRoll":58,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💀"}]'::jsonb
WHERE name = 'shark';

COMMIT;
