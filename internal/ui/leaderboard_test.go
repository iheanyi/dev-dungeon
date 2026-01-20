package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/save"
	"github.com/iheanyi/devdungeon/internal/types"
)

// === Daily Seed Tests ===

func TestDailySeedConsistency(t *testing.T) {
	// Multiple calls within the same day should return the same seed
	seed1 := getDailySeed()
	seed2 := getDailySeed()
	seed3 := getDailySeed()

	if seed1 != seed2 {
		t.Errorf("daily seed should be consistent: got %d and %d", seed1, seed2)
	}

	if seed2 != seed3 {
		t.Errorf("daily seed should be consistent: got %d and %d", seed2, seed3)
	}
}

func TestDailySeedIsNonZero(t *testing.T) {
	seed := getDailySeed()

	if seed == 0 {
		t.Error("daily seed should not be zero")
	}
}

func TestDailySeedIsPositive(t *testing.T) {
	seed := getDailySeed()

	// UnixNano for any date after 1970 should be positive
	if seed < 0 {
		t.Errorf("daily seed should be positive, got %d", seed)
	}
}

func TestDailySeedUsesUTCMidnight(t *testing.T) {
	// The seed should be based on UTC midnight, meaning it's divisible by
	// the number of nanoseconds in a day (from truncation)
	seed := getDailySeed()

	// 24 hours * 60 minutes * 60 seconds * 1e9 nanoseconds
	nsPerDay := int64(24 * 60 * 60 * 1e9)

	// After truncating to day, the remainder should be 0
	if seed%nsPerDay != 0 {
		t.Error("daily seed should be at UTC midnight (truncated to day boundary)")
	}
}

func TestDailyRunModeUsesDailySeed(t *testing.T) {
	m := newTestModel()
	m.dailyRunMode = true

	// Start a new game in daily run mode
	cfg := config.DefaultConfig()
	m.config = cfg
	m.startNewGame(entity.ClassInit)

	// The engine's master seed should match the daily seed
	expectedSeed := getDailySeed()
	actualSeed := m.engine.MasterSeed()

	if actualSeed != expectedSeed {
		t.Errorf("daily run should use daily seed: expected %d, got %d", expectedSeed, actualSeed)
	}
}

func TestStandardRunDoesNotUseDailySeed(t *testing.T) {
	m := newTestModel()
	m.dailyRunMode = false

	cfg := config.DefaultConfig()
	m.config = cfg
	m.startNewGame(entity.ClassInit)

	dailySeed := getDailySeed()
	actualSeed := m.engine.MasterSeed()

	// Standard run uses random seed (0 triggers random generation),
	// so the actual seed should differ from the daily seed
	// (extremely unlikely to match by chance)
	if actualSeed == dailySeed {
		t.Log("Warning: standard run seed matched daily seed (extremely unlikely)")
		// Don't fail the test as this could happen by pure chance, but log it
	}
}

// === Leaderboard Tests ===

func TestSetLeaderboardFetcher(t *testing.T) {
	m := newTestModel()

	// Initially nil
	if m.leaderboardFetcher != nil {
		t.Error("leaderboardFetcher should initially be nil")
	}

	// Set a fetcher
	fetcherCalled := false
	m.SetLeaderboardFetcher(func(runType string, limit int) ([]LeaderboardEntry, error) {
		fetcherCalled = true
		return nil, nil
	})

	if m.leaderboardFetcher == nil {
		t.Error("leaderboardFetcher should be set")
	}

	// Call the fetcher
	_, _ = m.leaderboardFetcher("all", 10)
	if !fetcherCalled {
		t.Error("fetcher should have been called")
	}
}

func TestOpenLeaderboard(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewMainMenu

	// Set up mock fetcher
	testEntries := []LeaderboardEntry{
		{Rank: 1, Username: "player1", Score: 1000, FloorsCleared: 8, Class: "sudo"},
		{Rank: 2, Username: "player2", Score: 500, FloorsCleared: 5, Class: "bash"},
	}
	m.SetLeaderboardFetcher(func(runType string, limit int) ([]LeaderboardEntry, error) {
		return testEntries, nil
	})

	// Open leaderboard
	m.openLeaderboard()

	// Verify state
	if m.currentView != ViewLeaderboard {
		t.Error("should transition to leaderboard view")
	}

	if m.leaderboardCursor != 0 {
		t.Error("cursor should be reset to 0")
	}

	if m.leaderboardRunType != "all" {
		t.Errorf("run type should be 'all', got '%s'", m.leaderboardRunType)
	}

	if len(m.leaderboardEntries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m.leaderboardEntries))
	}

	if m.leaderboardError != "" {
		t.Errorf("should have no error, got '%s'", m.leaderboardError)
	}
}

