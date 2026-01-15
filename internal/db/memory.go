package db

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryRepository is an in-memory implementation of Repository for testing.
type MemoryRepository struct {
	mu sync.RWMutex

	// Auto-incrementing IDs
	nextUserID        int
	nextGameSaveID    int
	nextLeaderboardID int
	nextWorldDropID   int

	// Data stores
	users              map[int]*User    // by ID
	usersByFingerprint map[string]*User // by fingerprint
	usersByUsername    map[string]*User // by username (lowercase)
	usersByNanoID      map[string]*User // by nanoid

	gameSaves    map[int]*GameSave     // by user ID
	metaProgress map[int]*MetaProgress // by user ID
	leaderboard  []LeaderboardEntry
	worldDrops   map[int]*WorldDrop // by ID
	dailySeeds   map[string]int64   // by date string

	authTokens  map[string]*authTokenData
	webSessions map[string]*webSessionData
}

type authTokenData struct {
	UserID    int
	ExpiresAt time.Time
	Used      bool
}

type webSessionData struct {
	UserID     int
	ExpiresAt  time.Time
	LastUsedAt time.Time
}

// NewMemoryRepository creates a new in-memory repository for testing.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextUserID:         1,
		nextGameSaveID:     1,
		nextLeaderboardID:  1,
		nextWorldDropID:    1,
		users:              make(map[int]*User),
		usersByFingerprint: make(map[string]*User),
		usersByUsername:    make(map[string]*User),
		usersByNanoID:      make(map[string]*User),
		gameSaves:          make(map[int]*GameSave),
		metaProgress:       make(map[int]*MetaProgress),
		leaderboard:        []LeaderboardEntry{},
		worldDrops:         make(map[int]*WorldDrop),
		dailySeeds:         make(map[string]int64),
		authTokens:         make(map[string]*authTokenData),
		webSessions:        make(map[string]*webSessionData),
	}
}

// Verify MemoryRepository implements Repository at compile time
var _ Repository = (*MemoryRepository)(nil)

// Close is a no-op for in-memory repository.
func (m *MemoryRepository) Close() {}

// --- User Operations ---

func (m *MemoryRepository) GetUserByFingerprint(ctx context.Context, fingerprint string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user := m.usersByFingerprint[fingerprint]
	if user == nil {
		return nil, nil
	}
	// Return a copy
	copy := *user
	return &copy, nil
}

func (m *MemoryRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user := m.usersByUsername[strings.ToLower(username)]
	if user == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (m *MemoryRepository) GetUserByNanoID(ctx context.Context, nanoid string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user := m.usersByNanoID[nanoid]
	if user == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (m *MemoryRepository) CreateUser(ctx context.Context, username, fingerprint string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	user := &User{
		ID:                   m.nextUserID,
		NanoID:               GenerateNanoID(),
		Username:             username,
		PublicKeyFingerprint: fingerprint,
		CreatedAt:            now,
		LastLogin:            now,
		IsBanned:             false,
	}
	m.nextUserID++

	m.users[user.ID] = user
	m.usersByFingerprint[fingerprint] = user
	m.usersByUsername[strings.ToLower(username)] = user
	m.usersByNanoID[user.NanoID] = user

	// Create default meta progress
	m.metaProgress[user.ID] = &MetaProgress{
		UserID:          user.ID,
		UnlockedClasses: []string{"init"},
		UnlockedItems:   []string{},
	}

	copy := *user
	return &copy, nil
}

func (m *MemoryRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user, ok := m.users[userID]; ok {
		user.LastLogin = time.Now()
	}
	return nil
}

func (m *MemoryRepository) UpdateUsername(ctx context.Context, userID int, newUsername string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[userID]
	if !ok {
		return nil
	}

	// Remove old username index
	delete(m.usersByUsername, strings.ToLower(user.Username))

	// Update username
	user.Username = newUsername
	m.usersByUsername[strings.ToLower(newUsername)] = user

	return nil
}

// --- Game Save Operations ---

func (m *MemoryRepository) GetGameSave(ctx context.Context, userID int) (*GameSave, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	save := m.gameSaves[userID]
	if save == nil {
		return nil, nil
	}
	copy := *save
	return &copy, nil
}

func (m *MemoryRepository) UpsertGameSave(ctx context.Context, userID int, saveData interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(saveData)
	if err != nil {
		return err
	}

	now := time.Now()
	existing := m.gameSaves[userID]
	if existing != nil {
		existing.SaveData = data
		existing.UpdatedAt = now
	} else {
		m.gameSaves[userID] = &GameSave{
			ID:        m.nextGameSaveID,
			NanoID:    GenerateNanoID(),
			UserID:    userID,
			SaveData:  data,
			CreatedAt: now,
			UpdatedAt: now,
		}
		m.nextGameSaveID++
	}
	return nil
}

func (m *MemoryRepository) DeleteGameSave(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.gameSaves, userID)
	return nil
}

// --- Meta Progress Operations ---

func (m *MemoryRepository) GetMetaProgress(ctx context.Context, userID int) (*MetaProgress, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	meta := m.metaProgress[userID]
	if meta == nil {
		return nil, nil
	}
	copy := *meta
	return &copy, nil
}

func (m *MemoryRepository) UpdateMetaProgress(ctx context.Context, meta *MetaProgress) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metaProgress[meta.UserID] = meta
	return nil
}

