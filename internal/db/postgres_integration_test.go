//go:build integration

package db

import (
	"context"
	"os"
	"testing"
	"time"
)

// getTestDatabaseURL returns the database URL for integration tests.
// Uses DATABASE_URL env var, defaulting to a local test database.
func getTestDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:postgres@localhost:5432/devdungeon_test?sslmode=disable"
}

// setupTestDB creates a new client and ensures the schema exists.
func setupTestDB(t *testing.T) (*Client, func()) {
	t.Helper()

	ctx := context.Background()
	client, err := NewClient(ctx, getTestDatabaseURL())
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create schema if not exists
	_, err = client.pool.Exec(ctx, testSchema)
	if err != nil {
		client.Close()
		t.Fatalf("failed to create schema: %v", err)
	}

	// Clean up function
	cleanup := func() {
		// Clean up test data in reverse dependency order
		client.pool.Exec(ctx, "DELETE FROM world_drops")
		client.pool.Exec(ctx, "DELETE FROM leaderboard_entries")
		client.pool.Exec(ctx, "DELETE FROM game_saves")
		client.pool.Exec(ctx, "DELETE FROM web_sessions")
		client.pool.Exec(ctx, "DELETE FROM auth_tokens")
		client.pool.Exec(ctx, "DELETE FROM meta_progress")
		client.pool.Exec(ctx, "DELETE FROM daily_seeds")
		client.pool.Exec(ctx, "DELETE FROM users")
		client.Close()
	}

	return client, cleanup
}

// testSchema contains the SQL to create all required tables
const testSchema = `
-- Enable citext extension for case-insensitive usernames
CREATE EXTENSION IF NOT EXISTS citext;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    username CITEXT UNIQUE NOT NULL,
    public_key_fingerprint VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    is_banned BOOLEAN DEFAULT FALSE
);

-- Game saves (replaces local JSON)
CREATE TABLE IF NOT EXISTS game_saves (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    save_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Meta progression
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
    run_type VARCHAR(32),
    seed BIGINT,
    score INT,
    floors_cleared INT,
    time_seconds INT,
    class VARCHAR(32),
    created_at TIMESTAMP DEFAULT NOW()
);

-- World drops (async multiplayer)
CREATE TABLE IF NOT EXISTS world_drops (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    floor_type VARCHAR(32),
    position_x INT,
    position_y INT,
    drop_type VARCHAR(16),
    content TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Daily seeds
CREATE TABLE IF NOT EXISTS daily_seeds (
    date DATE PRIMARY KEY,
    seed BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Auth tokens (magic links)
CREATE TABLE IF NOT EXISTS auth_tokens (
    token VARCHAR(64) PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE
);

-- Web sessions
CREATE TABLE IF NOT EXISTS web_sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP DEFAULT NOW()
);
`

// --- User Tests ---

func TestPostgres_CreateUser(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	user, err := client.CreateUser(ctx, "testuser", "SHA256:testfingerprint123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
	if user.PublicKeyFingerprint != "SHA256:testfingerprint123" {
		t.Errorf("unexpected fingerprint: %s", user.PublicKeyFingerprint)
	}
	if len(user.NanoID) != 21 {
		t.Errorf("expected 21-char nanoid, got %d chars", len(user.NanoID))
	}
	if user.IsBanned {
		t.Error("new user should not be banned")
	}
}

func TestPostgres_GetUserByFingerprint(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	fingerprint := "SHA256:uniquefingerprint456"

	// Create user first
	created, err := client.CreateUser(ctx, "fingerprintuser", fingerprint)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Retrieve by fingerprint
	user, err := client.GetUserByFingerprint(ctx, fingerprint)
	if err != nil {
		t.Fatalf("GetUserByFingerprint failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, user.ID)
	}

	// Non-existent fingerprint
	user, err = client.GetUserByFingerprint(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent fingerprint")
	}
}

func TestPostgres_GetUserByUsername(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	created, err := client.CreateUser(ctx, "usernametest", "SHA256:usernametestfp")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Case-insensitive lookup (CITEXT)
	user, err := client.GetUserByUsername(ctx, "USERNAMETEST")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, user.ID)
	}

	// Non-existent username
	user, err = client.GetUserByUsername(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent username")
	}
}

