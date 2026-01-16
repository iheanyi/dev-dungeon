package db

import (
	"context"
	"testing"
	"time"
)

func TestGenerateNanoID(t *testing.T) {
	id := GenerateNanoID()

	if len(id) != 21 {
		t.Errorf("NanoID should be 21 chars, got %d", len(id))
	}

	// Should only contain alphanumeric lowercase
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
			t.Errorf("NanoID contains invalid character: %c", c)
		}
	}
}

func TestGenerateNanoIDUniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateNanoID()
		if ids[id] {
			t.Errorf("Duplicate NanoID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateAuthToken(t *testing.T) {
	token, err := GenerateAuthToken()
	if err != nil {
		t.Fatalf("GenerateAuthToken failed: %v", err)
	}

	if len(token) != 64 {
		t.Errorf("Auth token should be 64 chars (256-bit hex), got %d", len(token))
	}

	// Should be valid hex
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Auth token contains invalid hex character: %c", c)
		}
	}
}

func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	if len(token) != 64 {
		t.Errorf("Session token should be 64 chars, got %d", len(token))
	}
}

// === Memory Repository Tests ===

func TestMemoryRepository_CreateUser(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, err := repo.CreateUser(ctx, "testuser", "SHA256:abc123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
	if user.PublicKeyFingerprint != "SHA256:abc123" {
		t.Errorf("expected fingerprint 'SHA256:abc123', got '%s'", user.PublicKeyFingerprint)
	}
	if len(user.NanoID) != 21 {
		t.Errorf("expected NanoID length 21, got %d", len(user.NanoID))
	}
	if user.IsBanned {
		t.Error("new user should not be banned")
	}
}

func TestMemoryRepository_CreateUser_DuplicateUsername(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create first user
	_, err := repo.CreateUser(ctx, "testuser", "SHA256:abc123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Try to create second user with same username (different fingerprint)
	_, err = repo.CreateUser(ctx, "testuser", "SHA256:different456")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	if err != ErrUsernameTaken {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}

	// Case-insensitive check
	_, err = repo.CreateUser(ctx, "TestUser", "SHA256:another789")
	if err == nil {
		t.Fatal("expected error for duplicate username (case-insensitive)")
	}
	if err != ErrUsernameTaken {
		t.Errorf("expected ErrUsernameTaken for case-insensitive match, got %v", err)
	}
}

func TestMemoryRepository_GetUserByFingerprint(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create user
	created, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	// Get by fingerprint
	user, err := repo.GetUserByFingerprint(ctx, "SHA256:abc123")
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
	user, err = repo.GetUserByFingerprint(ctx, "SHA256:nonexistent")
	if err != nil {
		t.Fatalf("GetUserByFingerprint failed: %v", err)
	}
	if user != nil {
		t.Error("expected nil for non-existent fingerprint")
	}
}

func TestMemoryRepository_GetUserByUsername(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	repo.CreateUser(ctx, "TestUser", "SHA256:abc123")

	// Case-insensitive lookup
	user, _ := repo.GetUserByUsername(ctx, "testuser")
	if user == nil {
		t.Fatal("expected user with lowercase lookup")
	}

	user, _ = repo.GetUserByUsername(ctx, "TESTUSER")
	if user == nil {
		t.Fatal("expected user with uppercase lookup")
	}

	user, _ = repo.GetUserByUsername(ctx, "nonexistent")
	if user != nil {
		t.Error("expected nil for non-existent username")
	}
}

func TestMemoryRepository_GetUserByNanoID(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	created, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	user, _ := repo.GetUserByNanoID(ctx, created.NanoID)
	if user == nil {
		t.Fatal("expected user by NanoID")
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
}

func TestMemoryRepository_UpdateUsername(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "oldname", "SHA256:abc123")

	err := repo.UpdateUsername(ctx, user.ID, "newname")
	if err != nil {
		t.Fatalf("UpdateUsername failed: %v", err)
	}

	// Old username should not work
	oldUser, _ := repo.GetUserByUsername(ctx, "oldname")
	if oldUser != nil {
		t.Error("old username should not find user")
	}

	// New username should work
	newUser, _ := repo.GetUserByUsername(ctx, "newname")
	if newUser == nil {
		t.Fatal("new username should find user")
	}
	if newUser.Username != "newname" {
		t.Errorf("expected 'newname', got '%s'", newUser.Username)
	}
}

func TestMemoryRepository_UpdateLastLogin(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")
	initialLogin := user.LastLogin

	time.Sleep(10 * time.Millisecond)
	repo.UpdateLastLogin(ctx, user.ID)

	updated, _ := repo.GetUserByFingerprint(ctx, "SHA256:abc123")
	if !updated.LastLogin.After(initialLogin) {
		t.Error("LastLogin should have been updated")
	}
}

func TestMemoryRepository_GameSave(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	// Initially no save
	save, _ := repo.GetGameSave(ctx, user.ID)
	if save != nil {
		t.Error("expected no save initially")
	}

	// Create save
	saveData := map[string]interface{}{"floor": 3, "hp": 100}
	err := repo.UpsertGameSave(ctx, user.ID, saveData)
	if err != nil {
		t.Fatalf("UpsertGameSave failed: %v", err)
	}

	// Get save
	save, _ = repo.GetGameSave(ctx, user.ID)
	if save == nil {
		t.Fatal("expected save after upsert")
	}
	if save.UserID != user.ID {
		t.Errorf("expected UserID %d, got %d", user.ID, save.UserID)
	}

	// Update save
	saveData["floor"] = 5
	repo.UpsertGameSave(ctx, user.ID, saveData)

	// Delete save
	repo.DeleteGameSave(ctx, user.ID)
	save, _ = repo.GetGameSave(ctx, user.ID)
	if save != nil {
		t.Error("expected no save after delete")
	}
}

func TestMemoryRepository_MetaProgress(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	// Default meta progress created with user
	meta, _ := repo.GetMetaProgress(ctx, user.ID)
	if meta == nil {
		t.Fatal("expected default meta progress")
	}
	if len(meta.UnlockedClasses) != 1 || meta.UnlockedClasses[0] != "init" {
		t.Error("expected 'init' class unlocked by default")
	}

	// Update meta progress
	meta.TotalExitCodes = 100
	meta.UnlockedClasses = append(meta.UnlockedClasses, "bash")
	repo.UpdateMetaProgress(ctx, meta)

	// Verify update
	updated, _ := repo.GetMetaProgress(ctx, user.ID)
	if updated.TotalExitCodes != 100 {
		t.Errorf("expected 100 exit codes, got %d", updated.TotalExitCodes)
	}
	if len(updated.UnlockedClasses) != 2 {
		t.Errorf("expected 2 unlocked classes, got %d", len(updated.UnlockedClasses))
	}
}

func TestMemoryRepository_Leaderboard(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user1, _ := repo.CreateUser(ctx, "player1", "SHA256:111")
	user2, _ := repo.CreateUser(ctx, "player2", "SHA256:222")

	// Add entries
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:        user1.ID,
		RunType:       "standard",
		Score:         1000,
		FloorsCleared: 8,
		Class:         "sudo",
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:        user2.ID,
		RunType:       "standard",
		Score:         500,
		FloorsCleared: 5,
		Class:         "bash",
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:        user1.ID,
		RunType:       "daily",
		Score:         800,
		FloorsCleared: 7,
		Class:         "vim",
	})

	// Get all leaderboard
	entries, _ := repo.GetLeaderboard(ctx, "", 10)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Should be sorted by score descending
	if entries[0].Score != 1000 {
		t.Errorf("expected top score 1000, got %d", entries[0].Score)
	}

	// Filter by run type
	entries, _ = repo.GetLeaderboard(ctx, "standard", 10)
	if len(entries) != 2 {
		t.Errorf("expected 2 standard entries, got %d", len(entries))
	}

	entries, _ = repo.GetLeaderboard(ctx, "daily", 10)
	if len(entries) != 1 {
		t.Errorf("expected 1 daily entry, got %d", len(entries))
	}

	// Test limit
	entries, _ = repo.GetLeaderboard(ctx, "", 1)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry with limit, got %d", len(entries))
	}
}

