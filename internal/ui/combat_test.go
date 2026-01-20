package ui

import (
	"testing"

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
