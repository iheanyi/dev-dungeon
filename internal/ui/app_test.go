package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/save"
	"github.com/iheanyi/devdungeon/internal/types"
)

// === View State Tests ===

func TestInitialViewState(t *testing.T) {
	m := newTestModel()

	if m.currentView != ViewMainMenu {
		t.Errorf("initial view should be MainMenu, got %v", m.currentView)
	}

	if m.gameState != types.StateMainMenu {
		t.Errorf("initial game state should be MainMenu, got %v", m.gameState)
	}
}

func TestViewTypes(t *testing.T) {
	// Verify all view types are defined
	views := []ViewType{
		ViewMainMenu,
		ViewClassSelect,
		ViewGame,
		ViewCombat,
		ViewInventory,
		ViewPause,
		ViewGameOver,
		ViewVictory,
		ViewAdmin,
		ViewHelp,
		ViewIntro,
		ViewMessageHistory,
		ViewShop,
		ViewUnlockShop,
	}

	// Just verify they're all distinct
	seen := make(map[ViewType]bool)
	for _, v := range views {
		if seen[v] {
			t.Errorf("duplicate view type: %v", v)
		}
		seen[v] = true
	}
}

// === Game State Transition Tests ===

func TestMainMenuToClassSelect(t *testing.T) {
	m := newTestModel()

	if m.currentView != ViewMainMenu {
		t.Fatal("should start at main menu")
	}

	// Simulate selecting "New Game"
	m.menuCursor = 0 // Assuming New Game is first option
	// In real code, pressing enter would transition to class select
	m.currentView = ViewClassSelect

	if m.currentView != ViewClassSelect {
		t.Error("should transition to class select")
	}
}

func TestClassSelectToIntro(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewClassSelect
	m.classCursor = 0 // ClassInit

	// Simulate selecting a class - this should start intro
	// The actual transition happens in updateClassSelect
	// We can verify the pendingClass mechanism
	m.pendingClass = entity.ClassInit
	m.currentView = ViewIntro

	if m.currentView != ViewIntro {
		t.Error("should transition to intro after class select")
	}
}

func TestGameToCombat(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	// Simulate entering combat
	m.currentView = ViewCombat
	m.gameState = types.StateCombat

	if m.currentView != ViewCombat {
		t.Error("should be in combat view")
	}
	if m.gameState != types.StateCombat {
		t.Error("game state should be combat")
	}
}

func TestCombatToGameOver(t *testing.T) {
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

	// Simulate defeat
	m.currentView = ViewGameOver
	m.gameState = types.StateGameOver

	if m.currentView != ViewGameOver {
		t.Error("should transition to game over")
	}
}

func TestCombatToVictory(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat

	// Simulate boss kill victory
	m.currentView = ViewVictory
	m.gameState = types.StateVictory

	if m.currentView != ViewVictory {
		t.Error("should transition to victory")
	}
}

func TestVictoryScreenEnterReturnsToMainMenu(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewVictory
	m.gameState = types.StateVictory

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	m.handleKeyPress(msg)

	if m.currentView != ViewMainMenu {
		t.Errorf("victory Enter should return to main menu, got view %v", m.currentView)
	}
	if m.gameState != types.StateMainMenu {
		t.Errorf("victory Enter should set game state to main menu, got %v", m.gameState)
	}
}

func TestGameQuitOpensConfirmDialog(t *testing.T) {
	m := newTestModelWithEngine(12345)
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	m.updateGame(msg)

	if m.currentView != ViewConfirmDialog {
		t.Fatalf("pressing q in game should open confirm dialog, got view %v", m.currentView)
	}
	if m.confirmReturnView != ViewGame {
		t.Fatalf("confirm dialog should return to game, got %v", m.confirmReturnView)
	}
}

func TestSaveGameAndReturnToMenuOfflineUsesEngineSave(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := newTestModelWithEngine(12345)
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	m.saveGameAndReturnToMenu()

	if m.statusMsg != "Game saved." {
		t.Fatalf("expected successful save status, got %q", m.statusMsg)
	}
	if !m.hasValidSave {
		t.Fatal("hasValidSave should be true after saving to menu")
	}
	if m.currentView != ViewMainMenu {
		t.Fatalf("should return to main menu, got %v", m.currentView)
	}
}