func TestMemoryRepository_WorldDrop(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user1, _ := repo.CreateUser(ctx, "player1", "SHA256:111")
	user2, _ := repo.CreateUser(ctx, "player2", "SHA256:222")

	// Create drop
	drop := &WorldDrop{
		UserID:    user1.ID,
		FloorType: "home",
		PositionX: 10,
		PositionY: 20,
		DropType:  "message",
		Content:   "Beware the daemon!",
	}
	repo.CreateWorldDrop(ctx, drop)

	// User2 should see the drop
	found, _ := repo.GetRandomDrop(ctx, "home", user2.ID)
	if found == nil {
		t.Fatal("expected to find drop")
	}
	if found.Content != "Beware the daemon!" {
		t.Errorf("expected content 'Beware the daemon!', got '%s'", found.Content)
	}

	// User1 should NOT see their own drop
	found, _ = repo.GetRandomDrop(ctx, "home", user1.ID)
	if found != nil {
		t.Error("user should not see their own drop")
	}

	// Wrong floor should not find drop
	found, _ = repo.GetRandomDrop(ctx, "tmp", user2.ID)
	if found != nil {
		t.Error("should not find drop on wrong floor")
	}
}

func TestMemoryRepository_DailySeed(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	seed1, err := repo.GetOrCreateDailySeed(ctx)
	if err != nil {
		t.Fatalf("GetOrCreateDailySeed failed: %v", err)
	}
	if seed1 == 0 {
		t.Error("seed should not be zero")
	}

	// Should return same seed on second call
	seed2, _ := repo.GetOrCreateDailySeed(ctx)
	if seed1 != seed2 {
		t.Errorf("expected same seed, got %d and %d", seed1, seed2)
	}
}

