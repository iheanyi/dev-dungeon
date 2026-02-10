package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// updateInventory handles inventory view input.
func (m *Model) updateInventory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.player == nil {
		return m, nil
	}

	key := normalizeKey(msg.String())

	// Total items = inventory + 4 equipment slots
	invLen := len(m.player.Inventory.Items)
	totalItems := invLen + 4 // weapon, armor, util1, util2
	if totalItems == 4 {
		totalItems = 4 // At minimum, show equipment slots
	}

	switch {
	case key == "esc" || m.isInventoryKey(key):
		if m.gameState == types.StateCombat {
			m.currentView = ViewCombat
		} else {
			m.currentView = ViewGame
		}
	case m.isMoveUpKey(key):
		m.invCursor--
		if m.invCursor < 0 {
			m.invCursor = totalItems - 1
		}
	case m.isMoveDownKey(key):
		m.invCursor++
		if m.invCursor >= totalItems {
			m.invCursor = 0
		}
	case key == "enter" || key == " ":
		m.useOrEquipItem()
	case key == "e":
		m.equipItem()
	case key == "d":
		m.dropItem()
	case key == "u":
		m.unequipItem()
	}
	return m, nil
}

// useOrEquipItem uses consumables or equips equipment.
func (m *Model) useOrEquipItem() {
	if m.player == nil || m.invCursor >= len(m.player.Inventory.Items) {
		return
	}

	item := m.player.Inventory.Items[m.invCursor]

	switch item.ItemType {
	case entity.ItemTypeConsumable:
		m.useConsumable(item)
	case entity.ItemTypeWeapon, entity.ItemTypeArmor, entity.ItemTypeUtility:
		m.equipItem()
	default:
		m.statusMsg = fmt.Sprintf("Cannot use %s.", item.Name())
	}
}

// useConsumable applies a consumable item's effects.
func (m *Model) useConsumable(item *entity.Item) {
	if item == nil {
		return
	}

	var effectMsg string
	for _, effect := range item.Effects {
		switch effect.Type {
		case entity.EffectHeal:
			oldRAM := m.player.Stats.RAM
			m.player.Heal(effect.Value)
			healed := m.player.Stats.RAM - oldRAM
			effectMsg = fmt.Sprintf("Allocated %d RAM.", healed)

		case entity.EffectRestoreFD:
			oldFD := m.player.Stats.FD
			m.player.RestoreFD(effect.Value)
			restored := m.player.Stats.FD - oldFD
			effectMsg = fmt.Sprintf("Restored %d FD.", restored)

		case entity.EffectDamage:
			// In combat, damage first enemy
			if m.combat != nil && len(m.combat.Enemies) > 0 {
				for _, enemy := range m.combat.Enemies {
					if enemy.IsAlive() {
						killed := enemy.TakeDamage(effect.Value)
						effectMsg = fmt.Sprintf("Dealt %d damage to %s!", effect.Value, enemy.Name())
						if killed {
							effectMsg += " OOM killed!"
						}
						break
					}
				}
			} else {
				// No target for damage item - don't consume
				return
			}

		case entity.EffectBuff:
			// Determine buff type based on item
			var buffType entity.BuffType
			var buffValue int
			var duration int

			switch item.TemplateID {
			case "sudo_potion":
				buffType = entity.BuffInvincible
				buffValue = 0
				duration = 3 // 3 turns of invincibility
				effectMsg = "ROOT ACCESS GRANTED! Immune to damage for 3 turns."
			case "nice_boost":
				buffType = entity.BuffHaste
				buffValue = 5 // -5 NICE
				duration = 5
				effectMsg = "NICE reduced! Acting faster for 5 turns."
			case "cpu_boost":
				buffType = entity.BuffStrength
				buffValue = 10 // +10 CPU
				duration = 5
				effectMsg = "CPU boosted! +10 attack for 5 turns."
			default:
				effectMsg = fmt.Sprintf("Gained buff from %s.", item.Name())
				buffType = entity.BuffStrength
				buffValue = 5
				duration = 3
			}

			m.player.AddBuff(entity.Buff{
				Type:     buffType,
				Name:     item.Name(),
				Duration: duration,
				Value:    buffValue,
			})

		case entity.EffectReveal:
			effectMsg = "Revealed floor contents."
			// TODO: Implement reveal

		default:
			effectMsg = fmt.Sprintf("Used %s.", item.Name())
		}
	}

	// Consume the item
	if item.Stackable && item.Quantity > 1 {
		item.Quantity--
	} else {
		m.player.Inventory.RemoveItem(item.ID())
		// Adjust cursor if needed
		if m.invCursor >= len(m.player.Inventory.Items) && m.invCursor > 0 {
			m.invCursor--
		}
	}

	m.addToHistory(effectMsg)

	// If in combat, add to combat log
	if m.combat != nil {
		m.combatLog = append(m.combatLog, effectMsg)
	}
}