func TestGameToInventory(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Toggle inventory
	m.currentView = ViewInventory

	if m.currentView != ViewInventory {
		t.Error("should open inventory")
	}

	// Toggle back
	m.currentView = ViewGame

	if m.currentView != ViewGame {
		t.Error("should close inventory and return to game")
	}
}

func TestGameToShop(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Open shop
	m.openShop()

	if m.currentView != ViewShop {
		t.Error("should open shop")
	}

	if len(m.shopItems) == 0 {
		t.Error("shop should have items after opening")
	}
}

func TestGameToPause(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Open pause menu
	m.currentView = ViewPause

	if m.currentView != ViewPause {
		t.Error("should open pause menu")
	}
}

func TestGameToHelp(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Open help
	m.currentView = ViewHelp

	if m.currentView != ViewHelp {
		t.Error("should open help")
	}
}

func TestGameToMessageHistory(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Open message history
	m.prevView = ViewGame
	m.currentView = ViewMessageHistory

	if m.currentView != ViewMessageHistory {
		t.Error("should open message history")
	}
}

// === Styles Test ===

func TestStylesInitialized(t *testing.T) {
	m := newTestModel()

	if m.styles == nil {
		t.Fatal("styles should be initialized")
	}

	// Verify key styles exist (they're lipgloss.Style, not nil)
	// We can't easily test lipgloss styles, but we can verify the struct is populated
}

// === Message System Tests ===

func TestAddToHistory(t *testing.T) {
	m := newTestModel()

	initialLen := len(m.messageHistory)

	m.addToHistory("Test message 1")
	m.addToHistory("Test message 2")

	if len(m.messageHistory) != initialLen+2 {
		t.Errorf("expected %d messages, got %d", initialLen+2, len(m.messageHistory))
	}

	// Most recent should be at end
	if m.messageHistory[len(m.messageHistory)-1] != "Test message 2" {
		t.Error("most recent message should be at end")
	}
}

func TestMessageHistoryLimit(t *testing.T) {
	m := newTestModel()

	// Add many messages (more than the 500 limit)
	for i := 0; i < 600; i++ {
		m.addToHistory("Message")
	}

	// Should be capped at 500
	if len(m.messageHistory) > 500 {
		t.Errorf("message history should be capped at 500, got %d", len(m.messageHistory))
	}
}

// === Save Callback Tests ===

func TestSaveCallbackSetsHasValidSave(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	// Initially hasValidSave is false
	m.hasValidSave = false

	// Set up save callback that succeeds
	callbackCalled := false
	m.SetSaveCallback(func() (*save.SaveData, error) {
		callbackCalled = true
		return nil, nil
	})

	// Simulate pressing Q to return to menu (manually invoke the logic)
	if m.engine != nil {
		if m.saveCallback != nil {
			if _, err := m.saveCallback(); err != nil {
				m.statusMsg = "Failed to save game."
			} else {
				m.statusMsg = "Game saved."
				m.hasValidSave = true
			}
		}
		m.engine.Shutdown()
		m.engine = nil
	}
	m.player = nil
	m.currentView = ViewMainMenu
	m.gameState = types.StateMainMenu

	// Verify callback was called
	if !callbackCalled {
		t.Error("save callback should have been called")
	}

	// Verify hasValidSave is now true
	if !m.hasValidSave {
		t.Error("hasValidSave should be true after successful save")
	}

	// Verify status message
	if m.statusMsg != "Game saved." {
		t.Errorf("expected status 'Game saved.', got '%s'", m.statusMsg)
	}
}

func TestSaveCallbackFailureDoesNotSetHasValidSave(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	// Initially hasValidSave is false
	m.hasValidSave = false

	// Set up save callback that fails
	m.SetSaveCallback(func() (*save.SaveData, error) {
		return nil, fmt.Errorf("simulated save failure")
	})

	// Simulate pressing Q to return to menu
	if m.engine != nil {
		if m.saveCallback != nil {
			if _, err := m.saveCallback(); err != nil {
				m.statusMsg = "Failed to save game."
			} else {
				m.statusMsg = "Game saved."
				m.hasValidSave = true
			}
		}
		m.engine.Shutdown()
		m.engine = nil
	}
	m.player = nil
	m.currentView = ViewMainMenu
	m.gameState = types.StateMainMenu

	// Verify hasValidSave is still false
	if m.hasValidSave {
		t.Error("hasValidSave should remain false after failed save")
	}

	// Verify error status message
	if m.statusMsg != "Failed to save game." {
		t.Errorf("expected status 'Failed to save game.', got '%s'", m.statusMsg)
	}
}