func TestMemoryRepository_AuthToken(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	// Create token
	token, err := repo.CreateAuthToken(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateAuthToken failed: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected 64-char token, got %d", len(token))
	}

	// Verify token
	verified, _ := repo.VerifyAuthToken(ctx, token)
	if verified == nil {
		t.Fatal("expected verified user")
	}
	if verified.ID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, verified.ID)
	}

	// Token should be marked as used now
	verified, _ = repo.VerifyAuthToken(ctx, token)
	if verified != nil {
		t.Error("token should not verify twice")
	}

	// Invalid token
	verified, _ = repo.VerifyAuthToken(ctx, "invalidtoken")
	if verified != nil {
		t.Error("invalid token should return nil")
	}
}

func TestMemoryRepository_WebSession(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	// Create session
	token, err := repo.CreateWebSession(ctx, user.ID)
	if err != nil {
		t.Fatalf("CreateWebSession failed: %v", err)
	}

	// Get session
	sessionUser, _ := repo.GetWebSession(ctx, token)
	if sessionUser == nil {
		t.Fatal("expected session user")
	}
	if sessionUser.ID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, sessionUser.ID)
	}

	// Delete session
	repo.DeleteWebSession(ctx, token)
	sessionUser, _ = repo.GetWebSession(ctx, token)
	if sessionUser != nil {
		t.Error("session should be deleted")
	}
}

