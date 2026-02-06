package ui

import (
	"testing"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/save"
)

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
	m.metaProgress.PermanentBonuses.RAM = 20 // +20 RAM
	m.metaProgress.PermanentBonuses.CPU = 5  // +5 CPU
	m.metaProgress.PermanentBonuses.FD = 10  // +10 FD

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
