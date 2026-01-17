-- Add unique constraint on (user_id, seed) for leaderboard entries.
-- This ensures only one entry per user per run (identified by seed).
-- For daily runs (shared seed), each user gets one entry per day.
-- For standard runs (unique seed), duplicate submissions are prevented.

-- Step 1: Delete duplicate entries, keeping only the one with the highest score
-- for each (user_id, seed) combination.
DELETE FROM leaderboard_entries
WHERE id NOT IN (
    SELECT DISTINCT ON (user_id, seed) id
    FROM leaderboard_entries
    ORDER BY user_id, seed, score DESC, created_at DESC
);

-- Step 2: Create the unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_leaderboard_user_seed
ON leaderboard_entries(user_id, seed);
