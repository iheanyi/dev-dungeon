-- Auth tokens for magic link authentication
-- Allows SSH users to authenticate their browser session via one-time tokens

CREATE TABLE IF NOT EXISTS auth_tokens (
    token VARCHAR(64) PRIMARY KEY,  -- 256-bit entropy (64 hex chars)
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE
);

-- Index for cleanup of expired tokens
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires ON auth_tokens(expires_at);

-- Index for finding tokens by user (for cleanup/listing)
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user ON auth_tokens(user_id);

-- Web sessions for authenticated browser access
-- Created when a magic link is verified successfully

CREATE TABLE IF NOT EXISTS web_sessions (
    token VARCHAR(64) PRIMARY KEY,  -- 256-bit session token (64 hex chars)
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for cleanup of expired sessions
CREATE INDEX IF NOT EXISTS idx_web_sessions_expires ON web_sessions(expires_at);

-- Index for finding sessions by user
CREATE INDEX IF NOT EXISTS idx_web_sessions_user ON web_sessions(user_id);
