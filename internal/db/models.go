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
	ID                   int       `db:"id"`
	NanoID               string    `db:"nanoid"`
	Username             string    `db:"username"`
	PublicKeyFingerprint string    `db:"public_key_fingerprint"`
	CreatedAt            time.Time `db:"created_at"`
	LastLogin            time.Time `db:"last_login"`
	IsBanned             bool      `db:"is_banned"`
}

// GameSave represents a saved game state.
type GameSave struct {
	ID        int       `db:"id"`
	NanoID    string    `db:"nanoid"`
	UserID    int       `db:"user_id"`
	SaveData  []byte    `db:"save_data"` // JSON blob
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
	ID            int       `db:"id"`
	NanoID        string    `db:"nanoid"`
	UserID        int       `db:"user_id"`
	Username      string    `db:"username"` // Denormalized for display
	RunType       string    `db:"run_type"` // 'standard', 'daily', 'seeded'
	Seed          int64     `db:"seed"`
	Score         int       `db:"score"`
	FloorsCleared int       `db:"floors_cleared"`
	TimeSeconds   int       `db:"time_seconds"`
	Class         string    `db:"class"`
	CreatedAt     time.Time `db:"created_at"`
}

// WorldDrop represents an async drop (message or item) left by a player.
type WorldDrop struct {
	ID        int       `db:"id"`
	NanoID    string    `db:"nanoid"`
	UserID    int       `db:"user_id"`
	Username  string    `db:"username"` // Denormalized
	FloorType string    `db:"floor_type"`
	PositionX int       `db:"position_x"`
	PositionY int       `db:"position_y"`
	DropType  string    `db:"drop_type"` // 'message', 'item', 'gravestone'
	Content   string    `db:"content"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// DailySeed represents the seed for a daily run.
type DailySeed struct {
	Date      time.Time `db:"date"`
	Seed      int64     `db:"seed"`
	CreatedAt time.Time `db:"created_at"`
}

// AuthToken represents a magic link token for browser authentication.
type AuthToken struct {
	Token     string    `db:"token"`     // 256-bit hex string (64 chars)
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
	Token      string    `db:"token"`      // 256-bit hex string (64 chars)
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
