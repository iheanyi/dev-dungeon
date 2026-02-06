package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// openShop initializes and opens the in-game shop.
func (m *Model) openShop() {
	m.shopCursor = 0
	m.shopItems = m.generateShopInventory()
	m.prevView = ViewGame
	m.currentView = ViewShop
}

// generateShopInventory creates the shop's item list based on current depth.
func (m *Model) generateShopInventory() []ShopItem {
	depth := 1
	if m.engine != nil {
		depth = m.engine.CurrentDepth()
	}

	// Base items always available
	items := []ShopItem{
		{TemplateID: "malloc", Name: "malloc()", Price: 10, InStock: true},
		{TemplateID: "fd_restore", Name: "close()", Price: 10, InStock: true},
		{TemplateID: "realloc", Name: "realloc()", Price: 25, InStock: true},
	}

	// Add gear based on depth
	if depth >= 2 {
		items = append(items, ShopItem{TemplateID: "basic_script", Name: "bash script", Price: 30, InStock: true})
		items = append(items, ShopItem{TemplateID: "basic_shell", Name: "/bin/sh", Price: 30, InStock: true})
		items = append(items, ShopItem{TemplateID: "env_vars", Name: "$PATH", Price: 20, InStock: true})
	}
	if depth >= 3 {
		items = append(items, ShopItem{TemplateID: "pipe_wrench", Name: "pipe |", Price: 50, InStock: true})
		items = append(items, ShopItem{TemplateID: "firewall", Name: "iptables", Price: 60, InStock: true})
		items = append(items, ShopItem{TemplateID: "ssh_key", Name: "id_rsa", Price: 40, InStock: true})
	}
	if depth >= 5 {
		items = append(items, ShopItem{TemplateID: "vim_blade", Name: ":wq!", Price: 100, InStock: true})
		items = append(items, ShopItem{TemplateID: "selinux_shield", Name: "SELinux", Price: 120, InStock: true})
		items = append(items, ShopItem{TemplateID: "sudo_potion", Name: "sudo potion", Price: 80, InStock: true})
	}
	if depth >= 7 {
		items = append(items, ShopItem{TemplateID: "kill_9", Name: "kill -9", Price: 150, InStock: true})
		items = append(items, ShopItem{TemplateID: "mmap", Name: "mmap()", Price: 100, InStock: true})
	}

	return items
}

// updateShop handles shop input.
func (m *Model) updateShop(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "$", "q":
		m.currentView = m.prevView
	case "up", "k", "w":
		m.shopCursor--
		if m.shopCursor < 0 {
			m.shopCursor = len(m.shopItems) - 1
		}
	case "down", "j", "s":
		m.shopCursor++
		if m.shopCursor >= len(m.shopItems) {
			m.shopCursor = 0
		}
	case "enter", " ":
		m.buyItem()
	}
	return m, nil
}

// buyItem attempts to purchase the selected shop item.
func (m *Model) buyItem() {
	if m.player == nil || m.shopCursor >= len(m.shopItems) {
		return
	}

	item := &m.shopItems[m.shopCursor]
	if !item.InStock {
		m.statusMsg = "Item out of stock!"
		return
	}

	if m.player.ExitCodes < item.Price {
		m.statusMsg = fmt.Sprintf("Not enough exit codes! Need %d, have %d.", item.Price, m.player.ExitCodes)
		return
	}

	// Create the item
	newItem := entity.NewItem(item.TemplateID, fmt.Sprintf("shop_%s_%d", item.TemplateID, m.player.ExitCodes), types.Position{})
	if newItem == nil {
		m.statusMsg = "Error creating item."
		return
	}

	// Try to add to inventory
	if !m.player.Inventory.AddItem(newItem) {
		m.statusMsg = "Inventory full!"
		return
	}

	// Deduct cost
	m.player.ExitCodes -= item.Price
	item.InStock = false // Sold out
	m.statusMsg = fmt.Sprintf("Purchased %s for %d exit codes!", item.Name, item.Price)
}

// --- Class Unlock Functions ---

// isClassUnlocked checks if a class is unlocked in meta-progress.
func (m *Model) isClassUnlocked(class entity.PlayerClass) bool {
	// init is always unlocked
	if class == entity.ClassInit {
		return true
	}
	if m.metaProgress == nil {
		return false
	}
	for _, unlocked := range m.metaProgress.UnlockedClasses {
		if unlocked == string(class) {
			return true
		}
	}
	return false
}

