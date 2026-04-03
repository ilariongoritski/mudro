-- Down migration for 016_chat.sql
BEGIN;

DROP INDEX IF EXISTS idx_chat_messages_room_id_desc;
DROP TABLE IF EXISTS chat_messages;

COMMIT;
