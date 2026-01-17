package entity

import (
	"testing"

	"github.com/iheanyi/devdungeon/internal/types"
)

func TestNewItem(t *testing.T) {
	// Test a known item
	item := NewItem("malloc", "test_malloc", types.Position{X: 1, Y: 2})
	if item == nil {
		t.Fatal("NewItem returned nil for valid template")
	}
	if item.TemplateID != "malloc" {
		t.Errorf("expected TemplateID 'malloc', got '%s'", item.TemplateID)
	}
	if item.ItemType != ItemTypeConsumable {
		t.Errorf("malloc should be consumable, got %s", item.ItemType)
	}
	if item.Position() != (types.Position{X: 1, Y: 2}) {
		t.Errorf("unexpected position %v", item.Position())
	}

	// Test unknown item
	unknown := NewItem("nonexistent_item", "test_id", types.Position{})
	if unknown != nil {
		t.Error("NewItem should return nil for unknown template")
	}
}

func TestItemTemplates(t *testing.T) {
	// Verify some key items exist in templates
	expectedItems := []string{
		"malloc", "realloc", "mmap", // Consumables
		"basic_script", "vim_blade", // Weapons
		"basic_shell", "firewall", // Armor
		"ssh_key", "env_vars", // Utilities
		"nice_boost", "cpu_boost", // Buffs
	}

	for _, itemID := range expectedItems {
		_, exists := ItemTemplates[itemID]
		if !exists {
			t.Errorf("expected item template '%s' to exist", itemID)
		}
	}
}