// getClassUnlockPrice returns the exit code price to unlock a class.
func (m *Model) getClassUnlockPrice(class entity.PlayerClass) int {
	switch class {
	case entity.ClassCron:
		return 50
	case entity.ClassBash:
		return 100
	case entity.ClassVim:
		return 200
	case entity.ClassSudo:
		return 500
	default:
		return 0
	}
}

// --- Unlock Shop Functions ---

// openUnlockShop initializes and opens the unlock shop.
func (m *Model) openUnlockShop() {
	m.unlockCursor = 0
	m.unlockCategory = 0
	m.prevView = ViewMainMenu
	m.currentView = ViewUnlockShop
}

// updateUnlockShop handles unlock shop input.
func (m *Model) updateUnlockShop(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Get the current list based on category
	itemCount := m.getUnlockCategoryItemCount()

	switch msg.String() {
	case "esc", "q":
		m.currentView = ViewMainMenu
	case "up", "k", "w":
		m.unlockCursor--
		if m.unlockCursor < 0 {
			m.unlockCursor = itemCount - 1
			if m.unlockCursor < 0 {
				m.unlockCursor = 0
			}
		}
	case "down", "j", "s":
		m.unlockCursor++
		if m.unlockCursor >= itemCount {
			m.unlockCursor = 0
		}
	case "left", "h":
		m.unlockCategory--
		if m.unlockCategory < 0 {
			m.unlockCategory = 2
		}
		m.unlockCursor = 0
	case "right", "l", "tab":
		m.unlockCategory++
		if m.unlockCategory > 2 {
			m.unlockCategory = 0
		}
		m.unlockCursor = 0
	case "enter", " ":
		m.purchaseUnlock()
	}
	return m, nil
}

// getUnlockCategoryItemCount returns the number of items in the current unlock category.
func (m *Model) getUnlockCategoryItemCount() int {
	switch m.unlockCategory {
	case 0: // Classes
		return 4 // cron, bash, vim, sudo (init is always unlocked)
	case 1: // Bonuses
		return 4 // RAM, CPU, MEM, NICE
	case 2: // Items
		return len(m.getUnlockableItems())
	default:
		return 0
	}
}

// purchaseUnlock attempts to purchase the selected unlock.
func (m *Model) purchaseUnlock() {
	if m.metaProgress == nil {
		m.statusMsg = "Error: No meta progress available"
		return
	}

	switch m.unlockCategory {
	case 0: // Classes
		m.purchaseClassUnlock()
	case 1: // Bonuses
		m.purchaseBonusUnlock()
	case 2: // Items
		m.purchaseItemUnlock()
	}
}

// purchaseClassUnlock handles class unlock purchases.
func (m *Model) purchaseClassUnlock() {
	classes := []entity.PlayerClass{entity.ClassCron, entity.ClassBash, entity.ClassVim, entity.ClassSudo}
	if m.unlockCursor >= len(classes) {
		return
	}

	class := classes[m.unlockCursor]
	if m.isClassUnlocked(class) {
		m.statusMsg = fmt.Sprintf("Class '%s' is already unlocked!", class)
		return
	}

	price := m.getClassUnlockPrice(class)
	if m.metaProgress.TotalExitCodes < price {
		m.statusMsg = fmt.Sprintf("Not enough exit codes! Need %d, have %d.", price, m.metaProgress.TotalExitCodes)
		return
	}

	// Purchase successful
	m.metaProgress.TotalExitCodes -= price
	m.metaProgress.UnlockedClasses = append(m.metaProgress.UnlockedClasses, string(class))

	// Save meta progress
	if m.saveManager != nil {
		_ = m.saveManager.SaveMetaProgress(m.metaProgress) // Best-effort save
	}

	m.statusMsg = fmt.Sprintf("Unlocked class '%s'!", class)
}

// purchaseBonusUnlock handles permanent bonus purchases.
func (m *Model) purchaseBonusUnlock() {
	bonuses := m.getUnlockableBonuses()
	if m.unlockCursor >= len(bonuses) {
		return
	}

	bonus := bonuses[m.unlockCursor]
	if bonus.CurrentLevel >= bonus.MaxLevel {
		m.statusMsg = fmt.Sprintf("%s is already at max level!", bonus.Name)
		return
	}

	price := m.getBonusPrice(bonus)
	if m.metaProgress.TotalExitCodes < price {
		m.statusMsg = fmt.Sprintf("Not enough exit codes! Need %d, have %d.", price, m.metaProgress.TotalExitCodes)
		return
	}

	// Purchase successful
	m.metaProgress.TotalExitCodes -= price

	// Apply the bonus
	switch bonus.StatType {
	case "RAM":
		m.metaProgress.PermanentBonuses.RAM += bonus.BonusAmount
	case "CPU":
		m.metaProgress.PermanentBonuses.CPU += bonus.BonusAmount
	case "MEM":
		m.metaProgress.PermanentBonuses.FD += bonus.BonusAmount
	case "NICE":
		m.metaProgress.PermanentBonuses.NICE += bonus.BonusAmount
	}

	// Save meta progress
	if m.saveManager != nil {
		_ = m.saveManager.SaveMetaProgress(m.metaProgress) // Best-effort save
	}

	m.statusMsg = fmt.Sprintf("Purchased %s upgrade!", bonus.Name)
}