func TestPostgres_GetUserByNanoID(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	created, err := client.CreateUser(ctx, "nanoiduser", "SHA256:nanoidfp")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	user, err := client.GetUserByNanoID(ctx, created.NanoID)
	if err != nil {
		t.Fatalf("GetUserByNanoID failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Username != "nanoiduser" {
		t.Errorf("expected username 'nanoiduser', got '%s'", user.Username)
	}

	// Non-existent nanoid
	user, err = client.GetUserByNanoID(ctx, "nonexistentnanoid1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent nanoid")
	}
}

func TestPostgres_UpdateLastLogin(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, err := client.CreateUser(ctx, "loginuser", "SHA256:loginfp")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	initialLogin := user.LastLogin

	time.Sleep(10 * time.Millisecond) // Ensure time difference

	err = client.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		t.Fatalf("UpdateLastLogin failed: %v", err)
	}

	// Verify update
	updated, err := client.GetUserByNanoID(ctx, user.NanoID)
	if err != nil {
		t.Fatalf("GetUserByNanoID failed: %v", err)
	}
	if !updated.LastLogin.After(initialLogin) {
		t.Error("last_login should be updated")
	}
}

func TestPostgres_UpdateUsername(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, err := client.CreateUser(ctx, "oldname", "SHA256:usernamefp")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = client.UpdateUsername(ctx, user.ID, "newname")
	if err != nil {
		t.Fatalf("UpdateUsername failed: %v", err)
	}

	updated, err := client.GetUserByNanoID(ctx, user.NanoID)
	if err != nil {
		t.Fatalf("GetUserByNanoID failed: %v", err)
	}
	if updated.Username != "newname" {
		t.Errorf("expected username 'newname', got '%s'", updated.Username)
	}
}

// --- Game Save Tests ---

func TestPostgres_GameSave_CRUD(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, err := client.CreateUser(ctx, "saveuser", "SHA256:savefp")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Initially no save
	save, err := client.GetGameSave(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetGameSave failed: %v", err)
	}
	if save != nil {
		t.Error("expected no save initially")
	}

	// Create save
	saveData := map[string]interface{}{
		"floor":  5,
		"health": 100,
		"items":  []string{"sword", "potion"},
	}
	err = client.UpsertGameSave(ctx, user.ID, saveData)
	if err != nil {
		t.Fatalf("UpsertGameSave failed: %v", err)
	}

	// Retrieve save
	save, err = client.GetGameSave(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetGameSave failed: %v", err)
	}
	if save == nil {
		t.Fatal("expected save, got nil")
	}
	if save.UserID != user.ID {
		t.Errorf("expected user_id %d, got %d", user.ID, save.UserID)
	}

	// Update save (upsert)
	saveData["floor"] = 10
	err = client.UpsertGameSave(ctx, user.ID, saveData)
	if err != nil {
		t.Fatalf("UpsertGameSave update failed: %v", err)
	}

	// Delete save
	err = client.DeleteGameSave(ctx, user.ID)
	if err != nil {
		t.Fatalf("DeleteGameSave failed: %v", err)
	}

	// Verify deletion
	save, err = client.GetGameSave(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetGameSave after delete failed: %v", err)
	}
	if save != nil {
		t.Error("expected nil after delete")
	}
}

// --- Meta Progress Tests ---

