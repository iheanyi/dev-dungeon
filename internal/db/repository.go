package db

import (
	"context"
	"time"
)

// Repository defines the interface for database operations.
// This allows swapping implementations for testing (in-memory) vs production (PostgreSQL).
type Repository interface {
	// User operations
	GetUserByFingerprint(ctx context.Context, fingerprint string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByNanoID(ctx context.Context, nanoid string) (*User, error)
	CreateUser(ctx context.Context, username, fingerprint string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID int) error
	UpdateUsername(ctx context.Context, userID int, newUsername string) error

	// Game save operations
	GetGameSave(ctx context.Context, userID int) (*GameSave, error)
	UpsertGameSave(ctx context.Context, userID int, saveData interface{}) error
	DeleteGameSave(ctx context.Context, userID int) error

	// Meta progress operations
	GetMetaProgress(ctx context.Context, userID int) (*MetaProgress, error)
	UpdateMetaProgress(ctx context.Context, meta *MetaProgress) error

	// Leaderboard operations
	AddLeaderboardEntry(ctx context.Context, entry *LeaderboardEntry) error
	GetTopScores(ctx context.Context, runType string, limit int) ([]LeaderboardEntry, error)
	GetLeaderboard(ctx context.Context, runType string, limit int) ([]LeaderboardEntry, error)
	GetDailyLeaderboard(ctx context.Context, date time.Time, limit int, cursor *LeaderboardCursor) ([]LeaderboardEntry, *LeaderboardCursor, error)
	GetPlayerDailyRank(ctx context.Context, date time.Time, userID int) (int, *LeaderboardEntry, error)

	// World drop operations
	CreateWorldDrop(ctx context.Context, drop *WorldDrop) error
	GetRandomDrop(ctx context.Context, floorType string, excludeUserID int) (*WorldDrop, error)
	CleanupExpiredDrops(ctx context.Context) (int64, error)

	// Daily seed operations
	GetOrCreateDailySeed(ctx context.Context) (int64, error)
	GetDailySeed(ctx context.Context, date time.Time) (*DailySeed, error)

	// Auth token operations
	CreateAuthToken(ctx context.Context, userID int) (string, error)
	VerifyAuthToken(ctx context.Context, token string) (*User, error)
	CleanupExpiredTokens(ctx context.Context) (int64, error)

	// Web session operations
	CreateWebSession(ctx context.Context, userID int) (string, error)
	GetWebSession(ctx context.Context, token string) (*User, error)
	DeleteWebSession(ctx context.Context, token string) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)

	// Lifecycle
	Close()
}

// Verify Client implements Repository at compile time
var _ Repository = (*Client)(nil)