// purchaseItemUnlock handles loot pool item unlock purchases.
func (m *Model) purchaseItemUnlock() {
	items := m.getUnlockableItems()
	if m.unlockCursor >= len(items) {
		return
	}

	item := items[m.unlockCursor]
	if item.Unlocked {
		m.statusMsg = fmt.Sprintf("'%s' is already unlocked!", item.Name)
		return
	}

	if m.metaProgress.TotalExitCodes < item.Price {
		m.statusMsg = fmt.Sprintf("Not enough exit codes! Need %d, have %d.", item.Price, m.metaProgress.TotalExitCodes)
		return
	}

	// Purchase successful
	m.metaProgress.TotalExitCodes -= item.Price
	m.metaProgress.UnlockedItems = append(m.metaProgress.UnlockedItems, item.TemplateID)

	// Save meta progress
	if m.saveManager != nil {
		_ = m.saveManager.SaveMetaProgress(m.metaProgress) // Best-effort save
	}

	m.statusMsg = fmt.Sprintf("Unlocked '%s'!", item.Name)
}

// getUnlockableBonuses returns the list of permanent stat bonuses available.
func (m *Model) getUnlockableBonuses() []UnlockableBonus {
	// Calculate current levels from meta progress
	ramLevel := m.metaProgress.PermanentBonuses.RAM / 5
	cpuLevel := m.metaProgress.PermanentBonuses.CPU / 2
	memLevel := m.metaProgress.PermanentBonuses.FD / 2
	niceLevel := m.metaProgress.PermanentBonuses.NICE / 1

	return []UnlockableBonus{
		{ID: "ram", Name: "+5 Base RAM", Description: "Increases starting health", StatType: "RAM", BonusAmount: 5, CurrentLevel: ramLevel, MaxLevel: 10, BasePrice: 25},
		{ID: "cpu", Name: "+2 Base CPU", Description: "Increases starting attack power", StatType: "CPU", BonusAmount: 2, CurrentLevel: cpuLevel, MaxLevel: 10, BasePrice: 50},
		{ID: "mem", Name: "+2 Base FD", Description: "Increases starting ability resource", StatType: "MEM", BonusAmount: 2, CurrentLevel: memLevel, MaxLevel: 10, BasePrice: 30},
		{ID: "nice", Name: "-1 Base NICE", Description: "Faster starting speed", StatType: "NICE", BonusAmount: 1, CurrentLevel: niceLevel, MaxLevel: 5, BasePrice: 75},
	}
}

// getBonusPrice calculates the price for the next level of a bonus.
func (m *Model) getBonusPrice(bonus UnlockableBonus) int {
	// Price doubles each level
	multiplier := 1
	for i := 0; i < bonus.CurrentLevel; i++ {
		multiplier *= 2
	}
	return bonus.BasePrice * multiplier
}

// getUnlockableItems returns the list of items available to unlock for the loot pool.
func (m *Model) getUnlockableItems() []UnlockableItem {
	// Define items that can be unlocked to add to the loot pool
	allItems := []UnlockableItem{
		{TemplateID: "rm_rf", Name: "rm -rf", Description: "Legendary weapon - devastating attack", Price: 200},
		{TemplateID: "sudo_armor", Name: "sudo armor", Description: "Legendary armor - root protection", Price: 200},
		{TemplateID: "fork_bomb", Name: "fork bomb", Description: "Legendary consumable - massive damage", Price: 150},
		{TemplateID: "kill_9", Name: "kill -9", Description: "Rare weapon - guaranteed process termination", Price: 100},
		{TemplateID: "mmap", Name: "mmap()", Description: "Rare consumable - large memory allocation", Price: 75},
		// Note: root_shard intentionally excluded from unlock shop for multiplayer fairness
		// It still spawns naturally in local single-player runs
	}

	// Mark items as unlocked based on meta progress
	for i := range allItems {
		for _, unlocked := range m.metaProgress.UnlockedItems {
			if unlocked == allItems[i].TemplateID {
				allItems[i].Unlocked = true
				break
			}
		}
	}

	return allItems
}