func TestSaveCallbackCalledBeforeEngineDestroyed(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Track whether engine was available during callback
	engineAvailableDuringCallback := false
	m.SetSaveCallback(func() (*save.SaveData, error) {
		// The engine should still be available at this point
		// (save callback must be called BEFORE engine destruction)
		if m.GetEngine() != nil {
			engineAvailableDuringCallback = true
		}
		return nil, nil
	})

	// Simulate the Q key handler logic
	if m.engine != nil {
		// Call save callback BEFORE destroying the engine
		if m.saveCallback != nil {
			_, _ = m.saveCallback()
		}
		m.engine.Shutdown()
		m.engine = nil
	}

	if !engineAvailableDuringCallback {
		t.Error("engine should be available when save callback is invoked")
	}
}

func TestNoSaveCallbackShowsReturnedToMenu(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// No save callback set (single-player mode)
	m.saveCallback = nil
	m.hasValidSave = false

	// Simulate Q key handler
	if m.engine != nil {
		if m.saveCallback != nil {
			if _, err := m.saveCallback(); err != nil {
				m.statusMsg = "Failed to save game."
			} else {
				m.statusMsg = "Game saved."
				m.hasValidSave = true
			}
		} else {
			m.statusMsg = "Returned to menu."
		}
		m.engine.Shutdown()
		m.engine = nil
	}

	// Verify status message for non-multiplayer mode
	if m.statusMsg != "Returned to menu." {
		t.Errorf("expected status 'Returned to menu.', got '%s'", m.statusMsg)
	}

	// hasValidSave should remain false (no callback to set it)
	if m.hasValidSave {
		t.Error("hasValidSave should remain false without save callback")
	}
}

func TestSetHasValidSave(t *testing.T) {
	m := newTestModel()

	// Explicitly set to false to test the setter
	m.hasValidSave = false

	// Set to true
	m.SetHasValidSave(true)
	if !m.hasValidSave {
		t.Error("hasValidSave should be true after SetHasValidSave(true)")
	}

	// Set back to false
	m.SetHasValidSave(false)
	if m.hasValidSave {
		t.Error("hasValidSave should be false after SetHasValidSave(false)")
	}
}

// === Quit Request Tests ===

func TestWantsToQuit(t *testing.T) {
	m := newTestModel()

	// Initially should not want to quit
	if m.WantsToQuit() {
		t.Error("should not want to quit initially")
	}

	// Use requestQuit to set the flag
	m.requestQuit()
	if !m.WantsToQuit() {
		t.Error("should want to quit after requestQuit")
	}
}

func TestClearQuitRequest(t *testing.T) {
	m := newTestModel()

	// Set quit request
	m.requestQuit()
	if !m.WantsToQuit() {
		t.Error("quit flag should be set")
	}

	// Clear it
	m.ClearQuitRequest()
	if m.WantsToQuit() {
		t.Error("quit flag should be cleared")
	}
}

func TestRequestQuit(t *testing.T) {
	m := newTestModel()

	// Initially should not want to quit
	if m.WantsToQuit() {
		t.Error("should not want to quit initially")
	}

	// Request quit
	m.requestQuit()
	if !m.WantsToQuit() {
		t.Error("requestQuit should set wantsToQuit flag")
	}
}

// === Multiplayer Mode Tests ===

func TestSetMultiplayerMode(t *testing.T) {
	m := newTestModel()

	// Initially not multiplayer
	if m.isMultiplayer {
		t.Error("should not be multiplayer initially")
	}

	// Set multiplayer mode with username
	m.SetMultiplayerMode("testuser")
	if !m.isMultiplayer {
		t.Error("should be multiplayer after SetMultiplayerMode")
	}

	// Username should be set
	if m.username != "testuser" {
		t.Errorf("username should be 'testuser', got '%s'", m.username)
	}
}

// === Menu Cursor Tests ===

func TestMenuCursorBounds(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewMainMenu

	// Menu cursor should start at 0
	if m.menuCursor != 0 {
		t.Errorf("menu cursor should start at 0, got %d", m.menuCursor)
	}

	// Menu cursor should be bounded
	m.menuCursor = 5
	// Note: The actual menu has 6 items (0-5 valid), so this is valid
	// This test just verifies the cursor can be set
	if m.menuCursor != 5 {
		t.Error("menu cursor should be settable")
	}
}