func TestMemoryRepository_CleanupExpiredDrops(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "SHA256:abc123")

	// Create an expired drop
	drop := &WorldDrop{
		UserID:    user.ID,
		FloorType: "home",
		DropType:  "message",
		Content:   "Old message",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	repo.CreateWorldDrop(ctx, drop)

	// Create a valid drop
	validDrop := &WorldDrop{
		UserID:    user.ID,
		FloorType: "home",
		DropType:  "message",
		Content:   "New message",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	repo.CreateWorldDrop(ctx, validDrop)

	// Cleanup
	count, _ := repo.CleanupExpiredDrops(ctx)
	if count != 1 {
		t.Errorf("expected 1 expired drop cleaned, got %d", count)
	}
}

// === Model Struct Tests ===

func TestUserStruct(t *testing.T) {
	user := User{
		ID:                   1,
		NanoID:               "abc123",
		Username:             "testuser",
		PublicKeyFingerprint: "SHA256:xyz",
		CreatedAt:            time.Now(),
		LastLogin:            time.Now(),
		IsBanned:             false,
	}

	if user.ID != 1 {
		t.Error("User ID not set correctly")
	}
}

func TestGameSaveStruct(t *testing.T) {
	save := GameSave{
		ID:        1,
		NanoID:    "abc123",
		UserID:    1,
		SaveData:  []byte(`{"test": true}`),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if save.UserID != 1 {
		t.Error("GameSave UserID not set correctly")
	}
}

func TestLeaderboardEntryStruct(t *testing.T) {
	entry := LeaderboardEntry{
		ID:            1,
		UserID:        1,
		Username:      "testuser",
		RunType:       "standard",
		Score:         1000,
		FloorsCleared: 8,
		TimeSeconds:   3600,
		Class:         "sudo",
	}

	if entry.Score != 1000 {
		t.Error("LeaderboardEntry Score not set correctly")
	}
}

func TestWorldDropStruct(t *testing.T) {
	drop := WorldDrop{
		ID:        1,
		UserID:    1,
		FloorType: "home",
		PositionX: 10,
		PositionY: 20,
		DropType:  "message",
		Content:   "Hello!",
	}

	if drop.DropType != "message" {
		t.Error("WorldDrop DropType not set correctly")
	}
}

func TestDailySeedStruct(t *testing.T) {
	seed := DailySeed{
		Date:      time.Now(),
		Seed:      12345,
		CreatedAt: time.Now(),
	}

	if seed.Seed != 12345 {
		t.Error("DailySeed Seed not set correctly")
	}
}

func TestAuthTokenStruct(t *testing.T) {
	token := AuthToken{
		Token:     "abc123",
		UserID:    1,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	if token.Used {
		t.Error("AuthToken should not be used initially")
	}
}

func TestWebSessionStruct(t *testing.T) {
	session := WebSession{
		Token:      "abc123",
		UserID:     1,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		LastUsedAt: time.Now(),
	}

	if session.UserID != 1 {
		t.Error("WebSession UserID not set correctly")
	}
}

// === Additional Coverage Tests ===

func TestMemoryRepository_Close(t *testing.T) {
	repo := NewMemoryRepository()

	// Close should be a no-op and not panic
	repo.Close()

	// Should be able to call multiple times
	repo.Close()
}

func TestMemoryRepository_GetTopScores(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create users
	user1, _ := repo.CreateUser(ctx, "player1", "fp1")
	user2, _ := repo.CreateUser(ctx, "player2", "fp2")
	user3, _ := repo.CreateUser(ctx, "player3", "fp3")

	// Add leaderboard entries with different scores
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user1.ID,
		RunType: "standard",
		Score:   500,
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user2.ID,
		RunType: "standard",
		Score:   1000,
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user3.ID,
		RunType: "standard",
		Score:   750,
	})
	// Add a daily run entry (different type)
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user1.ID,
		RunType: "daily",
		Score:   200,
	})

	// Get top scores for standard run type
	scores, err := repo.GetTopScores(ctx, "standard", 10)
	if err != nil {
		t.Fatalf("GetTopScores failed: %v", err)
	}

	if len(scores) != 3 {
		t.Errorf("expected 3 scores, got %d", len(scores))
	}

	// Should be sorted by score descending
	if scores[0].Score != 1000 {
		t.Errorf("expected top score 1000, got %d", scores[0].Score)
	}
	if scores[1].Score != 750 {
		t.Errorf("expected second score 750, got %d", scores[1].Score)
	}
	if scores[2].Score != 500 {
		t.Errorf("expected third score 500, got %d", scores[2].Score)
	}
}