// addToHistory adds a message to the status and history.
func (m *Model) addToHistory(msg string) {
	m.statusMsg = msg
	if msg != "" {
		m.messageHistory = append(m.messageHistory, msg)
		// Cap history at 500 messages
		if len(m.messageHistory) > 500 {
			m.messageHistory = m.messageHistory[1:]
		}
	}
}

// syncEngineMessages syncs messages from the engine into history.
func (m *Model) syncEngineMessages() {
	if m.engine == nil {
		return
	}
	messages := m.engine.Messages()
	for _, msg := range messages {
		m.addToHistory(msg)
	}
	m.engine.ClearMessages()
}

// equipItem equips the selected item.
func (m *Model) equipItem() {
	if m.player == nil || m.invCursor >= len(m.player.Inventory.Items) {
		return
	}

	item := m.player.Inventory.Items[m.invCursor]

	// Only equip weapons, armor, utility
	if item.EquipSlot == entity.SlotNone {
		m.statusMsg = fmt.Sprintf("Cannot equip %s.", item.Name())
		return
	}

	// Remove from inventory
	m.player.Inventory.RemoveItem(item.ID())

	// Equip (returns old item if any)
	oldItem := m.player.Equipment.Equip(item)

	// Put old item back in inventory
	if oldItem != nil {
		m.player.Inventory.AddItem(oldItem)
	}

	// Adjust cursor
	if m.invCursor >= len(m.player.Inventory.Items) && m.invCursor > 0 {
		m.invCursor--
	}

	m.statusMsg = fmt.Sprintf("Equipped %s.", item.Name())
}

// dropItem drops the selected item.
func (m *Model) dropItem() {
	if m.player == nil || m.invCursor >= len(m.player.Inventory.Items) {
		return
	}

	item := m.player.Inventory.Items[m.invCursor]
	m.player.Inventory.RemoveItem(item.ID())

	// Place item at player's position in the world
	if m.engine != nil {
		item.SetPosition(m.player.Position())
		m.engine.GetWorld().AddItem(item)
	}

	// Adjust cursor
	if m.invCursor >= len(m.player.Inventory.Items) && m.invCursor > 0 {
		m.invCursor--
	}

	m.statusMsg = fmt.Sprintf("Dropped %s.", item.Name())
}

// unequipItem removes equipped item and puts it back in inventory.
func (m *Model) unequipItem() {
	if m.player == nil {
		return
	}

	// Check if cursor is in equipment section (after inventory items)
	invLen := len(m.player.Inventory.Items)

	// Equipment slots: 0=weapon, 1=armor, 2=utility1, 3=utility2 (relative to after inventory)
	equipIdx := m.invCursor - invLen
	if equipIdx < 0 {
		m.statusMsg = "Select an equipped item to unequip."
		return
	}

	var item *entity.Item
	var slotName string

	switch equipIdx {
	case 0:
		item = m.player.Equipment.Weapon
		slotName = "weapon"
		if item != nil {
			m.player.Equipment.Weapon = nil
		}
	case 1:
		item = m.player.Equipment.Armor
		slotName = "armor"
		if item != nil {
			m.player.Equipment.Armor = nil
		}
	case 2:
		item = m.player.Equipment.Utility1
		slotName = "utility 1"
		if item != nil {
			m.player.Equipment.Utility1 = nil
		}
	case 3:
		item = m.player.Equipment.Utility2
		slotName = "utility 2"
		if item != nil {
			m.player.Equipment.Utility2 = nil
		}
	default:
		m.statusMsg = "Invalid equipment slot."
		return
	}

	if item == nil {
		m.statusMsg = fmt.Sprintf("No %s equipped.", slotName)
		return
	}

	// Add back to inventory
	if m.player.Inventory.AddItem(item) {
		m.statusMsg = fmt.Sprintf("Unequipped %s.", item.Name())
	} else {
		// Inventory full, re-equip
		switch equipIdx {
		case 0:
			m.player.Equipment.Weapon = item
		case 1:
			m.player.Equipment.Armor = item
		case 2:
			m.player.Equipment.Utility1 = item
		case 3:
			m.player.Equipment.Utility2 = item
		}
		m.statusMsg = "Inventory full, cannot unequip."
	}
}