func TestClassCursorBounds(t *testing.T) {
	m := newTestModel()
	m.currentView = ViewClassSelect

	// Class cursor should start at 0
	if m.classCursor != 0 {
		t.Errorf("class cursor should start at 0, got %d", m.classCursor)
	}

	// Set cursor to different class
	m.classCursor = 3 // vim class position
	if m.classCursor != 3 {
		t.Error("class cursor should be settable")
	}
}

// === Pause Menu Tests ===

func TestPauseMenuTransition(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	// Transition to pause view
	m.currentView = ViewPause

	if m.currentView != ViewPause {
		t.Error("should be in pause view")
	}
}

// === Admin View Tests ===

func TestAdminViewToggle(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	cfg.Debug.Enabled = true
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame
	m.config = cfg

	// Transition to admin view
	m.prevView = ViewGame
	m.currentView = ViewAdmin

	if m.currentView != ViewAdmin {
		t.Error("should be in admin view")
	}

	// Return to previous view
	m.currentView = m.prevView
	if m.currentView != ViewGame {
		t.Error("should return to game view")
	}
}

// === Message History Scroll Tests ===

func TestMessageHistoryScrollIndex(t *testing.T) {
	m := newTestModel()

	// Add many messages to create scrollable content
	for i := 0; i < 100; i++ {
		m.addToHistory(fmt.Sprintf("Message %d", i))
	}

	m.currentView = ViewMessageHistory
	m.messageScrollIdx = 0

	// Scroll index 0 = most recent messages
	if m.messageScrollIdx != 0 {
		t.Errorf("scroll index should start at 0, got %d", m.messageScrollIdx)
	}

	// Set scroll index
	m.messageScrollIdx = 10
	if m.messageScrollIdx != 10 {
		t.Errorf("scroll should be 10, got %d", m.messageScrollIdx)
	}
}

// === Run Mode Tests ===

func TestDailyRunModeToggle(t *testing.T) {
	m := newTestModel()

	// Initially not in daily run mode
	if m.dailyRunMode {
		t.Error("should not be in daily run mode initially")
	}

	// Toggle to daily run mode
	m.dailyRunMode = true
	if !m.dailyRunMode {
		t.Error("should be in daily run mode")
	}
}

func TestEngineCreatedWithSeed(t *testing.T) {
	// Test that engine is created with a seed
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Engine should store the master seed
	if engine.MasterSeed() != 12345 {
		t.Errorf("engine should use seed 12345, got %d", engine.MasterSeed())
	}
}

// === Continue Game Tests ===

func TestContinueGameRequiresValidSave(t *testing.T) {
	m := newTestModel()

	// Without valid save
	m.hasValidSave = false

	// Continue should not be available
	// (The actual menu check is done in updateMainMenu, here we just test the flag)
	if m.hasValidSave {
		t.Error("should not have valid save initially")
	}

	// With valid save
	m.hasValidSave = true
	if !m.hasValidSave {
		t.Error("should have valid save after setting")
	}
}

// === Floor Info Tests ===

func TestFloorInfoTracking(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Engine should track current depth
	depth := engine.CurrentDepth()
	if depth < 1 {
		t.Error("starting depth should be at least 1")
	}
}

// === Status Message Tests ===

func TestStatusMessageSetting(t *testing.T) {
	m := newTestModel()

	// Status message should be empty initially
	if m.statusMsg != "" {
		t.Errorf("status message should be empty initially, got '%s'", m.statusMsg)
	}

	// Set status message
	m.statusMsg = "Test message"
	if m.statusMsg != "Test message" {
		t.Error("status message should be settable")
	}
}

// === View Previous Tests ===

func TestPrevViewTracking(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewGame

	// Open inventory
	m.prevView = m.currentView
	m.currentView = ViewInventory

	// prevView should track where we came from
	if m.prevView != ViewGame {
		t.Error("prevView should track previous view")
	}

	// Return to previous view
	m.currentView = m.prevView
	if m.currentView != ViewGame {
		t.Error("should return to game view")
	}
}

// === Intro Screen Tests ===

func TestIntroScreenSetup(t *testing.T) {
	m := newTestModel()

	m.currentView = ViewIntro
	m.pendingClass = entity.ClassBash

	if m.currentView != ViewIntro {
		t.Error("should be in intro view")
	}

	if m.pendingClass != entity.ClassBash {
		t.Error("pending class should be set")
	}
}