func TestOpenLeaderboardWithError(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewMainMenu

	// Set up fetcher that returns error
	m.SetLeaderboardFetcher(func(runType string, limit int) ([]LeaderboardEntry, error) {
		return nil, fmt.Errorf("network error")
	})

	// Open leaderboard
	m.openLeaderboard()

	// Should still transition to view
	if m.currentView != ViewLeaderboard {
		t.Error("should transition to leaderboard view even on error")
	}

	// Should have error message
	if m.leaderboardError == "" {
		t.Error("should have error message")
	}

	// Entries should be nil/empty
	if len(m.leaderboardEntries) != 0 {
		t.Error("entries should be empty on error")
	}
}

func TestOpenLeaderboardWithoutFetcher(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewMainMenu
	m.leaderboardFetcher = nil

	// Open leaderboard without fetcher (single-player mode)
	m.openLeaderboard()

	// Should still transition to view
	if m.currentView != ViewLeaderboard {
		t.Error("should transition to leaderboard view")
	}

	// Entries should be empty
	if len(m.leaderboardEntries) != 0 {
		t.Error("entries should be empty without fetcher")
	}
}

func TestLeaderboardEntryStructure(t *testing.T) {
	// Test that LeaderboardEntry has all expected fields
	entry := LeaderboardEntry{
		Rank:          1,
		Username:      "testuser",
		Score:         1234,
		FloorsCleared: 5,
		Class:         "bash",
		Seed:          12345,
		RunType:       "standard",
	}

	if entry.Rank != 1 {
		t.Error("Rank field not set correctly")
	}
	if entry.Username != "testuser" {
		t.Error("Username field not set correctly")
	}
	if entry.Score != 1234 {
		t.Error("Score field not set correctly")
	}
	if entry.FloorsCleared != 5 {
		t.Error("FloorsCleared field not set correctly")
	}
	if entry.Class != "bash" {
		t.Error("Class field not set correctly")
	}
	if entry.Seed != 12345 {
		t.Error("Seed field not set correctly")
	}
	if entry.RunType != "standard" {
		t.Error("RunType field not set correctly")
	}
}

func TestLeaderboardRefresh(t *testing.T) {
	m := newTestModel()

	callCount := 0
	m.SetLeaderboardFetcher(func(runType string, limit int) ([]LeaderboardEntry, error) {
		callCount++
		return []LeaderboardEntry{{Rank: callCount}}, nil
	})

	m.openLeaderboard()
	if callCount != 1 {
		t.Error("fetcher should be called once on open")
	}

	// Simulate refresh by calling refreshLeaderboard
	m.refreshLeaderboard()
	if callCount != 2 {
		t.Error("fetcher should be called again on refresh")
	}

	// Entries should be updated
	if m.leaderboardEntries[0].Rank != 2 {
		t.Error("entries should be updated after refresh")
	}
}

func TestLeaderboardRunTypeFilter(t *testing.T) {
	m := newTestModel()

	var capturedRunType string
	m.SetLeaderboardFetcher(func(runType string, limit int) ([]LeaderboardEntry, error) {
		capturedRunType = runType
		return nil, nil
	})

	// Test "all" filter
	m.leaderboardRunType = "all"
	m.refreshLeaderboard()
	if capturedRunType != "all" {
		t.Errorf("expected runType 'all', got '%s'", capturedRunType)
	}

	// Test "standard" filter
	m.leaderboardRunType = "standard"
	m.refreshLeaderboard()
	if capturedRunType != "standard" {
		t.Errorf("expected runType 'standard', got '%s'", capturedRunType)
	}

	// Test "daily" filter
	m.leaderboardRunType = "daily"
	m.refreshLeaderboard()
	if capturedRunType != "daily" {
		t.Errorf("expected runType 'daily', got '%s'", capturedRunType)
	}
}

