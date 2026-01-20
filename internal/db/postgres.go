package db

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUsernameTaken is returned when attempting to create a user with a username that already exists.
var ErrUsernameTaken = errors.New("username already taken")

// Client provides database operations for /dev/dungeon.
type Client struct {
	pool *pgxpool.Pool
}

// NewClient creates a new database client.
func NewClient(ctx context.Context, databaseURL string) (*Client, error) {
	return NewClientWithOptions(ctx, databaseURL, false)
}

// NewClientWithOptions creates a new database client with SSL enforcement option.
// When requireSSL is true, connections without SSL will be rejected.
func NewClientWithOptions(ctx context.Context, databaseURL string, requireSSL bool) (*Client, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Check SSL configuration
	if config.ConnConfig.TLSConfig == nil {
		if requireSSL {
			return nil, fmt.Errorf("SSL is required but sslmode=disable in connection string")
		}
		// Log warning for non-SSL connections
		fmt.Println("[WARN] Database connection is NOT using SSL - do not use in production!")
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
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

// Pool returns the underlying connection pool.
// Used for migrations and other low-level operations.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
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
// Uses a transaction to ensure atomicity between user creation and meta_progress insertion.
func (c *Client) CreateUser(ctx context.Context, username, fingerprint string) (*User, error) {
	nanoid := GenerateNanoID()
	now := time.Now().UTC()

	// Use a transaction to ensure user and meta_progress are created atomically
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var user User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (nanoid, username, public_key_fingerprint, created_at, last_login)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
	`, nanoid, username, fingerprint, now).Scan(
		&user.ID, &user.NanoID, &user.Username, &user.PublicKeyFingerprint,
		&user.CreatedAt, &user.LastLogin, &user.IsBanned,
	)
	if err != nil {
		// Check for unique constraint violation (username already taken)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default meta progress within the same transaction
	_, err = tx.Exec(ctx, `
		INSERT INTO meta_progress (user_id, unlocked_classes, unlocked_items)
		VALUES ($1, ARRAY['init'], ARRAY[]::TEXT[])
	`, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create meta progress: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}

// UpdateLastLogin updates the user's last login timestamp.
func (c *Client) UpdateLastLogin(ctx context.Context, userID int) error {
	_, err := c.pool.Exec(ctx, `
		UPDATE users SET last_login = $1 WHERE id = $2
	`, time.Now().UTC(), userID)
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

	now := time.Now().UTC()
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

// AddLeaderboardEntry adds or updates a score entry.
// Uses UPSERT: if (user_id, seed) already exists, only update if new score is higher.
func (c *Client) AddLeaderboardEntry(ctx context.Context, entry *LeaderboardEntry) error {
	entry.NanoID = GenerateNanoID()
	entry.CreatedAt = time.Now().UTC()

	_, err := c.pool.Exec(ctx, `
		INSERT INTO leaderboard_entries
		(nanoid, user_id, run_type, seed, score, floors_cleared, time_seconds, class, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, seed) DO UPDATE SET
			score = EXCLUDED.score,
			floors_cleared = EXCLUDED.floors_cleared,
			time_seconds = EXCLUDED.time_seconds,
			class = EXCLUDED.class,
			created_at = EXCLUDED.created_at
		WHERE EXCLUDED.score > leaderboard_entries.score
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

	entries := []LeaderboardEntry{} // Initialize as empty slice, not nil
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
	drop.CreatedAt = time.Now().UTC()
	if drop.ExpiresAt.IsZero() {
		drop.ExpiresAt = time.Now().UTC().Add(48 * time.Hour)
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
// Uses INSERT ON CONFLICT to prevent TOCTOU race conditions when multiple
// concurrent requests try to create the seed for the same day.
func (c *Client) GetOrCreateDailySeed(ctx context.Context) (int64, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Generate cryptographically random seed (unpredictable)
	var seedBytes [8]byte
	if _, err := crypto_rand.Read(seedBytes[:]); err != nil {
		return 0, fmt.Errorf("failed to generate random seed: %w", err)
	}
	candidateSeed := int64(binary.BigEndian.Uint64(seedBytes[:]))

	// Use INSERT ON CONFLICT to atomically get-or-create the daily seed.
	// If the row already exists (conflict on date), DO NOTHING and then
	// SELECT the existing seed. This prevents TOCTOU race conditions.
	var seed int64
	err := c.pool.QueryRow(ctx, `
		INSERT INTO daily_seeds (date, seed, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (date) DO NOTHING
		RETURNING seed
	`, today, candidateSeed, time.Now().UTC()).Scan(&seed)

	if errors.Is(err, pgx.ErrNoRows) {
		// Row already existed (conflict), fetch the existing seed
		err = c.pool.QueryRow(ctx, `
			SELECT seed FROM daily_seeds WHERE date = $1
		`, today).Scan(&seed)
		if err != nil {
			return 0, fmt.Errorf("failed to get existing daily seed: %w", err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to create daily seed: %w", err)
	}

	return seed, nil
}

// GetDailySeed retrieves the seed for a specific date.
// Returns nil if no seed exists for that date.
func (c *Client) GetDailySeed(ctx context.Context, date time.Time) (*DailySeed, error) {
	normalizedDate := date.UTC().Truncate(24 * time.Hour)

	var ds DailySeed
	err := c.pool.QueryRow(ctx, `
		SELECT date, seed, created_at FROM daily_seeds WHERE date = $1
	`, normalizedDate).Scan(&ds.Date, &ds.Seed, &ds.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get daily seed: %w", err)
	}
	return &ds, nil
}

// GetUserByNanoID retrieves a user by their public NanoID.
func (c *Client) GetUserByNanoID(ctx context.Context, nanoid string) (*User, error) {
	var user User
	err := c.pool.QueryRow(ctx, `
		SELECT id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
		FROM users
		WHERE nanoid = $1
	`, nanoid).Scan(
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

// --- Auth Token Operations ---

// CreateAuthToken creates a new magic link token for browser authentication.
// Tokens expire after 5 minutes and can only be used once.
func (c *Client) CreateAuthToken(ctx context.Context, userID int) (string, error) {
	token, err := GenerateAuthToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(5 * time.Minute)
	_, err = c.pool.Exec(ctx, `
		INSERT INTO auth_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create auth token: %w", err)
	}

	return token, nil
}

// VerifyAuthToken verifies a token and marks it as used.
// Returns the user if valid, nil if invalid/expired/used.
func (c *Client) VerifyAuthToken(ctx context.Context, token string) (*User, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get and lock the token row
	var userID int
	var expiresAt time.Time
	var used bool
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at, used
		FROM auth_tokens
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&userID, &expiresAt, &used)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Check if token is valid
	if used || time.Now().UTC().After(expiresAt) {
		return nil, nil // Token already used or expired
	}

	// Mark token as used
	_, err = tx.Exec(ctx, `UPDATE auth_tokens SET used = TRUE WHERE token = $1`, token)
	if err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Get the user
	var user User
	err = tx.QueryRow(ctx, `
		SELECT id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.NanoID, &user.Username, &user.PublicKeyFingerprint,
		&user.CreatedAt, &user.LastLogin, &user.IsBanned,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}

// CleanupExpiredTokens removes expired auth tokens.
func (c *Client) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := c.pool.Exec(ctx, `DELETE FROM auth_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// --- Web Session Operations ---

// CreateWebSession creates a new browser session for a user.
// Sessions expire after 7 days.
func (c *Client) CreateWebSession(ctx context.Context, userID int) (string, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour) // 7 days
	_, err = c.pool.Exec(ctx, `
		INSERT INTO web_sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create web session: %w", err)
	}

	return token, nil
}

// GetWebSession retrieves and validates a web session.
// Returns the user if valid, nil if invalid/expired.
// Also updates last_used_at on successful validation.
func (c *Client) GetWebSession(ctx context.Context, token string) (*User, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get session and check expiry
	var userID int
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at
		FROM web_sessions
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&userID, &expiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if expired
	if time.Now().UTC().After(expiresAt) {
		// Delete expired session
		_, _ = tx.Exec(ctx, `DELETE FROM web_sessions WHERE token = $1`, token)
		_ = tx.Commit(ctx)
		return nil, nil
	}

	// Update last used time
	_, err = tx.Exec(ctx, `
		UPDATE web_sessions SET last_used_at = NOW() WHERE token = $1
	`, token)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Get the user
	var user User
	err = tx.QueryRow(ctx, `
		SELECT id, nanoid, username, public_key_fingerprint, created_at, last_login, is_banned
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.NanoID, &user.Username, &user.PublicKeyFingerprint,
		&user.CreatedAt, &user.LastLogin, &user.IsBanned,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}

// DeleteWebSession removes a web session (logout).
func (c *Client) DeleteWebSession(ctx context.Context, token string) error {
	_, err := c.pool.Exec(ctx, `DELETE FROM web_sessions WHERE token = $1`, token)
	return err
}

// CleanupExpiredSessions removes expired web sessions.
func (c *Client) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result, err := c.pool.Exec(ctx, `DELETE FROM web_sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// GetLeaderboard retrieves leaderboard entries, optionally filtered by run type.
func (c *Client) GetLeaderboard(ctx context.Context, runType string, limit int) ([]LeaderboardEntry, error) {
	var query string
	var args []interface{}

	if runType == "" {
		query = `
			SELECT le.id, le.nanoid, le.user_id, u.username, le.run_type, le.seed,
			       le.score, le.floors_cleared, le.time_seconds, le.class, le.created_at
			FROM leaderboard_entries le
			JOIN users u ON le.user_id = u.id
			ORDER BY le.score DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	} else {
		query = `
			SELECT le.id, le.nanoid, le.user_id, u.username, le.run_type, le.seed,
			       le.score, le.floors_cleared, le.time_seconds, le.class, le.created_at
			FROM leaderboard_entries le
			JOIN users u ON le.user_id = u.id
			WHERE le.run_type = $1
			ORDER BY le.score DESC
			LIMIT $2
		`
		args = []interface{}{runType, limit}
	}

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LeaderboardEntry{} // Initialize as empty slice, not nil
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

// GetDailyLeaderboard retrieves daily leaderboard entries for a specific date with cursor-based pagination.
// Returns entries with their calculated rank, and a cursor for fetching the next page.
func (c *Client) GetDailyLeaderboard(ctx context.Context, date time.Time, limit int, cursor *LeaderboardCursor) ([]LeaderboardEntry, *LeaderboardCursor, error) {
	normalizedDate := date.UTC().Truncate(24 * time.Hour)

	// Get the seed for this date
	var seed int64
	err := c.pool.QueryRow(ctx, `SELECT seed FROM daily_seeds WHERE date = $1`, normalizedDate).Scan(&seed)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil // No daily challenge for this date
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get daily seed: %w", err)
	}

	var query string
	var args []interface{}

	if cursor == nil {
		// First page - no cursor
		query = `
			WITH ranked AS (
				SELECT le.id, le.nanoid, le.user_id, u.username, le.run_type, le.seed,
				       le.score, le.floors_cleared, le.time_seconds, le.class, le.created_at,
				       RANK() OVER (ORDER BY le.score DESC) as rank
				FROM leaderboard_entries le
				JOIN users u ON le.user_id = u.id
				WHERE le.run_type = 'daily' AND le.seed = $1
			)
			SELECT id, nanoid, user_id, username, run_type, seed, score,
			       floors_cleared, time_seconds, class, created_at, rank
			FROM ranked
			ORDER BY score DESC, id ASC
			LIMIT $2
		`
		args = []interface{}{seed, limit}
	} else {
		// Subsequent pages - use cursor for stable pagination
		query = `
			WITH ranked AS (
				SELECT le.id, le.nanoid, le.user_id, u.username, le.run_type, le.seed,
				       le.score, le.floors_cleared, le.time_seconds, le.class, le.created_at,
				       RANK() OVER (ORDER BY le.score DESC) as rank
				FROM leaderboard_entries le
				JOIN users u ON le.user_id = u.id
				WHERE le.run_type = 'daily' AND le.seed = $1
			)
			SELECT id, nanoid, user_id, username, run_type, seed, score,
			       floors_cleared, time_seconds, class, created_at, rank
			FROM ranked
			WHERE (score < $2) OR (score = $2 AND id > $3)
			ORDER BY score DESC, id ASC
			LIMIT $4
		`
		args = []interface{}{seed, cursor.Score, cursor.ID, limit}
	}

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query daily leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		err := rows.Scan(
			&e.ID, &e.NanoID, &e.UserID, &e.Username, &e.RunType, &e.Seed,
			&e.Score, &e.FloorsCleared, &e.TimeSeconds, &e.Class, &e.CreatedAt, &e.Rank,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Build next cursor from last entry
	var nextCursor *LeaderboardCursor
	if len(entries) == limit {
		last := entries[len(entries)-1]
		nextCursor = &LeaderboardCursor{
			Score: last.Score,
			ID:    last.ID,
		}
	}

	return entries, nextCursor, nil
}

// GetPlayerDailyRank retrieves a player's rank and entry for a specific day's leaderboard.
// Returns (0, nil, nil) if the player has no entry for that day or no daily challenge exists.
func (c *Client) GetPlayerDailyRank(ctx context.Context, date time.Time, userID int) (int, *LeaderboardEntry, error) {
	normalizedDate := date.UTC().Truncate(24 * time.Hour)

	// Get the seed for this date
	var seed int64
	err := c.pool.QueryRow(ctx, `SELECT seed FROM daily_seeds WHERE date = $1`, normalizedDate).Scan(&seed)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil, nil // No daily challenge for this date
	}
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get daily seed: %w", err)
	}

	query := `
		WITH ranked AS (
			SELECT le.id, le.nanoid, le.user_id, u.username, le.run_type, le.seed,
			       le.score, le.floors_cleared, le.time_seconds, le.class, le.created_at,
			       RANK() OVER (ORDER BY le.score DESC) as rank
			FROM leaderboard_entries le
			JOIN users u ON le.user_id = u.id
			WHERE le.run_type = 'daily' AND le.seed = $1
		)
		SELECT id, nanoid, user_id, username, run_type, seed, score,
		       floors_cleared, time_seconds, class, created_at, rank
		FROM ranked
		WHERE user_id = $2
		ORDER BY score DESC
		LIMIT 1
	`

	var e LeaderboardEntry
	err = c.pool.QueryRow(ctx, query, seed, userID).Scan(
		&e.ID, &e.NanoID, &e.UserID, &e.Username, &e.RunType, &e.Seed,
		&e.Score, &e.FloorsCleared, &e.TimeSeconds, &e.Class, &e.CreatedAt, &e.Rank,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil, nil // Player has no entry for this day
	}
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get player daily rank: %w", err)
	}

	return e.Rank, &e, nil
}
