-- Migration: Initial schema for /dev/dungeon multiplayer
-- Version: 001

-- Enable citext extension for case-insensitive usernames
CREATE EXTENSION IF NOT EXISTS citext;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    username CITEXT UNIQUE NOT NULL,
    public_key_fingerprint VARCHAR(128) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_banned BOOLEAN DEFAULT FALSE
);

-- Index for fast fingerprint lookups (SSH auth)
CREATE INDEX IF NOT EXISTS idx_users_fingerprint ON users(public_key_fingerprint);

-- Game saves (one active save per user)
CREATE TABLE IF NOT EXISTS game_saves (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    save_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Meta progression (permanent unlocks)
CREATE TABLE IF NOT EXISTS meta_progress (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_exit_codes INT DEFAULT 0,
    unlocked_classes TEXT[] DEFAULT ARRAY['init'],
    permanent_bonuses JSONB DEFAULT '{}',
    unlocked_items TEXT[] DEFAULT ARRAY[]::TEXT[],
    runs_completed INT DEFAULT 0,
    deepest_floor INT DEFAULT 0,
    total_deaths INT DEFAULT 0
);

-- Leaderboards
CREATE TABLE IF NOT EXISTS leaderboard_entries (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    run_type VARCHAR(32) NOT NULL,  -- 'standard', 'daily', 'seeded'
    seed BIGINT,
    score INT NOT NULL,
    floors_cleared INT NOT NULL,
    time_seconds INT,
    class VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for leaderboard queries
CREATE INDEX IF NOT EXISTS idx_leaderboard_type_score ON leaderboard_entries(run_type, score DESC);

-- Async drops (messages/items left by players)
CREATE TABLE IF NOT EXISTS world_drops (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    floor_type VARCHAR(32) NOT NULL,
    position_x INT NOT NULL,
    position_y INT NOT NULL,
    drop_type VARCHAR(16) NOT NULL,  -- 'message', 'item', 'gravestone'
    content TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for finding drops on a floor
CREATE INDEX IF NOT EXISTS idx_drops_floor ON world_drops(floor_type, expires_at);

-- Daily seeds for daily runs
CREATE TABLE IF NOT EXISTS daily_seeds (
    date DATE PRIMARY KEY,
    seed BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