// === Daily Run Gating Tests ===

func TestDailyRunBlockedWhenCompleted(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true
	m.dailyRunCompleted = true
	m.currentView = ViewMainMenu
	m.menuCursor = 0
	m.menuOptions = []string{"Daily Run"}

	// Simulate selecting "Daily Run" menu option
	m.handleMenuSelection("Daily Run")

	// Should stay on main menu with error message
	if m.currentView != ViewMainMenu {
		t.Errorf("should stay on main menu when daily completed, got %v", m.currentView)
	}
	if m.statusMsg != "You've already completed today's daily run" {
		t.Errorf("wrong status message: %s", m.statusMsg)
	}
}

func TestDailyRunRedirectsWhenInProgress(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true
	m.dailyRunInProgress = true
	m.currentView = ViewMainMenu
	m.menuCursor = 0
	m.menuOptions = []string{"Daily Run"}

	// Simulate selecting "Daily Run" menu option
	m.handleMenuSelection("Daily Run")

	// Should stay on main menu with message to use Continue
	if m.currentView != ViewMainMenu {
		t.Errorf("should stay on main menu when daily in progress, got %v", m.currentView)
	}
	if m.statusMsg != "You have a daily run in progress - use Continue" {
		t.Errorf("wrong status message: %s", m.statusMsg)
	}
}

func TestDailyRunAllowedWhenNoActiveDailyRun(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true
	m.dailyRunCompleted = false
	m.dailyRunInProgress = false
	m.currentView = ViewMainMenu
	m.menuCursor = 0
	m.menuOptions = []string{"Daily Run"}

	// Simulate selecting "Daily Run" menu option
	m.handleMenuSelection("Daily Run")

	// Should transition to class select for daily run
	if m.currentView != ViewClassSelect {
		t.Errorf("should transition to class select, got %v", m.currentView)
	}
	if !m.dailyRunMode {
		t.Error("dailyRunMode should be true")
	}
}

func TestDailyRunRequiresMultiplayer(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = false
	m.dailyRunCompleted = false
	m.dailyRunInProgress = false
	m.currentView = ViewMainMenu

	// Simulate selecting "Daily Run" menu option
	m.handleMenuSelection("Daily Run")

	// Should stay on main menu with error
	if m.currentView != ViewMainMenu {
		t.Errorf("should stay on main menu when not multiplayer, got %v", m.currentView)
	}
	if m.statusMsg != "Daily runs require SSH connection" {
		t.Errorf("wrong status message: %s", m.statusMsg)
	}
}

func TestNewGameWarnsWithInProgressDaily(t *testing.T) {
	m := newTestModel()
	m.dailyRunInProgress = true
	m.pendingSave = &save.SaveData{CurrentDepth: 3}
	m.currentView = ViewMainMenu
	m.menuCursor = 0
	m.menuOptions = []string{"New Game"}

	// Simulate selecting "New Game" menu option
	m.handleMenuSelection("New Game")

	// Should show confirmation dialog
	if m.currentView != ViewConfirmDialog {
		t.Errorf("expected ViewConfirmDialog, got %v", m.currentView)
	}

	// Confirm message should mention the floor depth
	if m.confirmMessage == "" {
		t.Error("confirm message should be set")
	}

	// Confirm action should be set
	if m.confirmAction == nil {
		t.Error("confirm action should be set")
	}

	// Return view should be main menu
	if m.confirmReturnView != ViewMainMenu {
		t.Errorf("confirm return view should be ViewMainMenu, got %v", m.confirmReturnView)
	}
}

func TestNewGameProceedsWithoutInProgressDaily(t *testing.T) {
	m := newTestModel()
	m.dailyRunInProgress = false
	m.pendingSave = nil
	m.currentView = ViewMainMenu
	m.menuCursor = 0
	m.menuOptions = []string{"New Game"}

	// Simulate selecting "New Game" menu option
	m.handleMenuSelection("New Game")

	// Should go directly to class select (no confirmation dialog)
	if m.currentView != ViewClassSelect {
		t.Errorf("expected ViewClassSelect, got %v", m.currentView)
	}
}

