-- Remove unique constraint on (user_id, seed)
DROP INDEX IF EXISTS idx_leaderboard_user_seed;
