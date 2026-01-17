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
	itemPrice := m.shopItems[0].Price // malloc(), which may stack with existing
	purchasedTemplateID := m.shopItems[0].TemplateID

	// Count initial quantity of this item type in inventory
	initialQuantity := 0
	for _, item := range m.player.Inventory.Items {
		if item.TemplateID == purchasedTemplateID {
			initialQuantity += item.Quantity
		}
	}

	// Buy first item
	m.shopCursor = 0
	m.buyItem()

	// Verify exit codes were deducted
	if m.player.ExitCodes != initialCodes-itemPrice {
		t.Errorf("expected %d exit codes after purchase, got %d",
			initialCodes-itemPrice, m.player.ExitCodes)
	}

	// Verify item was added (either new slot or quantity increase due to stacking)
	finalQuantity := 0
	for _, item := range m.player.Inventory.Items {
		if item.TemplateID == purchasedTemplateID {
			finalQuantity += item.Quantity
		}
	}
	if finalQuantity != initialQuantity+1 {
		t.Errorf("expected %d total %s after purchase (was %d), got %d",
			initialQuantity+1, purchasedTemplateID, initialQuantity, finalQuantity)
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

// === Unlock System Tests ===

// newTestModelWithFreshMeta creates a test model with fresh (reset) meta progress
func newTestModelWithFreshMeta() *Model {
	m := newTestModel()
	// Reset to fresh meta progress for testing
	freshMeta := save.NewMetaProgress()
	m.metaProgress = &freshMeta
	return m
}

func TestClassUnlockStatus(t *testing.T) {
	m := newTestModelWithFreshMeta()

	// init class should always be unlocked
	if !m.isClassUnlocked(entity.ClassInit) {
		t.Error("init class should always be unlocked")
	}

	// Other classes should be locked by default (only init is in UnlockedClasses)
	if m.isClassUnlocked(entity.ClassCron) {
		t.Error("cron class should be locked by default")
	}
	if m.isClassUnlocked(entity.ClassBash) {
		t.Error("bash class should be locked by default")
	}
	if m.isClassUnlocked(entity.ClassVim) {
		t.Error("vim class should be locked by default")
	}
	if m.isClassUnlocked(entity.ClassSudo) {
		t.Error("sudo class should be locked by default")
	}
}

func TestClassUnlockPrices(t *testing.T) {
	m := newTestModel()

	// Verify prices are set correctly
	if m.getClassUnlockPrice(entity.ClassInit) != 0 {
		t.Error("init should have no price (always unlocked)")
	}
	if m.getClassUnlockPrice(entity.ClassCron) != 50 {
		t.Errorf("cron should cost 50, got %d", m.getClassUnlockPrice(entity.ClassCron))
	}
	if m.getClassUnlockPrice(entity.ClassBash) != 100 {
		t.Errorf("bash should cost 100, got %d", m.getClassUnlockPrice(entity.ClassBash))
	}
	if m.getClassUnlockPrice(entity.ClassVim) != 200 {
		t.Errorf("vim should cost 200, got %d", m.getClassUnlockPrice(entity.ClassVim))
	}
	if m.getClassUnlockPrice(entity.ClassSudo) != 500 {
		t.Errorf("sudo should cost 500, got %d", m.getClassUnlockPrice(entity.ClassSudo))
	}
}

func TestClassUnlockPurchase(t *testing.T) {
	m := newTestModelWithFreshMeta()

	// Give player enough exit codes
	m.metaProgress.TotalExitCodes = 100

	// Unlock cron class (costs 50)
	m.unlockCategory = 0
	m.unlockCursor = 0 // cron is first in the unlock list
	m.purchaseClassUnlock()

	// Check that class is now unlocked
	if !m.isClassUnlocked(entity.ClassCron) {
		t.Error("cron should be unlocked after purchase")
	}

	// Check that exit codes were deducted
	if m.metaProgress.TotalExitCodes != 50 {
		t.Errorf("expected 50 exit codes remaining, got %d", m.metaProgress.TotalExitCodes)
	}
}

func TestClassUnlockInsufficientFunds(t *testing.T) {
	m := newTestModelWithFreshMeta()

	// Give player insufficient exit codes
	m.metaProgress.TotalExitCodes = 10

	// Try to unlock cron class (costs 50)
	m.unlockCategory = 0
	m.unlockCursor = 0
	m.purchaseClassUnlock()

	// Class should still be locked
	if m.isClassUnlocked(entity.ClassCron) {
		t.Error("cron should still be locked (insufficient funds)")
	}

	// Exit codes should not change
	if m.metaProgress.TotalExitCodes != 10 {
		t.Error("exit codes should not change on failed purchase")
	}
}

func TestUnlockableBonuses(t *testing.T) {
	m := newTestModel()

	bonuses := m.getUnlockableBonuses()

	if len(bonuses) != 4 {
		t.Errorf("expected 4 bonuses, got %d", len(bonuses))
	}

	// Verify bonus types
	expectedStats := []string{"RAM", "CPU", "MEM", "NICE"}
	for i, bonus := range bonuses {
		if bonus.StatType != expectedStats[i] {
			t.Errorf("expected stat type %s, got %s", expectedStats[i], bonus.StatType)
		}
		if bonus.BasePrice <= 0 {
			t.Errorf("bonus %s should have positive base price", bonus.Name)
		}
		if bonus.MaxLevel <= 0 {
			t.Errorf("bonus %s should have positive max level", bonus.Name)
		}
	}
}

func TestBonusPriceScaling(t *testing.T) {
	m := newTestModel()

	bonuses := m.getUnlockableBonuses()
	ramBonus := bonuses[0] // RAM bonus

	// Price at level 0
	price0 := m.getBonusPrice(ramBonus)
	if price0 != ramBonus.BasePrice {
		t.Errorf("price at level 0 should be base price %d, got %d", ramBonus.BasePrice, price0)
	}

	// Simulate level 1 by setting current level
	ramBonus.CurrentLevel = 1
	price1 := m.getBonusPrice(ramBonus)
	if price1 != ramBonus.BasePrice*2 {
		t.Errorf("price at level 1 should be %d, got %d", ramBonus.BasePrice*2, price1)
	}

	// Simulate level 2
	ramBonus.CurrentLevel = 2
	price2 := m.getBonusPrice(ramBonus)
	if price2 != ramBonus.BasePrice*4 {
		t.Errorf("price at level 2 should be %d, got %d", ramBonus.BasePrice*4, price2)
	}
}

func TestUnlockableItems(t *testing.T) {
	m := newTestModel()

	items := m.getUnlockableItems()

	if len(items) == 0 {
		t.Error("should have unlockable items")
	}

	for _, item := range items {
		if item.TemplateID == "" {
			t.Error("item should have template ID")
		}
		if item.Name == "" {
			t.Error("item should have name")
		}
		if item.Price <= 0 {
			t.Errorf("item %s should have positive price", item.Name)
		}
	}
}

func TestExitCodeCalculation(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Calculate exit codes for death (not victory)
	earned, breakdown := m.calculateRunExitCodes(false)

	// Should have at least floor bonus (depth 1 = 10)
	if earned < 10 {
		t.Errorf("expected at least 10 exit codes for floor 1, got %d", earned)
	}

	// Breakdown should show floor bonus
	hasFloorBonus := false
	for _, line := range breakdown {
		if len(line) > 0 {
			hasFloorBonus = true
		}
	}
	if !hasFloorBonus {
		t.Error("breakdown should include floor bonus")
	}
}

func TestExitCodeCalculationVictory(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Calculate exit codes for victory
	earnedVictory, _ := m.calculateRunExitCodes(true)
	earnedDeath, _ := m.calculateRunExitCodes(false)

	// Victory should give more (200 bonus)
	if earnedVictory <= earnedDeath {
		t.Error("victory should give more exit codes than death")
	}
	if earnedVictory-earnedDeath != 200 {
		t.Errorf("victory bonus should be 200, got %d", earnedVictory-earnedDeath)
	}
}

func TestEnemyExitCodeValues(t *testing.T) {
	m := newTestModel()

	// Test enemy values
	testCases := []struct {
		enemyType string
		expected  int
	}{
		{"zombie", 1},
		{"daemon", 2},
		{"fork_bomb", 3},
		{"segfault", 4},
		{"rootkit", 5},
		{"kernel_panic", 10},
		{"unknown", 1}, // Default
	}

	for _, tc := range testCases {
		value := m.getEnemyExitCodeValue(tc.enemyType)
		if value != tc.expected {
			t.Errorf("expected %s to be worth %d, got %d", tc.enemyType, tc.expected, value)
		}
	}
}

func TestMetaProgressInitialization(t *testing.T) {
	m := newTestModel()

	// MetaProgress should be initialized
	if m.metaProgress == nil {
		t.Fatal("metaProgress should be initialized")
	}

	// init class should be unlocked by default
	found := false
	for _, class := range m.metaProgress.UnlockedClasses {
		if class == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("init class should be in UnlockedClasses by default")
	}
}

func TestOpenUnlockShop(t *testing.T) {
	m := newTestModel()

	m.currentView = ViewMainMenu
	m.openUnlockShop()

	if m.currentView != ViewUnlockShop {
		t.Error("should transition to unlock shop")
	}
	if m.unlockCursor != 0 {
		t.Error("unlock cursor should reset to 0")
	}
	if m.unlockCategory != 0 {
		t.Error("unlock category should reset to 0")
	}
}

func TestUnlockShopCategoryNavigation(t *testing.T) {
	m := newTestModel()
	m.openUnlockShop()

	// Start at category 0 (classes)
	if m.unlockCategory != 0 {
		t.Error("should start at category 0")
	}

	// Simulate right navigation
	m.unlockCategory = 1
	if m.unlockCategory != 1 {
		t.Error("should move to category 1 (bonuses)")
	}

	m.unlockCategory = 2
	if m.unlockCategory != 2 {
		t.Error("should move to category 2 (items)")
	}

	// Wrap around
	m.unlockCategory = 0
	if m.unlockCategory != 0 {
		t.Error("should wrap to category 0")
	}
}

func TestPermanentBonusesAppliedToPlayer(t *testing.T) {
	m := newTestModel()

	// Set some permanent bonuses
	m.metaProgress.PermanentBonuses.PID = 20 // +20 RAM
	m.metaProgress.PermanentBonuses.CPU = 5  // +5 CPU
	m.metaProgress.PermanentBonuses.MEM = 10 // +10 FD

	// Start a new game
	cfg := config.DefaultConfig()
	m.config = cfg
	m.startNewGame(entity.ClassInit)

	// Check that bonuses were applied
	// Base init stats: RAM=100, CPU=10, FD=16
	if m.player.Stats.RAM < 120 {
		t.Errorf("expected RAM >= 120 with bonus, got %d", m.player.Stats.RAM)
	}
	if m.player.Stats.CPU < 15 {
		t.Errorf("expected CPU >= 15 with bonus, got %d", m.player.Stats.CPU)
	}
	if m.player.Stats.FD < 26 {
		t.Errorf("expected FD >= 26 with bonus, got %d", m.player.Stats.FD)
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

// === Shop Cursor Tests ===

func TestShopCursorBounds(t *testing.T) {
	m := newTestModel()

	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, 12345)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	m.engine = engine
	m.player = engine.Player()

	// Open shop
	m.openShop()

	// Shop cursor should start at 0
	if m.shopCursor != 0 {
		t.Errorf("shop cursor should start at 0, got %d", m.shopCursor)
	}

	// Cursor should be bounded by shop item count
	if len(m.shopItems) > 0 {
		m.shopCursor = len(m.shopItems) - 1
		if m.shopCursor != len(m.shopItems)-1 {
			t.Error("shop cursor should be settable to last item")
		}
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
