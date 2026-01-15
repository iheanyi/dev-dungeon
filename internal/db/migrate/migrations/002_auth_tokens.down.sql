-- Rollback: Auth tokens for magic link authentication
-- Version: 002

DROP TABLE IF EXISTS web_sessions;
DROP TABLE IF EXISTS auth_tokens;