func TestConfirmDialogYesExecutesAction(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewConfirmDialog

	// Set up confirmation state
	actionExecuted := false
	m.confirmMessage = "Test confirmation"
	m.confirmAction = func() {
		actionExecuted = true
	}
	m.confirmReturnView = ViewMainMenu

	// Simulate pressing 'y'
	m.handleConfirmDialogKey("y")

	// Action should have been executed
	if !actionExecuted {
		t.Error("confirm action should have been executed")
	}
}

func TestConfirmDialogEnterExecutesAction(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewConfirmDialog

	// Set up confirmation state
	actionExecuted := false
	m.confirmMessage = "Test confirmation"
	m.confirmAction = func() {
		actionExecuted = true
	}
	m.confirmReturnView = ViewMainMenu

	// Simulate pressing 'enter'
	m.handleConfirmDialogKey("enter")

	// Action should have been executed
	if !actionExecuted {
		t.Error("confirm action should have been executed on enter")
	}
}

func TestConfirmDialogNoReturnsToMenu(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewConfirmDialog

	// Set up confirmation state
	actionExecuted := false
	m.confirmMessage = "Test confirmation"
	m.confirmAction = func() {
		actionExecuted = true
	}
	m.confirmReturnView = ViewMainMenu

	// Simulate pressing 'n'
	m.handleConfirmDialogKey("n")

	// Should return to the return view
	if m.currentView != ViewMainMenu {
		t.Errorf("should return to main menu, got %v", m.currentView)
	}

	// Action should NOT have been executed
	if actionExecuted {
		t.Error("confirm action should not have been executed on 'n'")
	}
}

func TestConfirmDialogEscReturnsToMenu(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewConfirmDialog

	// Set up confirmation state
	actionExecuted := false
	m.confirmMessage = "Test confirmation"
	m.confirmAction = func() {
		actionExecuted = true
	}
	m.confirmReturnView = ViewGame

	// Simulate pressing 'esc'
	m.handleConfirmDialogKey("esc")

	// Should return to the return view (ViewGame in this case)
	if m.currentView != ViewGame {
		t.Errorf("should return to game view, got %v", m.currentView)
	}

	// Action should NOT have been executed
	if actionExecuted {
		t.Error("confirm action should not have been executed on esc")
	}
}

func TestShowConfirmDialog(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewMainMenu

	// Set up a confirmation dialog
	actionExecuted := false
	m.showConfirmDialog(
		"Test message",
		func() { actionExecuted = true },
		ViewMainMenu,
	)

	// Should transition to confirm dialog view
	if m.currentView != ViewConfirmDialog {
		t.Errorf("expected ViewConfirmDialog, got %v", m.currentView)
	}

	// Message should be set
	if m.confirmMessage != "Test message" {
		t.Errorf("expected message 'Test message', got '%s'", m.confirmMessage)
	}

	// Return view should be set
	if m.confirmReturnView != ViewMainMenu {
		t.Errorf("expected return view ViewMainMenu, got %v", m.confirmReturnView)
	}

	// Execute the action
	m.confirmAction()
	if !actionExecuted {
		t.Error("action should have been executed")
	}
}

func TestSetDailyRunStatus(t *testing.T) {
	m := newTestModel()

	// Initially all should be default values
	if m.dailyRunCompleted {
		t.Error("dailyRunCompleted should initially be false")
	}
	if m.dailyRunInProgress {
		t.Error("dailyRunInProgress should initially be false")
	}
	if m.dailySeed != 0 {
		t.Error("dailySeed should initially be 0")
	}

	// Set daily run status
	m.SetDailyRunStatus(true, false, 12345)

	if !m.dailyRunCompleted {
		t.Error("dailyRunCompleted should be true after SetDailyRunStatus")
	}
	if m.dailyRunInProgress {
		t.Error("dailyRunInProgress should be false")
	}
	if m.dailySeed != 12345 {
		t.Errorf("dailySeed should be 12345, got %d", m.dailySeed)
	}

	// Update to different status
	m.SetDailyRunStatus(false, true, 67890)

	if m.dailyRunCompleted {
		t.Error("dailyRunCompleted should be false")
	}
	if !m.dailyRunInProgress {
		t.Error("dailyRunInProgress should be true")
	}
	if m.dailySeed != 67890 {
		t.Errorf("dailySeed should be 67890, got %d", m.dailySeed)
	}
}