func TestItemRarityString(t *testing.T) {
	tests := []struct {
		rarity   ItemRarity
		expected string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityEpic, "Epic"},
		{RarityLegendary, "Legendary"},
		{ItemRarity(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.rarity.String(); got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestInventory(t *testing.T) {
	inv := NewInventory(5)

	if len(inv.Items) != 0 {
		t.Error("new inventory should be empty")
	}
	if inv.MaxSlots != 5 {
		t.Errorf("expected max slots 5, got %d", inv.MaxSlots)
	}

	// Add items
	item1 := NewItem("malloc", "item1", types.Position{})
	if !inv.AddItem(item1) {
		t.Error("should be able to add item to inventory")
	}

	// Get item
	retrieved := inv.GetItem("item1")
	if retrieved == nil {
		t.Error("should be able to get item by ID")
	}
	if retrieved != item1 {
		t.Error("retrieved item should match added item")
	}

	// Remove item
	removed := inv.RemoveItem("item1")
	if removed == nil {
		t.Error("should be able to remove item")
	}
	if inv.GetItem("item1") != nil {
		t.Error("item should be removed from inventory")
	}
}

func TestInventoryStacking(t *testing.T) {
	inv := NewInventory(5)

	// Add stackable items with SAME ID (stacking is based on ID, not TemplateID)
	item1 := NewItem("malloc", "same_id", types.Position{})
	item1.Quantity = 3
	inv.AddItem(item1)

	item2 := NewItem("malloc", "same_id", types.Position{})
	item2.Quantity = 2
	inv.AddItem(item2)

	// Should stack into one slot since IDs match
	if len(inv.Items) != 1 {
		t.Errorf("stackable items with same ID should combine: expected 1 slot, got %d", len(inv.Items))
	}
	if inv.Items[0].Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", inv.Items[0].Quantity)
	}
}

func TestInventoryNonStacking(t *testing.T) {
	inv := NewInventory(5)

	// Add stackable items with DIFFERENT instance IDs but SAME TemplateID - they SHOULD stack
	item1 := NewItem("malloc", "id1", types.Position{})
	item1.Quantity = 3
	inv.AddItem(item1)

	item2 := NewItem("malloc", "id2", types.Position{})
	item2.Quantity = 2
	inv.AddItem(item2)

	// Should stack since TemplateIDs match (both are "malloc")
	if len(inv.Items) != 1 {
		t.Errorf("items with same TemplateID should stack: expected 1 slot, got %d", len(inv.Items))
	}
	if inv.Items[0].Quantity != 5 {
		t.Errorf("stacked quantity should be 5, got %d", inv.Items[0].Quantity)
	}
}

func TestInventoryFull(t *testing.T) {
	inv := NewInventory(2)

	// Fill inventory with non-stackable items
	weapon1 := NewItem("basic_script", "weapon1", types.Position{})
	weapon2 := NewItem("vim_blade", "weapon2", types.Position{})
	weapon3 := NewItem("pipe_wrench", "weapon3", types.Position{})

	if !inv.AddItem(weapon1) {
		t.Error("should add first item")
	}
	if !inv.AddItem(weapon2) {
		t.Error("should add second item")
	}
	if inv.AddItem(weapon3) {
		t.Error("should not add third item when inventory is full")
	}
}

func TestInventoryClear(t *testing.T) {
	inv := NewInventory(5)

	inv.AddItem(NewItem("malloc", "item1", types.Position{}))
	inv.AddItem(NewItem("realloc", "item2", types.Position{}))

	if len(inv.Items) != 2 {
		t.Error("inventory should have 2 items")
	}

	inv.Clear()

	if len(inv.Items) != 0 {
		t.Errorf("cleared inventory should be empty, got %d items", len(inv.Items))
	}
}

func TestEquipment(t *testing.T) {
	eq := NewEquipment()

	// All slots should be empty
	if eq.Weapon != nil || eq.Armor != nil || eq.Utility1 != nil || eq.Utility2 != nil {
		t.Error("new equipment should have empty slots")
	}

	// Equip weapon
	weapon := NewItem("vim_blade", "test_weapon", types.Position{})
	old := eq.Equip(weapon)
	if old != nil {
		t.Error("should return nil when equipping to empty slot")
	}
	if eq.Weapon != weapon {
		t.Error("weapon should be equipped")
	}

	// Replace weapon
	weapon2 := NewItem("basic_script", "test_weapon2", types.Position{})
	old = eq.Equip(weapon2)
	if old != weapon {
		t.Error("should return old weapon when replacing")
	}
	if eq.Weapon != weapon2 {
		t.Error("new weapon should be equipped")
	}
}

func TestEquipmentSlots(t *testing.T) {
	eq := NewEquipment()

	weapon := NewItem("vim_blade", "weapon", types.Position{})
	armor := NewItem("firewall", "armor", types.Position{})
	util := NewItem("ssh_key", "util1", types.Position{})

	eq.Equip(weapon)
	eq.Equip(armor)
	eq.Equip(util)

	if eq.Weapon == nil {
		t.Error("weapon should be equipped")
	}
	if eq.Armor == nil {
		t.Error("armor should be equipped")
	}
	if eq.Utility1 == nil {
		t.Error("utility1 should be equipped")
	}
}

func TestEquipmentUtilityAutoFill(t *testing.T) {
	eq := NewEquipment()

	// Both utilities have SlotUtility1, but second should auto-fill Utility2
	util1 := NewItem("ssh_key", "util1", types.Position{})
	util2 := NewItem("env_vars", "util2", types.Position{})

	eq.Equip(util1)
	if eq.Utility1 != util1 {
		t.Error("first utility should be in slot 1")
	}

	// Second utility should auto-fill Utility2 instead of replacing
	old := eq.Equip(util2)
	if old != nil {
		t.Error("should not replace anything when auto-filling slot 2")
	}
	if eq.Utility1 != util1 {
		t.Error("utility1 should still have first item")
	}
	if eq.Utility2 != util2 {
		t.Error("utility2 should have second item (auto-filled)")
	}
}

func TestEquipmentUtilityReplaceWhenBothFull(t *testing.T) {
	eq := NewEquipment()

	util1 := NewItem("ssh_key", "util1", types.Position{})
	util2 := NewItem("env_vars", "util2", types.Position{})
	util3 := NewItem("gpg_ring", "util3", types.Position{})

	eq.Equip(util1) // Goes to Utility1
	eq.Equip(util2) // Goes to Utility2 (auto-fill)

	// Third should replace Utility1 (since they all have SlotUtility1)
	old := eq.Equip(util3)
	if old != util1 {
		t.Errorf("should return replaced item from slot 1, got %v", old)
	}
	if eq.Utility1 != util3 {
		t.Error("utility1 should now have third item")
	}
	if eq.Utility2 != util2 {
		t.Error("utility2 should still have second item")
	}
}

func TestEquipmentUnequip(t *testing.T) {
	eq := NewEquipment()

	weapon := NewItem("vim_blade", "weapon", types.Position{})
	eq.Equip(weapon)

	unequipped := eq.Unequip(SlotWeapon)
	if unequipped != weapon {
		t.Error("should return unequipped weapon")
	}
	if eq.Weapon != nil {
		t.Error("weapon slot should be empty after unequip")
	}

	// Unequip empty slot
	unequipped = eq.Unequip(SlotArmor)
	if unequipped != nil {
		t.Error("unequipping empty slot should return nil")
	}
}

func TestEquipmentGetAll(t *testing.T) {
	eq := NewEquipment()

	// Empty equipment
	all := eq.GetAll()
	if len(all) != 0 {
		t.Errorf("empty equipment should return 0 items, got %d", len(all))
	}

	// Add items
	eq.Equip(NewItem("vim_blade", "w", types.Position{}))
	eq.Equip(NewItem("firewall", "a", types.Position{}))
	eq.Equip(NewItem("ssh_key", "u1", types.Position{}))
	eq.Equip(NewItem("env_vars", "u2", types.Position{}))

	all = eq.GetAll()
	if len(all) != 4 {
		t.Errorf("should have 4 equipped items, got %d", len(all))
	}
}
