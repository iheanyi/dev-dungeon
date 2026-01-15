// Package db provides database models and access for /dev/dungeon multiplayer.
package db

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// NanoID alphabet: alphanumeric lowercase only
const nanoIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// GenerateNanoID creates a new 21-character alphanumeric ID.
func GenerateNanoID() string {
	id, err := gonanoid.Generate(nanoIDAlphabet, 21)
	if err != nil {
		// Fallback to default if custom alphabet fails
		id, _ = gonanoid.New()
	}
	return id
}

// User represents a registered player account.
type User struct {
	ID                   int       `db:"id" json:"-"`
	NanoID               string    `db:"nanoid" json:"public_id"`
	Username             string    `db:"username" json:"username"`
	PublicKeyFingerprint string    `db:"public_key_fingerprint" json:"-"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	LastLogin            time.Time `db:"last_login" json:"last_login"`
	IsBanned             bool      `db:"is_banned" json:"-"`
}

// GameSave represents a saved game state.
type GameSave struct {
	ID        int       `db:"id" json:"-"`
	NanoID    string    `db:"nanoid" json:"-"`
	UserID    int       `db:"user_id" json:"-"`
	SaveData  []byte    `db:"save_data" json:"-"` // JSON blob - internal
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// MetaProgress represents permanent unlocks that persist across runs.
type MetaProgress struct {
	UserID           int      `db:"user_id"`
	TotalExitCodes   int      `db:"total_exit_codes"`
	UnlockedClasses  []string `db:"unlocked_classes"`
	PermanentBonuses []byte   `db:"permanent_bonuses"` // JSON blob
	UnlockedItems    []string `db:"unlocked_items"`
	RunsCompleted    int      `db:"runs_completed"`
	DeepestFloor     int      `db:"deepest_floor"`
	TotalDeaths      int      `db:"total_deaths"`
}

// LeaderboardEntry represents a score entry.
type LeaderboardEntry struct {
	ID            int       `db:"id" json:"-"`
	NanoID        string    `db:"nanoid" json:"-"`
	UserID        int       `db:"user_id" json:"-"`
	Username      string    `db:"username" json:"username"`
	RunType       string    `db:"run_type" json:"run_type"`
	Seed          int64     `db:"seed" json:"seed,omitempty"`
	Score         int       `db:"score" json:"score"`
	FloorsCleared int       `db:"floors_cleared" json:"floors_cleared"`
	TimeSeconds   int       `db:"time_seconds" json:"time_seconds"`
	Class         string    `db:"class" json:"class"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	Rank          int       `db:"-" json:"rank,omitempty"`
}

// LeaderboardCursor enables stable cursor-based pagination for leaderboards.
// Uses (Score, ID) tuple to handle ties and concurrent insertions.
type LeaderboardCursor struct {
	Score int // Last seen score
	ID    int // Tiebreaker for equal scores
}

// WorldDrop represents an async drop (message or item) left by a player.
type WorldDrop struct {
	ID        int       `db:"id" json:"-"`
	NanoID    string    `db:"nanoid" json:"-"`
	UserID    int       `db:"user_id" json:"-"`
	Username  string    `db:"username" json:"username"`
	FloorType string    `db:"floor_type" json:"floor_type"`
	PositionX int       `db:"position_x" json:"position_x"`
	PositionY int       `db:"position_y" json:"position_y"`
	DropType  string    `db:"drop_type" json:"drop_type"`
	Content   string    `db:"content" json:"content"`
	ExpiresAt time.Time `db:"expires_at" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// DailySeed represents the seed for a daily run.
type DailySeed struct {
	Date      time.Time `db:"date" json:"date"`
	Seed      int64     `db:"seed" json:"seed"`
	CreatedAt time.Time `db:"created_at" json:"-"`
}

// AuthToken represents a magic link token for browser authentication.
type AuthToken struct {
	Token     string    `db:"token"` // 256-bit hex string (64 chars)
	UserID    int       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
}

// GenerateAuthToken creates a cryptographically secure 256-bit token.
// Returns a 64-character hex string.
func GenerateAuthToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// WebSession represents an authenticated browser session.
type WebSession struct {
	Token      string    `db:"token"` // 256-bit hex string (64 chars)
	UserID     int       `db:"user_id"`
	CreatedAt  time.Time `db:"created_at"`
	ExpiresAt  time.Time `db:"expires_at"`
	LastUsedAt time.Time `db:"last_used_at"`
}

// GenerateSessionToken creates a cryptographically secure session token.
// Returns a 64-character hex string.
func GenerateSessionToken() (string, error) {
	return GenerateAuthToken() // Same format
}