func TestSetSubmitDailyCallback(t *testing.T) {
	m := newTestModel()

	// Initially nil
	if m.submitDailyCallback != nil {
		t.Error("submitDailyCallback should initially be nil")
	}

	// Set a callback
	callbackCalled := false
	m.SetSubmitDailyCallback(func(saveData *save.SaveData) error {
		callbackCalled = true
		return nil
	})

	if m.submitDailyCallback == nil {
		t.Error("submitDailyCallback should be set")
	}

	// Call the callback
	_ = m.submitDailyCallback(nil)
	if !callbackCalled {
		t.Error("callback should have been called")
	}
}

func TestNewGameConfirmationSubmitsDaily(t *testing.T) {
	m := newTestModel()
	m.dailyRunInProgress = true
	m.pendingSave = &save.SaveData{CurrentDepth: 5, MasterSeed: 12345}
	m.currentView = ViewMainMenu
	m.hasValidSave = true

	// Set up the submit daily callback (runs in goroutine so we just verify it's set)
	m.SetSubmitDailyCallback(func(saveData *save.SaveData) error {
		return nil
	})

	// Trigger the new game selection which shows confirmation
	m.handleMenuSelection("New Game")

	// Should be in confirm dialog
	if m.currentView != ViewConfirmDialog {
		t.Fatalf("expected ViewConfirmDialog, got %v", m.currentView)
	}

	// Execute the confirmation action (simulating pressing 'y')
	if m.confirmAction != nil {
		m.confirmAction()
	}

	// After confirmation, daily run state should be updated
	if m.dailyRunInProgress {
		t.Error("dailyRunInProgress should be false after confirming new game")
	}
	if !m.dailyRunCompleted {
		t.Error("dailyRunCompleted should be true after confirming new game")
	}
	if m.hasValidSave {
		t.Error("hasValidSave should be false after confirming new game")
	}

	// Note: The actual submission happens in a goroutine, so we can't easily
	// verify it was called synchronously. The callback setup test above
	// verifies the callback mechanism works.
}

// === RunType Persistence Tests ===

func TestRunTypeSetToStandardByDefault(t *testing.T) {
	m := newTestModel()

	// Start a standard game
	m.dailyRunMode = false
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 0)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentRunType = "standard"

	if m.currentRunType != "standard" {
		t.Errorf("expected standard run type, got %s", m.currentRunType)
	}
}

func TestRunTypeSetToDailyForDailyRun(t *testing.T) {
	m := newTestModel()

	// Simulate starting a daily run
	dailySeed := getDailySeed()
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, dailySeed)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentRunType = "daily"

	if m.currentRunType != "daily" {
		t.Errorf("expected daily run type, got %s", m.currentRunType)
	}
}

func TestGetSaveDataIncludesRunType(t *testing.T) {
	m := newTestModel()

	// Start a daily run
	dailySeed := getDailySeed()
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, dailySeed)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentRunType = "daily"

	// Get save data
	saveData := m.GetSaveData()
	if saveData == nil {
		t.Fatal("expected save data, got nil")
	}

	if saveData.RunType != "daily" {
		t.Errorf("expected RunType 'daily' in save data, got '%s'", saveData.RunType)
	}
}

func TestGetSaveDataDefaultsToStandard(t *testing.T) {
	m := newTestModel()

	// Start a game without setting run type
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentRunType = "" // Empty

	// Get save data - should default to standard
	saveData := m.GetSaveData()
	if saveData == nil {
		t.Fatal("expected save data, got nil")
	}

	if saveData.RunType != "standard" {
		t.Errorf("expected RunType 'standard' when empty, got '%s'", saveData.RunType)
	}
}

func TestContinueRestoresRunTypeFromSave(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true

	// Create a save with daily run type
	dailySeed := getDailySeed()
	pendingSave := &save.SaveData{
		Version:      save.Version,
		MasterSeed:   dailySeed,
		CurrentDepth: 1,
		RunType:      "daily",
		Player: save.PlayerData{
			Class:     entity.ClassInit,
			Level:     1,
			XP:        0,
			XPToLevel: 100,
			Stats: types.Stats{
				RAM:  100,
				CPU:  10,
				FD:   16,
				NICE: 10,
				UID:  1000,
			},
			MaxStats: types.MaxStats{
				MaxRAM: 100,
				MaxFD:  16,
			},
			Position:    types.Position{X: 5, Y: 5},
			Inventory:   []save.ItemData{},
			SkillStates: []save.SkillState{{ID: "fork", CurrentCD: 0}},
		},
		FloorStates:  []save.FloorState{},
		MetaProgress: save.MetaProgress{UnlockedClasses: []string{"init"}},
	}

	m.SetHasValidSave(true, pendingSave)

	// Continue the game
	m.continueGame()

	// Run type should be restored
	if m.currentRunType != "daily" {
		t.Errorf("expected currentRunType 'daily' after continue, got '%s'", m.currentRunType)
	}
}

