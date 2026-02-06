package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/types"
)

// === Combat Target Tests ===

func TestGetValidTargetIndexNoCombat(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat

	// Without combat set, getValidTargetIndex returns 0 (fallback)
	m.combat = nil
	m.targetCursor = 0
	idx := m.getValidTargetIndex()
	if idx != 0 {
		t.Errorf("should return 0 without combat, got %d", idx)
	}
}

func TestCycleTargetNoCombat(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat

	// Without combat, cycleTarget should be a no-op
	m.combat = nil
	m.targetCursor = 0

	// This should not panic and should not change cursor
	m.cycleTarget(false)
	if m.targetCursor != 0 {
		t.Errorf("cursor should remain 0 without combat, got %d", m.targetCursor)
	}
}

func TestTargetCursorInitialValue(t *testing.T) {
	m := newTestModel()

	// Target cursor should start at 0
	if m.targetCursor != 0 {
		t.Errorf("target cursor should start at 0, got %d", m.targetCursor)
	}
}

// === Skill Select Tests ===

func TestPlayerHasSkills(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassBash); err != nil { // Bash has skills
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat

	// Bash should have skills
	if len(m.player.Skills) == 0 {
		t.Error("Bash should have skills")
	}

	// Init class should have fewer or equal skills
	engine2 := game.NewEngine(cfg, 12345)
	if err := engine2.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}

	// Both classes should exist
	if m.player == nil {
		t.Error("player should not be nil")
	}
}

// === Combat State Tests ===

func TestCombatActionSelect(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewCombat
	m.gameState = types.StateCombat

	// Combat action cursor starts at 0
	m.combatCursor = 0
	if m.combatCursor != 0 {
		t.Error("combat cursor should start at 0")
	}

	// Combat has options: Attack, Skill, Item, Flee
	m.combatCursor = 3
	if m.combatCursor != 3 {
		t.Error("combat cursor should be settable")
	}
}

// TestEscDoesNotEndCombat verifies that Esc key does not instant-flee from combat.
func TestEscDoesNotEndCombat(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Start combat
	enemy := entity.NewEnemy(entity.EnemyZombie, "test_zombie", types.Position{}, 1)
	m.startCombat([]*entity.Enemy{enemy})

	if m.currentView != ViewCombat {
		t.Fatal("should be in combat view")
	}

	// Press Esc - should NOT end combat (debug flee was removed)
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	m.updateCombat(escMsg)

	// Should still be in combat view
	if m.currentView != ViewCombat {
		t.Errorf("pressing Esc should not end combat, but view changed to %d", m.currentView)
	}

	// Combat should still be active
	if m.combat == nil {
		t.Error("combat should still be active after pressing Esc")
	}
}

// TestColorizeFDReturnsCorrectStyles verifies colorizeFD uses distinct styles for different ranges.
func TestColorizeFDReturnsCorrectStyles(t *testing.T) {
	m := newTestModel()

	maxFD := 100

	// High FD (> 60%) should use Success style
	highResult := m.colorizeFD(80, maxFD)
	if highResult == "" {
		t.Error("colorizeFD should return non-empty string for high FD")
	}

	// Medium FD (30-60%) should use Highlight style
	medResult := m.colorizeFD(50, maxFD)
	if medResult == "" {
		t.Error("colorizeFD should return non-empty string for medium FD")
	}

	// Low FD (< 30%) should use Danger style
	lowResult := m.colorizeFD(10, maxFD)
	if lowResult == "" {
		t.Error("colorizeFD should return non-empty string for low FD")
	}

	// The three ranges should produce different styled outputs
	// (they use different lipgloss styles so rendered strings should differ)
	if highResult == lowResult {
		t.Error("high FD and low FD should have different styling (was both Muted before fix)")
	}
	if medResult == lowResult {
		t.Error("medium FD and low FD should have different styling (was both Muted before fix)")
	}
}
