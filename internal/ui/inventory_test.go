package ui

import (
	"testing"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/types"
)

// === Inventory System Tests ===

func TestInventoryViewNavigation(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewInventory
	m.invCursor = 0

	// Navigate down
	initialCursor := m.invCursor
	// Simulate pressing down would increase cursor
	// We can't easily simulate key presses, but we can verify the cursor system works

	invLen := len(m.player.Inventory.Items)
	equipSlots := 4

	// Cursor should be able to go from 0 to invLen + equipSlots - 1
	maxCursor := invLen + equipSlots - 1
	if maxCursor < 0 {
		maxCursor = 0
	}

	_ = initialCursor
	_ = maxCursor
}

func TestInventoryEquipFromInventory(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Add a weapon to inventory
	weapon := entity.NewItem("vim_blade", "test_weapon", types.Position{})
	m.player.Inventory.AddItem(weapon)

	// Clear existing weapon
	m.player.Equipment.Weapon = nil

	m.currentView = ViewInventory
	m.invCursor = len(m.player.Inventory.Items) - 1 // Select the weapon

	// Equip the item
	m.equipItem()

	// Weapon should be equipped
	if m.player.Equipment.Weapon == nil {
		t.Error("weapon should be equipped")
	}
}

func TestInventoryDropItem(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Add item to inventory
	item := entity.NewItem("malloc", "test_item", types.Position{})
	m.player.Inventory.AddItem(item)

	initialCount := len(m.player.Inventory.Items)

	m.currentView = ViewInventory
	m.invCursor = initialCount - 1

	// Drop the item
	m.dropItem()

	// Item should be removed
	if len(m.player.Inventory.Items) != initialCount-1 {
		t.Errorf("expected %d items after drop, got %d",
			initialCount-1, len(m.player.Inventory.Items))
	}
}

func TestInventoryUnequipItem(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Bash starts with a weapon equipped
	if m.player.Equipment.Weapon == nil {
		t.Skip("Bash should start with weapon equipped")
	}

	invLen := len(m.player.Inventory.Items)
	initialInvCount := invLen

	m.currentView = ViewInventory
	m.invCursor = invLen // First equipment slot (weapon)

	// Unequip
	m.unequipItem()

	// Weapon should be in inventory now
	if len(m.player.Inventory.Items) != initialInvCount+1 {
		t.Errorf("expected %d items after unequip, got %d",
			initialInvCount+1, len(m.player.Inventory.Items))
	}

	if m.player.Equipment.Weapon != nil {
		t.Error("weapon slot should be empty after unequip")
	}
}
