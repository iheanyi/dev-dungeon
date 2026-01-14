package entity

import "github.com/iheanyi/devdungeon/internal/types"

// ItemType represents different categories of items.
type ItemType string

const (
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeUtility    ItemType = "utility"
	ItemTypeSkill      ItemType = "skill" // Man pages
)

// ItemRarity represents item rarity levels.
type ItemRarity int

const (
	RarityCommon ItemRarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

// Item represents an item in the game.
type Item struct {
	*BaseEntity
	ItemType    ItemType
	Rarity      ItemRarity
	Description string
	Stackable   bool
	MaxStack    int
	Quantity    int
	Effects     []ItemEffect
	EquipSlot   EquipSlot // For equipment
	StatBonus   types.Stats
}

// ItemEffect represents an effect an item can have.
type ItemEffect struct {
	Type   EffectType
	Value  int
	Target EffectTarget
}

// EffectType represents types of item effects.
type EffectType string

const (
	EffectHeal        EffectType = "heal"       // Restore RAM
	EffectRestoreFD   EffectType = "restore_fd" // Restore file descriptors
	EffectDamage      EffectType = "damage"
	EffectBuff        EffectType = "buff"
	EffectReveal      EffectType = "reveal"
	EffectInstantKill EffectType = "instant_kill"
)

// EffectTarget represents what an effect targets.
type EffectTarget string

const (
	TargetSelf  EffectTarget = "self"
	TargetEnemy EffectTarget = "enemy"
	TargetAll   EffectTarget = "all"
	TargetFloor EffectTarget = "floor"
)

// EquipSlot represents equipment slots.
type EquipSlot string

const (
	SlotNone     EquipSlot = ""
	SlotWeapon   EquipSlot = "weapon"
	SlotArmor    EquipSlot = "armor"
	SlotUtility1 EquipSlot = "utility1"
	SlotUtility2 EquipSlot = "utility2"
)

// ItemTemplate defines an item type.
type ItemTemplate struct {
	ID          string
	Name        string
	Glyph       rune
	ItemType    ItemType
	Rarity      ItemRarity
	Description string
	Stackable   bool
	MaxStack    int
	Effects     []ItemEffect
	EquipSlot   EquipSlot
	StatBonus   types.Stats
}

// ItemTemplates holds all item definitions.
var ItemTemplates = map[string]ItemTemplate{
	// Consumables
	"sudo_potion": {
		ID:          "sudo_potion",
		Name:        "sudo potion",
		Glyph:       '!',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityRare,
		Description: "Grants temporary invincibility (root mode)",
		Stackable:   true,
		MaxStack:    3,
		Effects: []ItemEffect{
			{Type: EffectBuff, Value: 3, Target: TargetSelf}, // 3 turns invincible
		},
	},
	"grep_scroll": {
		ID:          "grep_scroll",
		Name:        "grep scroll",
		Glyph:       '?',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityUncommon,
		Description: "Reveals all items on the floor",
		Stackable:   true,
		MaxStack:    5,
		Effects: []ItemEffect{
			{Type: EffectReveal, Value: 1, Target: TargetFloor},
		},
	},
	"malloc": {
		ID:          "malloc",
		Name:        "malloc()",
		Glyph:       '+',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityCommon,
		Description: "Allocates 30 RAM",
		Stackable:   true,
		MaxStack:    10,
		Effects: []ItemEffect{
			{Type: EffectHeal, Value: 30, Target: TargetSelf},
		},
	},
	"fd_restore": {
		ID:          "fd_restore",
		Name:        "close()",
		Glyph:       '*',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityCommon,
		Description: "Closes unused FDs, restores 4",
		Stackable:   true,
		MaxStack:    10,
		Effects: []ItemEffect{
			{Type: EffectRestoreFD, Value: 4, Target: TargetSelf},
		},
	},
	// Weapons
	"kill_9": {
		ID:          "kill_9",
		Name:        "kill -9",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityRare,
		Description: "SIGKILL: instant kill on low RAM enemies",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 10},
		Effects: []ItemEffect{
			{Type: EffectInstantKill, Value: 20, Target: TargetEnemy}, // Kills if <20 RAM
		},
	},
	"basic_script": {
		ID:          "basic_script",
		Name:        "bash script",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityCommon,
		Description: "A simple attack script",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 3},
	},
	// Armor
	"chmod_x": {
		ID:          "chmod_x",
		Name:        "chmod +x",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityUncommon,
		Description: "Increases execution permissions and RAM",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 20, UID: -100}, // Lower UID = more power
	},
	"basic_shell": {
		ID:          "basic_shell",
		Name:        "/bin/sh",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityCommon,
		Description: "Basic protective shell",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 10},
	},
	// Loot drops
	"memory_fragment": {
		ID:          "memory_fragment",
		Name:        "RAM fragment",
		Glyph:       '%',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityCommon,
		Description: "Allocates 10 RAM",
		Stackable:   true,
		MaxStack:    20,
		Effects: []ItemEffect{
			{Type: EffectHeal, Value: 10, Target: TargetSelf},
		},
	},
	"service_token": {
		ID:          "service_token",
		Name:        "systemd token",
		Glyph:       '$',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityUncommon,
		Description: "Allocates 20 RAM",
		Stackable:   true,
		MaxStack:    10,
		Effects: []ItemEffect{
			{Type: EffectHeal, Value: 20, Target: TargetSelf},
		},
	},
	"cpu_cycle": {
		ID:          "cpu_cycle",
		Name:        "CPU cycle",
		Glyph:       'o',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityCommon,
		Description: "Temporary +5 CPU for this combat",
		Stackable:   true,
		MaxStack:    10,
		Effects: []ItemEffect{
			{Type: EffectBuff, Value: 5, Target: TargetSelf},
		},
	},
	"core_dump": {
		ID:          "core_dump",
		Name:        "core dump",
		Glyph:       '#',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityUncommon,
		Description: "Throws 40 damage at an enemy",
		Stackable:   true,
		MaxStack:    5,
		Effects: []ItemEffect{
			{Type: EffectDamage, Value: 40, Target: TargetEnemy},
		},
	},
	"root_shard": {
		ID:          "root_shard",
		Name:        "root shard",
		Glyph:       '^',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityRare,
		Description: "Permanently lowers UID by 100",
		Stackable:   false,
		Effects: []ItemEffect{
			{Type: EffectBuff, Value: -100, Target: TargetSelf}, // Lower UID = more power
		},
	},
}