func TestMemoryRepository_GetTopScoresWithLimit(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create user and add many entries
	user, _ := repo.CreateUser(ctx, "player", "fp")
	for i := 0; i < 20; i++ {
		repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
			UserID:  user.ID,
			RunType: "standard",
			Score:   i * 100,
		})
	}

	// Get top 5 only
	scores, err := repo.GetTopScores(ctx, "standard", 5)
	if err != nil {
		t.Fatalf("GetTopScores failed: %v", err)
	}

	if len(scores) != 5 {
		t.Errorf("expected 5 scores with limit, got %d", len(scores))
	}

	// Highest should be 1900 (19 * 100)
	if scores[0].Score != 1900 {
		t.Errorf("expected top score 1900, got %d", scores[0].Score)
	}
}

func TestMemoryRepository_CleanupExpiredTokens(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create user
	user, _ := repo.CreateUser(ctx, "testuser", "fp")

	// Create some auth tokens
	token1, _ := repo.CreateAuthToken(ctx, user.ID)
	token2, _ := repo.CreateAuthToken(ctx, user.ID)

	// Manually expire the tokens by manipulating the internal state
	// (In production this would happen over time)
	repo.mu.Lock()
	for _, data := range repo.authTokens {
		data.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	}
	repo.mu.Unlock()

	// Cleanup should remove expired tokens
	count, err := repo.CleanupExpiredTokens(ctx)
	if err != nil {
		t.Fatalf("CleanupExpiredTokens failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 tokens cleaned up, got %d", count)
	}

	// Tokens should no longer verify
	verifiedUser, _ := repo.VerifyAuthToken(ctx, token1)
	if verifiedUser != nil {
		t.Error("expired token should not verify")
	}
	verifiedUser, _ = repo.VerifyAuthToken(ctx, token2)
	if verifiedUser != nil {
		t.Error("expired token should not verify")
	}
}

func TestMemoryRepository_CleanupExpiredSessions(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create user
	user, _ := repo.CreateUser(ctx, "testuser", "fp")

	// Create some web sessions
	session1, _ := repo.CreateWebSession(ctx, user.ID)
	session2, _ := repo.CreateWebSession(ctx, user.ID)

	// Manually expire the sessions
	repo.mu.Lock()
	for _, data := range repo.webSessions {
		data.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	}
	repo.mu.Unlock()

	// Cleanup should remove expired sessions
	count, err := repo.CleanupExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("CleanupExpiredSessions failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 sessions cleaned up, got %d", count)
	}

	// Sessions should no longer return user
	sessionUser, _ := repo.GetWebSession(ctx, session1)
	if sessionUser != nil {
		t.Error("expired session should not return user")
	}
	sessionUser, _ = repo.GetWebSession(ctx, session2)
	if sessionUser != nil {
		t.Error("expired session should not return user")
	}
}

func TestMemoryRepository_GetUserByNanoID_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Try to get non-existent user
	user, err := repo.GetUserByNanoID(ctx, "nonexistent123456789")
	if err != nil {
		t.Fatalf("GetUserByNanoID should not error: %v", err)
	}

	if user != nil {
		t.Error("should return nil for non-existent NanoID")
	}
}

func TestMemoryRepository_GetMetaProgress_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Try to get meta progress for non-existent user
	meta, err := repo.GetMetaProgress(ctx, 99999)
	if err != nil {
		t.Fatalf("GetMetaProgress should not error: %v", err)
	}

	if meta != nil {
		t.Error("should return nil for non-existent user")
	}
}

func TestMemoryRepository_UpdateUsername_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Try to update non-existent user
	err := repo.UpdateUsername(ctx, 99999, "newname")
	if err != nil {
		t.Fatalf("UpdateUsername should not error for non-existent user: %v", err)
	}
}