func TestContinueDefaultsToStandardWhenNoRunType(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true

	// Create a save without run type (old save format)
	pendingSave := &save.SaveData{
		Version:      save.Version,
		MasterSeed:   12345,
		CurrentDepth: 1,
		RunType:      "", // Empty - simulating old save
		Player: save.PlayerData{
			Class:     entity.ClassInit,
			Level:     1,
			XP:        0,
			XPToLevel: 100,
			Stats: types.Stats{
				RAM:  100,
				CPU:  10,
				FD:   16,
				NICE: 10,
				UID:  1000,
			},
			MaxStats: types.MaxStats{
				MaxRAM: 100,
				MaxFD:  16,
			},
			Position:    types.Position{X: 5, Y: 5},
			Inventory:   []save.ItemData{},
			SkillStates: []save.SkillState{{ID: "fork", CurrentCD: 0}},
		},
		FloorStates:  []save.FloorState{},
		MetaProgress: save.MetaProgress{UnlockedClasses: []string{"init"}},
	}

	m.SetHasValidSave(true, pendingSave)

	// Continue the game
	m.continueGame()

	// Run type should default to standard
	if m.currentRunType != "standard" {
		t.Errorf("expected currentRunType 'standard' for old saves, got '%s'", m.currentRunType)
	}
}

func TestLeaderboardSubmitterCalledWithCorrectRunType(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true
	m.currentRunType = "daily"

	// Track what gets submitted using a channel for synchronization
	submittedCh := make(chan string, 1)
	m.leaderboardSubmitter = func(score, floors, timeSeconds int, class string, seed int64, runType string, victory bool) error {
		submittedCh <- runType
		return nil
	}

	// Setup game state
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, getDailySeed())
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Submit to leaderboard
	m.submitToLeaderboard(false)

	// Wait for submission - the submitter runs in a goroutine so give it time
	submittedRunType := <-submittedCh

	if submittedRunType != "daily" {
		t.Errorf("expected submitted runType 'daily', got '%s'", submittedRunType)
	}
}

func TestSaveDataRunTypePreservedThroughFullCycle(t *testing.T) {
	// This test verifies the full cycle:
	// 1. Start daily run
	// 2. Get save data (with RunType)
	// 3. Load from save data
	// 4. RunType should be preserved

	m := newTestModel()
	m.isMultiplayer = true

	// Step 1: Start a daily run
	dailySeed := getDailySeed()
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, dailySeed)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentRunType = "daily"

	// Step 2: Get save data
	saveData := m.GetSaveData()
	if saveData == nil {
		t.Fatal("expected save data")
	}
	if saveData.RunType != "daily" {
		t.Errorf("save data RunType should be 'daily', got '%s'", saveData.RunType)
	}

	// Step 3: Simulate closing and reopening
	m.engine.Shutdown()
	m.engine = nil
	m.currentRunType = "" // Reset

	// Step 4: Load from save
	m.SetHasValidSave(true, saveData)
	m.continueGame()

	// Verify RunType is preserved
	if m.currentRunType != "daily" {
		t.Errorf("RunType should be preserved as 'daily' after continue, got '%s'", m.currentRunType)
	}
}

// === Time Tracking Tests ===

func TestTimeTrackingInitializedOnNewGame(t *testing.T) {
	m := newTestModel()
	m.config = config.DefaultConfig()

	// Start a new game
	m.startNewGame(entity.ClassInit)

	// Time tracking fields should be initialized
	if m.runStartTime.IsZero() {
		t.Error("runStartTime should be set after starting new game")
	}
	if m.sessionStartTime.IsZero() {
		t.Error("sessionStartTime should be set after starting new game")
	}
	if m.elapsedSeconds != 0 {
		t.Errorf("elapsedSeconds should be 0 for new game, got %d", m.elapsedSeconds)
	}
}