func TestPostgres_MetaProgress_CRUD(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, err := client.CreateUser(ctx, "metauser", "SHA256:metafp")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// CreateUser should create default meta progress
	meta, err := client.GetMetaProgress(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetMetaProgress failed: %v", err)
	}
	if meta == nil {
		t.Fatal("expected default meta progress, got nil")
	}
	if len(meta.UnlockedClasses) != 1 || meta.UnlockedClasses[0] != "init" {
		t.Errorf("expected default class 'init', got %v", meta.UnlockedClasses)
	}

	// Update meta progress
	meta.TotalExitCodes = 50
	meta.UnlockedClasses = []string{"init", "bash", "sudo"}
	meta.RunsCompleted = 10
	meta.DeepestFloor = 7
	meta.TotalDeaths = 25

	err = client.UpdateMetaProgress(ctx, meta)
	if err != nil {
		t.Fatalf("UpdateMetaProgress failed: %v", err)
	}

	// Verify update
	updated, err := client.GetMetaProgress(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetMetaProgress after update failed: %v", err)
	}
	if updated.TotalExitCodes != 50 {
		t.Errorf("expected 50 exit codes, got %d", updated.TotalExitCodes)
	}
	if len(updated.UnlockedClasses) != 3 {
		t.Errorf("expected 3 unlocked classes, got %d", len(updated.UnlockedClasses))
	}
	if updated.DeepestFloor != 7 {
		t.Errorf("expected deepest floor 7, got %d", updated.DeepestFloor)
	}
}

// --- Leaderboard Tests ---

func TestPostgres_Leaderboard(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create users for leaderboard entries
	user1, _ := client.CreateUser(ctx, "player1", "SHA256:fp1")
	user2, _ := client.CreateUser(ctx, "player2", "SHA256:fp2")
	user3, _ := client.CreateUser(ctx, "player3", "SHA256:fp3")

	// Add entries
	entries := []*LeaderboardEntry{
		{UserID: user1.ID, RunType: "standard", Score: 1000, FloorsCleared: 5, Class: "init"},
		{UserID: user2.ID, RunType: "standard", Score: 2000, FloorsCleared: 8, Class: "bash"},
		{UserID: user3.ID, RunType: "daily", Score: 1500, FloorsCleared: 6, Class: "sudo"},
		{UserID: user1.ID, RunType: "standard", Score: 3000, FloorsCleared: 10, Class: "init"},
	}

	for _, entry := range entries {
		err := client.AddLeaderboardEntry(ctx, entry)
		if err != nil {
			t.Fatalf("AddLeaderboardEntry failed: %v", err)
		}
	}

	// Get top scores for standard
	topStandard, err := client.GetTopScores(ctx, "standard", 10)
	if err != nil {
		t.Fatalf("GetTopScores failed: %v", err)
	}
	if len(topStandard) != 3 {
		t.Errorf("expected 3 standard entries, got %d", len(topStandard))
	}
	// Should be sorted by score descending
	if topStandard[0].Score != 3000 {
		t.Errorf("expected top score 3000, got %d", topStandard[0].Score)
	}

	// Get leaderboard with no filter
	all, err := client.GetLeaderboard(ctx, "", 10)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}
	if len(all) != 4 {
		t.Errorf("expected 4 total entries, got %d", len(all))
	}

	// Get leaderboard with filter
	dailyOnly, err := client.GetLeaderboard(ctx, "daily", 10)
	if err != nil {
		t.Fatalf("GetLeaderboard with filter failed: %v", err)
	}
	if len(dailyOnly) != 1 {
		t.Errorf("expected 1 daily entry, got %d", len(dailyOnly))
	}
}

// --- World Drop Tests ---