func TestMemoryRepository_UpsertGameSave_Update(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create user
	user, _ := repo.CreateUser(ctx, "testuser", "fp")

	// Create initial save
	saveData1 := map[string]interface{}{"level": 1}
	err := repo.UpsertGameSave(ctx, user.ID, saveData1)
	if err != nil {
		t.Fatalf("Initial UpsertGameSave failed: %v", err)
	}

	// Update the save
	saveData2 := map[string]interface{}{"level": 5}
	err = repo.UpsertGameSave(ctx, user.ID, saveData2)
	if err != nil {
		t.Fatalf("Update UpsertGameSave failed: %v", err)
	}

	// Should only have one save entry
	save, _ := repo.GetGameSave(ctx, user.ID)
	if save == nil {
		t.Fatal("save should exist")
	}

	// ID should remain the same (it was updated, not created)
	if save.ID != 1 {
		t.Errorf("save ID should remain 1, got %d", save.ID)
	}
}

func TestMemoryRepository_AuthToken_AlreadyUsed(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "fp")
	token, _ := repo.CreateAuthToken(ctx, user.ID)

	// First verification should succeed
	verified1, _ := repo.VerifyAuthToken(ctx, token)
	if verified1 == nil {
		t.Fatal("first verification should succeed")
	}

	// Second verification should fail (token already used)
	verified2, _ := repo.VerifyAuthToken(ctx, token)
	if verified2 != nil {
		t.Error("second verification should fail - token already used")
	}
}

func TestMemoryRepository_WebSession_Expired(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "fp")
	sessionToken, _ := repo.CreateWebSession(ctx, user.ID)

	// Manually expire the session
	repo.mu.Lock()
	if data, ok := repo.webSessions[sessionToken]; ok {
		data.ExpiresAt = time.Now().Add(-1 * time.Hour)
	}
	repo.mu.Unlock()

	// Should not return user for expired session
	sessionUser, _ := repo.GetWebSession(ctx, sessionToken)
	if sessionUser != nil {
		t.Error("expired session should not return user")
	}

	// Session should be deleted
	repo.mu.RLock()
	_, exists := repo.webSessions[sessionToken]
	repo.mu.RUnlock()
	if exists {
		t.Error("expired session should be deleted when accessed")
	}
}

func TestMemoryRepository_WebSession_UserDeleted(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "fp")
	sessionToken, _ := repo.CreateWebSession(ctx, user.ID)

	// Manually remove user (simulate deletion)
	repo.mu.Lock()
	delete(repo.users, user.ID)
	repo.mu.Unlock()

	// Should not return user for session with deleted user
	sessionUser, _ := repo.GetWebSession(ctx, sessionToken)
	if sessionUser != nil {
		t.Error("session for deleted user should not return user")
	}
}

func TestMemoryRepository_AuthToken_UserDeleted(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "testuser", "fp")
	token, _ := repo.CreateAuthToken(ctx, user.ID)

	// Manually remove user (simulate deletion)
	repo.mu.Lock()
	delete(repo.users, user.ID)
	repo.mu.Unlock()

	// Should not return user for token with deleted user
	verified, _ := repo.VerifyAuthToken(ctx, token)
	if verified != nil {
		t.Error("token for deleted user should not verify")
	}
}

// === Daily Leaderboard Tests ===

func TestMemoryRepository_GetDailySeed(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Get seed for today (should not exist initially)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	seed, err := repo.GetDailySeed(ctx, today)
	if err != nil {
		t.Fatalf("GetDailySeed failed: %v", err)
	}
	if seed != nil {
		t.Error("expected nil for non-existent seed")
	}

	// Create seed via GetOrCreate
	createdSeed, err := repo.GetOrCreateDailySeed(ctx)
	if err != nil {
		t.Fatalf("GetOrCreateDailySeed failed: %v", err)
	}

	// Now GetDailySeed should return it
	seed, err = repo.GetDailySeed(ctx, today)
	if err != nil {
		t.Fatalf("GetDailySeed failed: %v", err)
	}
	if seed == nil {
		t.Fatal("expected seed after creation")
	}
	if seed.Seed != createdSeed {
		t.Errorf("expected seed %d, got %d", createdSeed, seed.Seed)
	}
}