// --- Leaderboard Operations ---

func (m *MemoryRepository) AddLeaderboardEntry(ctx context.Context, entry *LeaderboardEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry.ID = m.nextLeaderboardID
	entry.NanoID = GenerateNanoID()
	entry.CreatedAt = time.Now()
	m.nextLeaderboardID++

	// Copy the entry
	copy := *entry
	m.leaderboard = append(m.leaderboard, copy)
	return nil
}

func (m *MemoryRepository) GetTopScores(ctx context.Context, runType string, limit int) ([]LeaderboardEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter by run type
	var filtered []LeaderboardEntry
	for _, e := range m.leaderboard {
		if e.RunType == runType {
			filtered = append(filtered, e)
		}
	}

	// Sort by score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func (m *MemoryRepository) GetLeaderboard(ctx context.Context, runType string, limit int) ([]LeaderboardEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []LeaderboardEntry
	for _, e := range m.leaderboard {
		if runType == "" || e.RunType == runType {
			// Denormalize username
			if user, ok := m.users[e.UserID]; ok {
				e.Username = user.Username
			}
			filtered = append(filtered, e)
		}
	}

	// Sort by score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// GetDailyLeaderboard retrieves daily leaderboard entries for a specific date with cursor-based pagination.
func (m *MemoryRepository) GetDailyLeaderboard(ctx context.Context, date time.Time, limit int, cursor *LeaderboardCursor) ([]LeaderboardEntry, *LeaderboardCursor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normalizedDate := date.UTC().Truncate(24 * time.Hour)
	key := normalizedDate.Format("2006-01-02")

	// Check if seed exists for this date
	seed, ok := m.dailySeeds[key]
	if !ok {
		return nil, nil, nil // No daily challenge for this date
	}

	// Filter daily entries for this seed
	var filtered []LeaderboardEntry
	for _, e := range m.leaderboard {
		if e.RunType == "daily" && e.Seed == seed {
			entry := e
			// Denormalize username
			if user, ok := m.users[e.UserID]; ok {
				entry.Username = user.Username
			}
			filtered = append(filtered, entry)
		}
	}

	// Sort by score descending, then ID ascending (for stable pagination)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Score != filtered[j].Score {
			return filtered[i].Score > filtered[j].Score
		}
		return filtered[i].ID < filtered[j].ID
	})

	// Assign ranks
	for i := range filtered {
		filtered[i].Rank = i + 1
	}

	// Apply cursor filter
	if cursor != nil {
		startIdx := 0
		for i, e := range filtered {
			if e.Score < cursor.Score || (e.Score == cursor.Score && e.ID > cursor.ID) {
				startIdx = i
				break
			}
			startIdx = len(filtered) // Past all entries
		}
		if startIdx < len(filtered) {
			filtered = filtered[startIdx:]
		} else {
			filtered = nil
		}
	}

	// Limit results
	var nextCursor *LeaderboardCursor
	if len(filtered) > limit {
		filtered = filtered[:limit]
		last := filtered[len(filtered)-1]
		nextCursor = &LeaderboardCursor{
			Score: last.Score,
			ID:    last.ID,
		}
	}

	return filtered, nextCursor, nil
}

// GetPlayerDailyRank retrieves a player's rank and entry for a specific day's leaderboard.
func (m *MemoryRepository) GetPlayerDailyRank(ctx context.Context, date time.Time, userID int) (int, *LeaderboardEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normalizedDate := date.UTC().Truncate(24 * time.Hour)
	key := normalizedDate.Format("2006-01-02")

	// Check if seed exists for this date
	seed, ok := m.dailySeeds[key]
	if !ok {
		return 0, nil, nil // No daily challenge for this date
	}

	// Get all daily entries for this seed and sort them
	var allEntries []LeaderboardEntry
	for _, e := range m.leaderboard {
		if e.RunType == "daily" && e.Seed == seed {
			entry := e
			if user, ok := m.users[e.UserID]; ok {
				entry.Username = user.Username
			}
			allEntries = append(allEntries, entry)
		}
	}

	// Sort by score descending
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Score > allEntries[j].Score
	})

	// Find the player's entry and assign rank
	for i, e := range allEntries {
		if e.UserID == userID {
			e.Rank = i + 1
			return e.Rank, &e, nil
		}
	}

	return 0, nil, nil // Player has no entry for this day
}

