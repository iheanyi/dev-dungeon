package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client provides database operations for /dev/dungeon.
type Client struct {
	pool *pgxpool.Pool
}

// NewClient creates a new database client.
func NewClient(ctx context.Context, databaseURL string) (*Client, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{pool: pool}, nil
}

// Close closes the database connection pool.
func (c *Client) Close() {
	c.pool.Close()
}

// --- User Operations ---

// GetUserByFingerprint retrieves a user by their SSH public key fingerprint.
func (c *Client) GetUserByFingerprint(ctx context.Context, fingerprint string) (*User, error) {
	var user User
	err := c.pool.QueryRow(ctx, `
		SELECT id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
		FROM users
		WHERE public_key_fingerprint = $1
	`, fingerprint).Scan(
		&user.ID, &user.NanoID, &user.Username, &user.PublicKeyFingerprint,
		&user.CreatedAt, &user.LastLogin, &user.IsBanned,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username (case-insensitive).
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := c.pool.QueryRow(ctx, `
		SELECT id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
		FROM users
		WHERE username = $1
	`, username).Scan(
		&user.ID, &user.NanoID, &user.Username, &user.PublicKeyFingerprint,
		&user.CreatedAt, &user.LastLogin, &user.IsBanned,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// CreateUser creates a new user account.
func (c *Client) CreateUser(ctx context.Context, username, fingerprint string) (*User, error) {
	nanoid := GenerateNanoID()
	now := time.Now()

	var user User
	err := c.pool.QueryRow(ctx, `
		INSERT INTO users (nanoid, username, public_key_fingerprint, created_at, last_login)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
	`, nanoid, username, fingerprint, now).Scan(
		&user.ID, &user.NanoID, &user.Username, &user.PublicKeyFingerprint,
		&user.CreatedAt, &user.LastLogin, &user.IsBanned,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default meta progress
	_, err = c.pool.Exec(ctx, `
		INSERT INTO meta_progress (user_id, unlocked_classes, unlocked_items)
		VALUES ($1, ARRAY['init'], ARRAY[]::TEXT[])
	`, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create meta progress: %w", err)
	}

	return &user, nil
}

// UpdateLastLogin updates the user's last login timestamp.
func (c *Client) UpdateLastLogin(ctx context.Context, userID int) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE users SET last_login = $1 WHERE id = $2
	`, time.Now(), userID)
	return err
}

// UpdateUsername changes a user's username.
func (c *Client) UpdateUsername(ctx context.Context, userID int, newUsername string) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE users SET username = $1 WHERE id = $2
	`, newUsername, userID)
	return err
}

// --- Game Save Operations ---

// GetGameSave retrieves the user's current game save.
func (c *Client) GetGameSave(ctx context.Context, userID int) (*GameSave, error) {
	var save GameSave
	err := c.pool.QueryRow(ctx, `
		SELECT id, nanoid, user_id, save_data, created_at, updated_at
		FROM game_saves
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, userID).Scan(
		&save.ID, &save.NanoID, &save.UserID, &save.SaveData,
		&save.CreatedAt, &save.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game save: %w", err)
	}
	return &save, nil
}

// UpsertGameSave creates or updates the user's game save.
func (c *Client) UpsertGameSave(ctx context.Context, userID int, saveData interface{}) error {
	data, err := json.Marshal(saveData)
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	now := time.Now()
	_, err = c.pool.Exec(ctx, `
		INSERT INTO game_saves (nanoid, user_id, save_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET save_data = EXCLUDED.save_data, updated_at = EXCLUDED.updated_at
	`, GenerateNanoID(), userID, data, now)
	if err != nil {
		return fmt.Errorf("failed to upsert game save: %w", err)
	}
	return nil
}

// DeleteGameSave removes the user's game save (on death/victory).
func (c *Client) DeleteGameSave(ctx context.Context, userID int) error {
	_, err := c.pool.Exec(ctx, `DELETE FROM game_saves WHERE user_id = $1`, userID)
	return err
}

// --- Meta Progress Operations ---

// GetMetaProgress retrieves the user's permanent progress.
func (c *Client) GetMetaProgress(ctx context.Context, userID int) (*MetaProgress, error) {
	var meta MetaProgress
	err := c.pool.QueryRow(ctx, `
		SELECT user_id, total_exit_codes, unlocked_classes, permanent_bonuses,
		       unlocked_items, runs_completed, deepest_floor, total_deaths
		FROM meta_progress
		WHERE user_id = $1
	`, userID).Scan(
		&meta.UserID, &meta.TotalExitCodes, &meta.UnlockedClasses, &meta.PermanentBonuses,
		&meta.UnlockedItems, &meta.RunsCompleted, &meta.DeepestFloor, &meta.TotalDeaths,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get meta progress: %w", err)
	}
	return &meta, nil
}

// UpdateMetaProgress updates the user's permanent progress.
func (c *Client) UpdateMetaProgress(ctx context.Context, meta *MetaProgress) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE meta_progress
		SET total_exit_codes = $2, unlocked_classes = $3, permanent_bonuses = $4,
		    unlocked_items = $5, runs_completed = $6, deepest_floor = $7, total_deaths = $8
		WHERE user_id = $1
	`, meta.UserID, meta.TotalExitCodes, meta.UnlockedClasses, meta.PermanentBonuses,
		meta.UnlockedItems, meta.RunsCompleted, meta.DeepestFloor, meta.TotalDeaths)
	return err
}

// --- Leaderboard Operations ---

// AddLeaderboardEntry adds a new score entry.
func (c *Client) AddLeaderboardEntry(ctx context.Context, entry *LeaderboardEntry) error {
	entry.NanoID = GenerateNanoID()
	entry.CreatedAt = time.Now()

	_, err := c.pool.Exec(ctx, `
		INSERT INTO leaderboard_entries
		(nanoid, user_id, run_type, seed, score, floors_cleared, time_seconds, class, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, entry.NanoID, entry.UserID, entry.RunType, entry.Seed, entry.Score,
		entry.FloorsCleared, entry.TimeSeconds, entry.Class, entry.CreatedAt)
	return err
}

// GetTopScores retrieves the top N scores for a run type.
func (c *Client) GetTopScores(ctx context.Context, runType string, limit int) ([]LeaderboardEntry, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT le.id, le.nanoid, le.user_id, u.username, le.run_type, le.seed,
		       le.score, le.floors_cleared, le.time_seconds, le.class, le.created_at
		FROM leaderboard_entries le
		JOIN users u ON le.user_id = u.id
		WHERE le.run_type = $1
		ORDER BY le.score DESC
		LIMIT $2
	`, runType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		err := rows.Scan(
			&e.ID, &e.NanoID, &e.UserID, &e.Username, &e.RunType, &e.Seed,
			&e.Score, &e.FloorsCleared, &e.TimeSeconds, &e.Class, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// --- World Drop Operations ---

// CreateWorldDrop creates a new async drop.
func (c *Client) CreateWorldDrop(ctx context.Context, drop *WorldDrop) error {
	drop.NanoID = GenerateNanoID()
	drop.CreatedAt = time.Now()
	if drop.ExpiresAt.IsZero() {
		drop.ExpiresAt = time.Now().Add(48 * time.Hour)
	}

	_, err := c.pool.Exec(ctx, `
		INSERT INTO world_drops
		(nanoid, user_id, floor_type, position_x, position_y, drop_type, content, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, drop.NanoID, drop.UserID, drop.FloorType, drop.PositionX, drop.PositionY,
		drop.DropType, drop.Content, drop.ExpiresAt, drop.CreatedAt)
	return err
}

// GetRandomDrop retrieves a random non-expired drop for the given floor.
func (c *Client) GetRandomDrop(ctx context.Context, floorType string, excludeUserID int) (*WorldDrop, error) {
	var drop WorldDrop
	err := c.pool.QueryRow(ctx, `
		SELECT wd.id, wd.nanoid, wd.user_id, u.username, wd.floor_type,
		       wd.position_x, wd.position_y, wd.drop_type, wd.content, wd.expires_at, wd.created_at
		FROM world_drops wd
		JOIN users u ON wd.user_id = u.id
		WHERE wd.floor_type = $1
		  AND wd.user_id != $2
		  AND wd.expires_at > NOW()
		ORDER BY RANDOM()
		LIMIT 1
	`, floorType, excludeUserID).Scan(
		&drop.ID, &drop.NanoID, &drop.UserID, &drop.Username, &drop.FloorType,
		&drop.PositionX, &drop.PositionY, &drop.DropType, &drop.Content,
		&drop.ExpiresAt, &drop.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &drop, nil
}

// CleanupExpiredDrops removes expired world drops.
func (c *Client) CleanupExpiredDrops(ctx context.Context) (int64, error) {
	result, err := c.pool.Exec(ctx, `DELETE FROM world_drops WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// --- Daily Seed Operations ---

// GetOrCreateDailySeed gets today's seed, creating if needed.
func (c *Client) GetOrCreateDailySeed(ctx context.Context) (int64, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	var seed int64
	err := c.pool.QueryRow(ctx, `
		SELECT seed FROM daily_seeds WHERE date = $1
	`, today).Scan(&seed)

	if errors.Is(err, pgx.ErrNoRows) {
		// Generate new seed from date
		seed = today.UnixNano()
		_, err = c.pool.Exec(ctx, `
			INSERT INTO daily_seeds (date, seed, created_at)
			VALUES ($1, $2, $3)
		`, today, seed, time.Now())
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	return seed, nil
}