func TestMemoryRepository_DailySeedRandomness(t *testing.T) {
	// Two different repositories should generate different seeds
	// (demonstrating randomness, not date-based)
	repo1 := NewMemoryRepository()
	repo2 := NewMemoryRepository()
	ctx := context.Background()

	seed1, _ := repo1.GetOrCreateDailySeed(ctx)
	seed2, _ := repo2.GetOrCreateDailySeed(ctx)

	// Seeds should be different (with overwhelming probability)
	// Note: there's a 1 in 2^64 chance they match, which is acceptable
	if seed1 == seed2 {
		t.Log("Warning: seeds matched (extremely unlikely if truly random)")
	}

	// Seeds should not be simple date-based values
	today := time.Now().UTC().Truncate(24 * time.Hour)
	dateSeed := today.UnixNano()
	if seed1 == dateSeed {
		t.Error("seed should not be simple date.UnixNano()")
	}
}

func TestMemoryRepository_GetDailyLeaderboard(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create users
	user1, _ := repo.CreateUser(ctx, "player1", "SHA256:key1")
	user2, _ := repo.CreateUser(ctx, "player2", "SHA256:key2")
	user3, _ := repo.CreateUser(ctx, "player3", "SHA256:key3")

	// Get today's seed
	seed, _ := repo.GetOrCreateDailySeed(ctx)

	// Add daily leaderboard entries
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:        user1.ID,
		RunType:       "daily",
		Seed:          seed,
		Score:         5000,
		FloorsCleared: 5,
		Class:         "init",
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:        user2.ID,
		RunType:       "daily",
		Seed:          seed,
		Score:         8000,
		FloorsCleared: 8,
		Class:         "cron",
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:        user3.ID,
		RunType:       "daily",
		Seed:          seed,
		Score:         3000,
		FloorsCleared: 3,
		Class:         "bash",
	})

	// Get daily leaderboard
	today := time.Now().UTC().Truncate(24 * time.Hour)
	entries, nextCursor, err := repo.GetDailyLeaderboard(ctx, today, 10, nil)
	if err != nil {
		t.Fatalf("GetDailyLeaderboard failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Should be sorted by score descending
	if entries[0].Score != 8000 {
		t.Errorf("expected first entry score 8000, got %d", entries[0].Score)
	}
	if entries[0].Rank != 1 {
		t.Errorf("expected first entry rank 1, got %d", entries[0].Rank)
	}
	if entries[1].Score != 5000 {
		t.Errorf("expected second entry score 5000, got %d", entries[1].Score)
	}
	if entries[2].Score != 3000 {
		t.Errorf("expected third entry score 3000, got %d", entries[2].Score)
	}

	// No next cursor since we got all entries
	if nextCursor != nil {
		t.Error("expected no next cursor when all entries fit")
	}
}

func TestMemoryRepository_GetDailyLeaderboard_Pagination(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create users and seed
	seed, _ := repo.GetOrCreateDailySeed(ctx)

	// Add 5 entries
	for i := 1; i <= 5; i++ {
		user, _ := repo.CreateUser(ctx, "player"+string(rune('0'+i)), "SHA256:key"+string(rune('0'+i)))
		repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
			UserID:        user.ID,
			RunType:       "daily",
			Seed:          seed,
			Score:         i * 1000, // 1000, 2000, 3000, 4000, 5000
			FloorsCleared: i,
			Class:         "init",
		})
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Get first page (limit 2)
	entries, cursor, err := repo.GetDailyLeaderboard(ctx, today, 2, nil)
	if err != nil {
		t.Fatalf("GetDailyLeaderboard page 1 failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries on page 1, got %d", len(entries))
	}
	if entries[0].Score != 5000 {
		t.Errorf("expected first entry score 5000, got %d", entries[0].Score)
	}
	if cursor == nil {
		t.Fatal("expected cursor for next page")
	}

	// Get second page
	entries, cursor, err = repo.GetDailyLeaderboard(ctx, today, 2, cursor)
	if err != nil {
		t.Fatalf("GetDailyLeaderboard page 2 failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries on page 2, got %d", len(entries))
	}
	if entries[0].Score != 3000 {
		t.Errorf("expected page 2 first entry score 3000, got %d", entries[0].Score)
	}
}