// --- World Drop Operations ---

func (m *MemoryRepository) CreateWorldDrop(ctx context.Context, drop *WorldDrop) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	drop.ID = m.nextWorldDropID
	drop.NanoID = GenerateNanoID()
	drop.CreatedAt = time.Now()
	if drop.ExpiresAt.IsZero() {
		drop.ExpiresAt = time.Now().Add(48 * time.Hour)
	}
	m.nextWorldDropID++

	copy := *drop
	m.worldDrops[drop.ID] = &copy
	return nil
}

func (m *MemoryRepository) GetRandomDrop(ctx context.Context, floorType string, excludeUserID int) (*WorldDrop, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	for _, drop := range m.worldDrops {
		if drop.FloorType == floorType &&
			drop.UserID != excludeUserID &&
			drop.ExpiresAt.After(now) {
			// Return first match (not truly random, but good enough for tests)
			copy := *drop
			if user, ok := m.users[drop.UserID]; ok {
				copy.Username = user.Username
			}
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) CleanupExpiredDrops(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var count int64
	for id, drop := range m.worldDrops {
		if drop.ExpiresAt.Before(now) {
			delete(m.worldDrops, id)
			count++
		}
	}
	return count, nil
}

// --- Daily Seed Operations ---

func (m *MemoryRepository) GetOrCreateDailySeed(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	key := today.Format("2006-01-02")

	if seed, ok := m.dailySeeds[key]; ok {
		return seed, nil
	}

	// Generate cryptographically random seed (unpredictable)
	var seedBytes [8]byte
	if _, err := crypto_rand.Read(seedBytes[:]); err != nil {
		return 0, err
	}
	seed := int64(binary.BigEndian.Uint64(seedBytes[:]))
	m.dailySeeds[key] = seed
	return seed, nil
}

// GetDailySeed retrieves the seed for a specific date.
func (m *MemoryRepository) GetDailySeed(ctx context.Context, date time.Time) (*DailySeed, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normalizedDate := date.UTC().Truncate(24 * time.Hour)
	key := normalizedDate.Format("2006-01-02")

	seed, ok := m.dailySeeds[key]
	if !ok {
		return nil, nil
	}

	return &DailySeed{
		Date:      normalizedDate,
		Seed:      seed,
		CreatedAt: normalizedDate, // Approximate for in-memory
	}, nil
}

// --- Auth Token Operations ---

func (m *MemoryRepository) CreateAuthToken(ctx context.Context, userID int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, err := GenerateAuthToken()
	if err != nil {
		return "", err
	}

	m.authTokens[token] = &authTokenData{
		UserID:    userID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	return token, nil
}

func (m *MemoryRepository) VerifyAuthToken(ctx context.Context, token string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.authTokens[token]
	if !ok {
		return nil, nil
	}

	if data.Used || time.Now().After(data.ExpiresAt) {
		return nil, nil
	}

	data.Used = true

	user := m.users[data.UserID]
	if user == nil {
		return nil, nil
	}

	copy := *user
	return &copy, nil
}

func (m *MemoryRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var count int64
	for token, data := range m.authTokens {
		if data.ExpiresAt.Before(now) {
			delete(m.authTokens, token)
			count++
		}
	}
	return count, nil
}

// --- Web Session Operations ---

func (m *MemoryRepository) CreateWebSession(ctx context.Context, userID int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	m.webSessions[token] = &webSessionData{
		UserID:     userID,
		ExpiresAt:  now.Add(7 * 24 * time.Hour),
		LastUsedAt: now,
	}

	return token, nil
}

func (m *MemoryRepository) GetWebSession(ctx context.Context, token string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.webSessions[token]
	if !ok {
		return nil, nil
	}

	if time.Now().After(data.ExpiresAt) {
		delete(m.webSessions, token)
		return nil, nil
	}

	data.LastUsedAt = time.Now()

	user := m.users[data.UserID]
	if user == nil {
		return nil, nil
	}

	copy := *user
	return &copy, nil
}

func (m *MemoryRepository) DeleteWebSession(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.webSessions, token)
	return nil
}

func (m *MemoryRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var count int64
	for token, data := range m.webSessions {
		if data.ExpiresAt.Before(now) {
			delete(m.webSessions, token)
			count++
		}
	}
	return count, nil
}
