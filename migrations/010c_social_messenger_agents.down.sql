-- Down migration for 010_social_messenger_agents.sql
BEGIN;

-- 6. AGENTS and related (reverse order)
DROP INDEX IF EXISTS agent_messages_from_agent_idx;
DROP INDEX IF EXISTS agent_messages_task_id_idx;
DROP TABLE IF EXISTS agent_messages;

DROP INDEX IF EXISTS agent_task_events_agent_id_idx;
ALTER TABLE agent_task_events DROP COLUMN IF EXISTS agent_id;

DROP INDEX IF EXISTS agent_queue_agent_id_idx;
ALTER TABLE agent_queue DROP COLUMN IF EXISTS agent_id;

DROP TABLE IF EXISTS agents;

-- 5. NOTIFICATIONS
DROP INDEX IF EXISTS notifications_user_id_created_idx;
DROP INDEX IF EXISTS notifications_user_unread_idx;
DROP TABLE IF EXISTS notifications;

-- 4. CONVERSATIONS & MESSAGES
DROP TABLE IF EXISTS message_reads;
DROP INDEX IF EXISTS messages_sender_id_idx;
DROP INDEX IF EXISTS messages_conversation_created_idx;
DROP TABLE IF EXISTS messages;
DROP INDEX IF EXISTS conversation_members_user_id_idx;
DROP TABLE IF EXISTS conversation_members;
DROP TABLE IF EXISTS conversations;

-- 3. NATIVE POSTS — revert author_id and source check
DROP INDEX IF EXISTS posts_author_id_idx;
ALTER TABLE posts DROP COLUMN IF EXISTS author_id;

ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_source_check;
ALTER TABLE posts ADD CONSTRAINT posts_source_check CHECK (source IN ('vk','tg'));

-- 2. FOLLOWS
DROP INDEX IF EXISTS follows_following_id_idx;
DROP TABLE IF EXISTS follows;

-- 1. PROFILES — revert user columns and accounts.user_id
DROP INDEX IF EXISTS accounts_user_id_idx;
ALTER TABLE accounts DROP COLUMN IF EXISTS user_id;

ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;

COMMIT;
