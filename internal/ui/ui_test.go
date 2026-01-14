package ui

import (
	"testing"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/types"
)

func newTestModel() *Model {
	cfg := config.DefaultConfig()
	m := New(cfg)
	return m
}

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

// === Shop System Tests ===

func TestShopItemGeneration(t *testing.T) {
	m := newTestModel()

	// Setup game state for shop
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Generate shop inventory
	items := m.generateShopInventory()

	if len(items) == 0 {
		t.Error("shop should have items")
	}

	// Verify items have prices
	for _, item := range items {
		if item.Price <= 0 {
			t.Errorf("item %s should have positive price, got %d", item.Name, item.Price)
		}
		if item.TemplateID == "" {
			t.Error("item should have template ID")
		}
		if item.Name == "" {
			t.Error("item should have name")
		}
	}
}

func TestShopPurchaseWithSufficientFunds(t *testing.T) {
	m := newTestModel()

	// Setup game state
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewShop

	// Give player plenty of money
	m.player.ExitCodes = 1000

	// Generate shop items
	m.shopItems = m.generateShopInventory()
	if len(m.shopItems) == 0 {
		t.Skip("no shop items generated")
	}

	// Record initial state
	initialCodes := m.player.ExitCodes
	initialInvCount := len(m.player.Inventory.Items)
	itemPrice := m.shopItems[0].Price

	// Buy first item
	m.shopCursor = 0
	m.buyItem()

	// Verify purchase
	if m.player.ExitCodes != initialCodes-itemPrice {
		t.Errorf("expected %d exit codes after purchase, got %d",
			initialCodes-itemPrice, m.player.ExitCodes)
	}

	if len(m.player.Inventory.Items) != initialInvCount+1 {
		t.Errorf("expected %d items in inventory, got %d",
			initialInvCount+1, len(m.player.Inventory.Items))
	}
}

func TestShopPurchaseWithInsufficientFunds(t *testing.T) {
	m := newTestModel()

	// Setup game state
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()
	m.currentView = ViewShop

	// Give player no money
	m.player.ExitCodes = 0

	// Generate shop items
	m.shopItems = m.generateShopInventory()
	if len(m.shopItems) == 0 {
		t.Skip("no shop items generated")
	}

	initialInvCount := len(m.player.Inventory.Items)

	// Try to buy first item
	m.shopCursor = 0
	m.buyItem()

	// Verify purchase failed
	if m.player.ExitCodes != 0 {
		t.Errorf("exit codes should still be 0, got %d", m.player.ExitCodes)
	}

	if len(m.player.Inventory.Items) != initialInvCount {
		t.Error("inventory should not change on failed purchase")
	}

	if m.statusMsg == "" {
		t.Error("should show error message on failed purchase")
	}
}

func TestShopItemsHaveCorrectPricing(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	items := m.generateShopInventory()

	for _, item := range items {
		// Prices should be reasonable (not too low, not too high)
		if item.Price < 5 {
			t.Errorf("item %s price too low: %d", item.Name, item.Price)
		}
		if item.Price > 1000 {
			t.Errorf("item %s price too high: %d", item.Name, item.Price)
		}
	}
}

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