func TestMemoryRepository_GetDailyLeaderboard_NoSeed(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Get leaderboard for a date with no seed
	yesterday := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)
	entries, cursor, err := repo.GetDailyLeaderboard(ctx, yesterday, 10, nil)
	if err != nil {
		t.Fatalf("GetDailyLeaderboard failed: %v", err)
	}
	if entries != nil {
		t.Error("expected nil entries for non-existent seed")
	}
	if cursor != nil {
		t.Error("expected nil cursor for non-existent seed")
	}
}

func TestMemoryRepository_GetPlayerDailyRank(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create users
	user1, _ := repo.CreateUser(ctx, "player1", "SHA256:key1")
	user2, _ := repo.CreateUser(ctx, "player2", "SHA256:key2")
	user3, _ := repo.CreateUser(ctx, "player3", "SHA256:key3")

	seed, _ := repo.GetOrCreateDailySeed(ctx)

	// Add entries with different scores
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user1.ID,
		RunType: "daily",
		Seed:    seed,
		Score:   5000,
		Class:   "init",
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user2.ID,
		RunType: "daily",
		Seed:    seed,
		Score:   8000, // highest
		Class:   "cron",
	})
	repo.AddLeaderboardEntry(ctx, &LeaderboardEntry{
		UserID:  user3.ID,
		RunType: "daily",
		Seed:    seed,
		Score:   3000,
		Class:   "bash",
	})

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Check rank for user2 (should be #1)
	rank, entry, err := repo.GetPlayerDailyRank(ctx, today, user2.ID)
	if err != nil {
		t.Fatalf("GetPlayerDailyRank failed: %v", err)
	}
	if rank != 1 {
		t.Errorf("expected rank 1, got %d", rank)
	}
	if entry == nil {
		t.Fatal("expected entry")
	}
	if entry.Score != 8000 {
		t.Errorf("expected score 8000, got %d", entry.Score)
	}

	// Check rank for user1 (should be #2)
	rank, entry, _ = repo.GetPlayerDailyRank(ctx, today, user1.ID)
	if rank != 2 {
		t.Errorf("expected rank 2, got %d", rank)
	}

	// Check rank for user3 (should be #3)
	rank, entry, _ = repo.GetPlayerDailyRank(ctx, today, user3.ID)
	if rank != 3 {
		t.Errorf("expected rank 3, got %d", rank)
	}
}

func TestMemoryRepository_GetPlayerDailyRank_NotOnBoard(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "player1", "SHA256:key1")
	_, _ = repo.GetOrCreateDailySeed(ctx)

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// User hasn't submitted a score
	rank, entry, err := repo.GetPlayerDailyRank(ctx, today, user.ID)
	if err != nil {
		t.Fatalf("GetPlayerDailyRank failed: %v", err)
	}
	if rank != 0 {
		t.Errorf("expected rank 0 for player not on board, got %d", rank)
	}
	if entry != nil {
		t.Error("expected nil entry for player not on board")
	}
}

func TestMemoryRepository_GetPlayerDailyRank_NoSeed(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	user, _ := repo.CreateUser(ctx, "player1", "SHA256:key1")

	// Check rank for a date with no daily seed
	yesterday := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)
	rank, entry, err := repo.GetPlayerDailyRank(ctx, yesterday, user.ID)
	if err != nil {
		t.Fatalf("GetPlayerDailyRank failed: %v", err)
	}
	if rank != 0 {
		t.Errorf("expected rank 0 for non-existent seed, got %d", rank)
	}
	if entry != nil {
		t.Error("expected nil entry for non-existent seed")
	}
}