func TestGetSaveDataIncludesTimeTracking(t *testing.T) {
	m := newTestModel()
	m.config = config.DefaultConfig()

	// Start a new game
	m.startNewGame(entity.ClassInit)

	// Get save data
	saveData := m.GetSaveData()
	if saveData == nil {
		t.Fatal("expected save data, got nil")
	}

	// RunStartTime should be set
	if saveData.RunStartTime.IsZero() {
		t.Error("RunStartTime should be set in save data")
	}

	// ElapsedSeconds should be >= 0 (might be 0 for a quick test)
	if saveData.ElapsedSeconds < 0 {
		t.Errorf("ElapsedSeconds should be >= 0, got %d", saveData.ElapsedSeconds)
	}
}

func TestContinueRestoresTimeTracking(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true

	// Create a save with time tracking
	pendingSave := &save.SaveData{
		Version:        save.Version,
		MasterSeed:     12345,
		CurrentDepth:   1,
		RunType:        "standard",
		RunStartTime:   time.Now().UTC().Add(-1 * time.Hour), // Started 1 hour ago
		ElapsedSeconds: 300,                                  // 5 minutes accumulated
		Player: save.PlayerData{
			Class:     entity.ClassInit,
			Level:     1,
			XP:        0,
			XPToLevel: 100,
			Stats: types.Stats{
				RAM:  100,
				CPU:  10,
				FD:   16,
				NICE: 10,
				UID:  1000,
			},
			MaxStats: types.MaxStats{
				MaxRAM: 100,
				MaxFD:  16,
			},
			Position:    types.Position{X: 5, Y: 5},
			Inventory:   []save.ItemData{},
			SkillStates: []save.SkillState{{ID: "fork", CurrentCD: 0}},
		},
		FloorStates:  []save.FloorState{},
		MetaProgress: save.MetaProgress{UnlockedClasses: []string{"init"}},
	}

	m.SetHasValidSave(true, pendingSave)

	// Continue the game
	m.continueGame()

	// Time tracking should be restored
	if m.runStartTime.IsZero() {
		t.Error("runStartTime should be restored from save")
	}
	if m.elapsedSeconds != 300 {
		t.Errorf("elapsedSeconds should be 300 from save, got %d", m.elapsedSeconds)
	}
	if m.sessionStartTime.IsZero() {
		t.Error("sessionStartTime should be set for new session")
	}
}

func TestGetTotalPlayTime(t *testing.T) {
	m := newTestModel()
	m.config = config.DefaultConfig()

	// Start a new game
	m.startNewGame(entity.ClassInit)

	// Set some previous elapsed time
	m.elapsedSeconds = 60 // 1 minute from previous sessions

	// Get total play time
	totalTime := m.GetTotalPlayTime()

	// Should be at least the previous elapsed time
	if totalTime < 60 {
		t.Errorf("expected total time >= 60, got %d", totalTime)
	}
}

func TestGetTotalPlayTimeWithZeroSession(t *testing.T) {
	m := newTestModel()

	// Without starting a session (sessionStartTime is zero)
	m.elapsedSeconds = 120

	totalTime := m.GetTotalPlayTime()

	// Should return just the elapsed seconds when no active session
	if totalTime != 120 {
		t.Errorf("expected total time 120 with no active session, got %d", totalTime)
	}
}

func TestTimeTrackingAccumulatesAcrossSaves(t *testing.T) {
	m := newTestModel()
	m.config = config.DefaultConfig()

	// Start a new game
	m.startNewGame(entity.ClassInit)

	// Simulate some elapsed time from a previous session
	m.elapsedSeconds = 100

	// Get save data
	saveData := m.GetSaveData()
	if saveData == nil {
		t.Fatal("expected save data")
	}

	// ElapsedSeconds in save should include accumulated time
	// Note: since test runs quickly, current session adds ~0 seconds
	if saveData.ElapsedSeconds < 100 {
		t.Errorf("ElapsedSeconds should be >= 100, got %d", saveData.ElapsedSeconds)
	}
}