func TestPostgres_WorldDrops(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user1, _ := client.CreateUser(ctx, "dropper1", "SHA256:dropfp1")
	user2, _ := client.CreateUser(ctx, "dropper2", "SHA256:dropfp2")

	// Create drops
	drop1 := &WorldDrop{
		UserID:    user1.ID,
		FloorType: "home",
		PositionX: 10,
		PositionY: 20,
		DropType:  "message",
		Content:   "Watch out for the daemon!",
	}
	err := client.CreateWorldDrop(ctx, drop1)
	if err != nil {
		t.Fatalf("CreateWorldDrop failed: %v", err)
	}

	// Create drop with custom expiry
	drop2 := &WorldDrop{
		UserID:    user2.ID,
		FloorType: "home",
		PositionX: 15,
		PositionY: 25,
		DropType:  "gravestone",
		Content:   "Here lies user2, killed by fork bomb",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = client.CreateWorldDrop(ctx, drop2)
	if err != nil {
		t.Fatalf("CreateWorldDrop with expiry failed: %v", err)
	}

	// Get random drop (excluding user1)
	randomDrop, err := client.GetRandomDrop(ctx, "home", user1.ID)
	if err != nil {
		t.Fatalf("GetRandomDrop failed: %v", err)
	}
	if randomDrop == nil {
		t.Fatal("expected a drop, got nil")
	}
	if randomDrop.UserID == user1.ID {
		t.Error("should not return drops from excluded user")
	}

	// No drops on different floor
	randomDrop, err = client.GetRandomDrop(ctx, "dev", user1.ID)
	if err != nil {
		t.Fatalf("GetRandomDrop for empty floor failed: %v", err)
	}
	if randomDrop != nil {
		t.Error("expected nil for floor with no drops")
	}
}

func TestPostgres_CleanupExpiredDrops(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, _ := client.CreateUser(ctx, "expiredropper", "SHA256:expiredfp")

	// Create an already-expired drop
	_, err := client.pool.Exec(ctx, `
		INSERT INTO world_drops (nanoid, user_id, floor_type, position_x, position_y, drop_type, content, expires_at)
		VALUES ($1, $2, 'home', 0, 0, 'message', 'expired', NOW() - INTERVAL '1 hour')
	`, GenerateNanoID(), user.ID)
	if err != nil {
		t.Fatalf("failed to insert expired drop: %v", err)
	}

	// Create a valid drop
	drop := &WorldDrop{
		UserID:    user.ID,
		FloorType: "home",
		DropType:  "message",
		Content:   "valid",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	client.CreateWorldDrop(ctx, drop)

	// Cleanup
	deleted, err := client.CleanupExpiredDrops(ctx)
	if err != nil {
		t.Fatalf("CleanupExpiredDrops failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

// --- Daily Seed Tests ---

func TestPostgres_DailySeed(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Get or create today's seed
	seed1, err := client.GetOrCreateDailySeed(ctx)
	if err != nil {
		t.Fatalf("GetOrCreateDailySeed failed: %v", err)
	}
	if seed1 == 0 {
		t.Error("expected non-zero seed")
	}

	// Second call should return same seed
	seed2, err := client.GetOrCreateDailySeed(ctx)
	if err != nil {
		t.Fatalf("GetOrCreateDailySeed second call failed: %v", err)
	}
	if seed2 != seed1 {
		t.Errorf("expected same seed %d, got %d", seed1, seed2)
	}
}

// --- Auth Token Tests ---

func TestPostgres_AuthToken(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, _ := client.CreateUser(ctx, "authuser", "SHA256:authfp")

	// Create token
	token, err := client.CreateAuthToken(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateAuthToken failed: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected 64-char token, got %d chars", len(token))
	}

	// Verify token
	verifiedUser, err := client.VerifyAuthToken(ctx, token)
	if err != nil {
		t.Fatalf("VerifyAuthToken failed: %v", err)
	}
	if verifiedUser == nil {
		t.Fatal("expected user, got nil")
	}
	if verifiedUser.ID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, verifiedUser.ID)
	}

	// Token should be marked as used - second verify should fail
	verifiedUser, err = client.VerifyAuthToken(ctx, token)
	if err != nil {
		t.Fatalf("VerifyAuthToken second call failed: %v", err)
	}
	if verifiedUser != nil {
		t.Error("expected nil for already-used token")
	}

	// Invalid token
	verifiedUser, err = client.VerifyAuthToken(ctx, "invalidtoken")
	if err != nil {
		t.Fatalf("VerifyAuthToken invalid token failed: %v", err)
	}
	if verifiedUser != nil {
		t.Error("expected nil for invalid token")
	}
}

func TestPostgres_CleanupExpiredTokens(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, _ := client.CreateUser(ctx, "tokencleanup", "SHA256:tokenfp")

	// Insert an already-expired token directly (64 chars)
	_, err := client.pool.Exec(ctx, `
		INSERT INTO auth_tokens (token, user_id, expires_at)
		VALUES ($1, $2, NOW() - INTERVAL '1 hour')
	`, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", user.ID)
	if err != nil {
		t.Fatalf("failed to insert expired token: %v", err)
	}

	// Create a valid token
	_, _ = client.CreateAuthToken(ctx, user.ID)

	// Cleanup
	deleted, err := client.CleanupExpiredTokens(ctx)
	if err != nil {
		t.Fatalf("CleanupExpiredTokens failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

// --- Web Session Tests ---

func TestPostgres_WebSession(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, _ := client.CreateUser(ctx, "sessionuser", "SHA256:sessionfp")

	// Create session
	token, err := client.CreateWebSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateWebSession failed: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected 64-char token, got %d chars", len(token))
	}

	// Get session
	sessionUser, err := client.GetWebSession(ctx, token)
	if err != nil {
		t.Fatalf("GetWebSession failed: %v", err)
	}
	if sessionUser == nil {
		t.Fatal("expected user, got nil")
	}
	if sessionUser.ID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, sessionUser.ID)
	}

	// Delete session (logout)
	err = client.DeleteWebSession(ctx, token)
	if err != nil {
		t.Fatalf("DeleteWebSession failed: %v", err)
	}

	// Session should be gone
	sessionUser, err = client.GetWebSession(ctx, token)
	if err != nil {
		t.Fatalf("GetWebSession after delete failed: %v", err)
	}
	if sessionUser != nil {
		t.Error("expected nil after session delete")
	}
}

func TestPostgres_WebSession_Expired(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, _ := client.CreateUser(ctx, "expiredsession", "SHA256:expiredsessionfp")

	// Insert an already-expired session directly
	token := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	_, err := client.pool.Exec(ctx, `
		INSERT INTO web_sessions (token, user_id, expires_at)
		VALUES ($1, $2, NOW() - INTERVAL '1 hour')
	`, token, user.ID)
	if err != nil {
		t.Fatalf("failed to insert expired session: %v", err)
	}

	// Get expired session should return nil
	sessionUser, err := client.GetWebSession(ctx, token)
	if err != nil {
		t.Fatalf("GetWebSession for expired failed: %v", err)
	}
	if sessionUser != nil {
		t.Error("expected nil for expired session")
	}
}

func TestPostgres_CleanupExpiredSessions(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user, _ := client.CreateUser(ctx, "sessioncleanup", "SHA256:sessioncleanupfp")

	// Insert expired session (64 chars)
	_, err := client.pool.Exec(ctx, `
		INSERT INTO web_sessions (token, user_id, expires_at)
		VALUES ($1, $2, NOW() - INTERVAL '1 hour')
	`, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", user.ID)
	if err != nil {
		t.Fatalf("failed to insert expired session: %v", err)
	}

	// Create valid session
	_, _ = client.CreateWebSession(ctx, user.ID)

	// Cleanup
	deleted, err := client.CleanupExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("CleanupExpiredSessions failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

// --- Connection Tests ---

func TestPostgres_Close(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx, getTestDatabaseURL())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	client.Close()

	// After close, operations should fail
	_, err = client.GetUserByUsername(ctx, "anyone")
	if err == nil {
		t.Error("expected error after close")
	}
}
