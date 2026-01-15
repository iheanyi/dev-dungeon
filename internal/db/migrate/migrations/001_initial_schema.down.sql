-- Rollback: Initial schema for /dev/dungeon multiplayer
-- Version: 001

DROP TABLE IF EXISTS daily_seeds;
DROP TABLE IF EXISTS world_drops;
DROP TABLE IF EXISTS leaderboard_entries;
DROP TABLE IF EXISTS meta_progress;
DROP TABLE IF EXISTS game_saves;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