// NewItem creates a new item from a template.
func NewItem(templateID string, id string, pos types.Position) *Item {
	template, ok := ItemTemplates[templateID]
	if !ok {
		return nil
	}

	return &Item{
		BaseEntity: NewBaseEntity(
			id,
			template.Name,
			pos,
			template.Glyph,
			false,
		),
		ItemType:    template.ItemType,
		Rarity:      template.Rarity,
		Description: template.Description,
		Stackable:   template.Stackable,
		MaxStack:    template.MaxStack,
		Quantity:    1,
		Effects:     template.Effects,
		EquipSlot:   template.EquipSlot,
		StatBonus:   template.StatBonus,
	}
}

// Inventory manages the player's items.
type Inventory struct {
	Items    []*Item
	MaxSlots int
}

// NewInventory creates a new inventory.
func NewInventory(maxSlots int) *Inventory {
	return &Inventory{
		Items:    make([]*Item, 0, maxSlots),
		MaxSlots: maxSlots,
	}
}

// AddItem adds an item to the inventory.
func (inv *Inventory) AddItem(item *Item) bool {
	// Try to stack with existing item
	if item.Stackable {
		for _, existing := range inv.Items {
			if existing.ID() == item.ID() && existing.Quantity < existing.MaxStack {
				space := existing.MaxStack - existing.Quantity
				if item.Quantity <= space {
					existing.Quantity += item.Quantity
					return true
				} else {
					existing.Quantity = existing.MaxStack
					item.Quantity -= space
				}
			}
		}
	}

	// Add as new item if space available
	if len(inv.Items) < inv.MaxSlots {
		inv.Items = append(inv.Items, item)
		return true
	}

	return false
}

// RemoveItem removes an item from the inventory.
func (inv *Inventory) RemoveItem(id string) *Item {
	for i, item := range inv.Items {
		if item.ID() == id {
			inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			return item
		}
	}
	return nil
}

// GetItem returns an item by ID.
func (inv *Inventory) GetItem(id string) *Item {
	for _, item := range inv.Items {
		if item.ID() == id {
			return item
		}
	}
	return nil
}

// Equipment manages equipped items.
type Equipment struct {
	Weapon   *Item
	Armor    *Item
	Utility1 *Item
	Utility2 *Item
}

// NewEquipment creates a new equipment set.
func NewEquipment() *Equipment {
	return &Equipment{}
}

// Equip equips an item in the appropriate slot.
func (eq *Equipment) Equip(item *Item) *Item {
	var old *Item
	switch item.EquipSlot {
	case SlotWeapon:
		old = eq.Weapon
		eq.Weapon = item
	case SlotArmor:
		old = eq.Armor
		eq.Armor = item
	case SlotUtility1:
		old = eq.Utility1
		eq.Utility1 = item
	case SlotUtility2:
		old = eq.Utility2
		eq.Utility2 = item
	}
	return old
}

// GetStatBonus returns total stat bonuses from equipment.
func (eq *Equipment) GetStatBonus() types.Stats {
	bonus := types.Stats{}
	if eq.Weapon != nil {
		bonus.CPU += eq.Weapon.StatBonus.CPU
		bonus.RAM += eq.Weapon.StatBonus.RAM
		bonus.FD += eq.Weapon.StatBonus.FD
	}
	if eq.Armor != nil {
		bonus.CPU += eq.Armor.StatBonus.CPU
		bonus.RAM += eq.Armor.StatBonus.RAM
		bonus.FD += eq.Armor.StatBonus.FD
		bonus.UID += eq.Armor.StatBonus.UID
	}
	if eq.Utility1 != nil {
		bonus.CPU += eq.Utility1.StatBonus.CPU
		bonus.RAM += eq.Utility1.StatBonus.RAM
		bonus.FD += eq.Utility1.StatBonus.FD
	}
	if eq.Utility2 != nil {
		bonus.CPU += eq.Utility2.StatBonus.CPU
		bonus.RAM += eq.Utility2.StatBonus.RAM
		bonus.FD += eq.Utility2.StatBonus.FD
	}
	return bonus
}
