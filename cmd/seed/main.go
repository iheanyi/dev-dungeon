// Package main provides a database seeding utility for /dev/dungeon.
// Usage: DATABASE_URL=... go run ./cmd/seed/
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/iheanyi/devdungeon/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()

	client, err := db.NewClientWithOptions(ctx, databaseURL, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("Seeding database...")

	// Create test user if doesn't exist
	testUser, err := client.GetUserByUsername(ctx, "testplayer")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking user: %v\n", err)
		os.Exit(1)
	}

	if testUser == nil {
		testUser, err = client.CreateUser(ctx, "testplayer", "SHA256:test-fingerprint-"+fmt.Sprint(rng.Int63()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create test user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created test user: %s (ID: %s)\n", testUser.Username, testUser.NanoID)
	} else {
		fmt.Printf("Using existing user: %s (ID: %s)\n", testUser.Username, testUser.NanoID)
	}

	classes := []string{"init", "cron", "bash", "vim", "sudo"}

	// Insert standard leaderboard entry
	standardScore := 1500 + rng.Intn(2000)
	standardEntry := &db.LeaderboardEntry{
		UserID:        testUser.ID,
		Username:      testUser.Username,
		RunType:       "standard",
		Score:         standardScore,
		FloorsCleared: 3 + rng.Intn(5),
		TimeSeconds:   300 + rng.Intn(600),
		Class:         classes[rng.Intn(len(classes))],
		Seed:          int64(rng.Int63()),
	}

	if err := client.AddLeaderboardEntry(ctx, standardEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert standard entry: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Inserted standard leaderboard entry: score=%d, class=%s, floors=%d\n",
		standardEntry.Score, standardEntry.Class, standardEntry.FloorsCleared)

	// Insert daily leaderboard entry
	// First, ensure today's daily seed exists
	dailySeed, err := client.GetOrCreateDailySeed(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get/create daily seed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Today's daily seed: %d\n", dailySeed)

	dailyScore := 1000 + rng.Intn(1500)
	dailyEntry := &db.LeaderboardEntry{
		UserID:        testUser.ID,
		Username:      testUser.Username,
		RunType:       "daily",
		Score:         dailyScore,
		FloorsCleared: 2 + rng.Intn(4),
		TimeSeconds:   200 + rng.Intn(400),
		Class:         classes[rng.Intn(len(classes))],
		Seed:          dailySeed,
	}

	if err := client.AddLeaderboardEntry(ctx, dailyEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert daily entry: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Inserted daily leaderboard entry: score=%d, class=%s, floors=%d\n",
		dailyEntry.Score, dailyEntry.Class, dailyEntry.FloorsCleared)

	fmt.Println("\nSeeding complete!")
}
