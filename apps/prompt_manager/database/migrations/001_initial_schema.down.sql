-- Rollback migration for initial schema
-- Version: 001

-- Drop triggers first
DROP TRIGGER IF EXISTS update_conversation_timestamp;
DROP TRIGGER IF EXISTS update_conversation_stats;

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_start_time;
DROP INDEX IF EXISTS idx_sessions_session_id;
DROP INDEX IF EXISTS idx_ratings_message_id;
DROP INDEX IF EXISTS idx_ratings_conversation_id;
DROP INDEX IF EXISTS idx_messages_timestamp;
DROP INDEX IF EXISTS idx_messages_conversation_id;
DROP INDEX IF EXISTS idx_conversations_created_at;
DROP INDEX IF EXISTS idx_conversations_session_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS conversation_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS ratings;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS conversations;