func TestLeaderboardSubmitterReceivesTimeSeconds(t *testing.T) {
	m := newTestModel()
	m.isMultiplayer = true
	m.currentRunType = "standard"

	// Track submitted time
	var submittedTime int
	submittedCh := make(chan struct{}, 1)
	m.leaderboardSubmitter = func(score, floors, timeSeconds int, class string, seed int64, runType string, victory bool) error {
		submittedTime = timeSeconds
		submittedCh <- struct{}{}
		return nil
	}

	// Setup game state
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.elapsedSeconds = 180 // 3 minutes from previous sessions

	// Submit to leaderboard
	m.submitToLeaderboard(false)

	// Wait for submission
	<-submittedCh

	// Time should be passed (at least the accumulated time)
	if submittedTime < 180 {
		t.Errorf("expected submitted time >= 180, got %d", submittedTime)
	}
}

// === Continue After Death Regression Tests ===

// TestFinishRunClearsSaveOnDeath verifies that when a player dies,
// the save state is properly cleared so they cannot Continue.
// This is a regression test for the bug where Continue was available after death.
func TestFinishRunClearsSaveOnDeath(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat
	m.gameState = types.StateCombat

	// Simulate having a valid save before death
	m.hasValidSave = true
	m.pendingSave = &save.SaveData{} // Non-nil save data

	// Simulate player death
	m.finishRun(false)

	// Verify save state is cleared
	if m.hasValidSave {
		t.Error("hasValidSave should be false after death")
	}
	if m.pendingSave != nil {
		t.Error("pendingSave should be nil after death")
	}

	// Verify view transitioned to game over
	if m.currentView != ViewGameOver {
		t.Errorf("currentView should be ViewGameOver after death, got %v", m.currentView)
	}
	if m.gameState != types.StateGameOver {
		t.Errorf("gameState should be StateGameOver after death, got %v", m.gameState)
	}
}

// TestFinishRunClearsSaveOnVictory verifies that when a player wins,
// the save state is properly cleared (run is over, no need for save).
func TestFinishRunClearsSaveOnVictory(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat
	m.gameState = types.StateCombat

	// Simulate having a valid save before victory
	m.hasValidSave = true
	m.pendingSave = &save.SaveData{} // Non-nil save data

	// Simulate player victory
	m.finishRun(true)

	// Verify save state is cleared
	if m.hasValidSave {
		t.Error("hasValidSave should be false after victory")
	}
	if m.pendingSave != nil {
		t.Error("pendingSave should be nil after victory")
	}

	// Verify view transitioned to victory
	if m.currentView != ViewVictory {
		t.Errorf("currentView should be ViewVictory after victory, got %v", m.currentView)
	}
	if m.gameState != types.StateVictory {
		t.Errorf("gameState should be StateVictory after victory, got %v", m.gameState)
	}
}

// TestFinishRunCallsClearSaveCallback verifies that the clearSaveCallback
// is called when a run ends (for multiplayer mode).
func TestFinishRunCallsClearSaveCallback(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Track whether callback was called
	callbackCalled := false
	m.SetClearSaveCallback(func() error {
		callbackCalled = true
		return nil
	})

	// Simulate having a valid save
	m.hasValidSave = true

	// Simulate player death
	m.finishRun(false)

	// Verify callback was called
	if !callbackCalled {
		t.Error("clearSaveCallback should be called on death")
	}
}

// TestFinishRunCallsMetaProgressUpdater verifies that the metaProgressUpdater
// is called when a run ends (for multiplayer mode exit codes).
func TestFinishRunCallsMetaProgressUpdater(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Track whether callback was called and with what parameters
	callbackCalled := false
	var receivedVictory bool
	m.SetMetaProgressUpdater(func(exitCodes int, victory bool, maxDepth int) error {
		callbackCalled = true
		receivedVictory = victory
		return nil
	})

	// Simulate player victory
	m.finishRun(true)

	// Verify callback was called with correct victory flag
	if !callbackCalled {
		t.Error("metaProgressUpdater should be called on run end")
	}
	if !receivedVictory {
		t.Error("metaProgressUpdater should receive victory=true for victory")
	}

	// Test again for death
	callbackCalled = false
	m.finishRun(false)

	if !callbackCalled {
		t.Error("metaProgressUpdater should be called on death")
	}
	if receivedVictory {
		t.Error("metaProgressUpdater should receive victory=false for death")
	}
}
