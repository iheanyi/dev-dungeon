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

// String returns the string representation of a rarity.
func (r ItemRarity) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityEpic:
		return "Epic"
	case RarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// Item represents an item in the game.
type Item struct {
	*BaseEntity
	TemplateID  string // ID of the template this item was created from
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
			{Type: EffectBuff, Value: 1, Target: TargetSelf}, // 1 turn invincible (nerfed from 3 for multiplayer balance)
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
			{Type: EffectInstantKill, Value: 15, Target: TargetEnemy}, // Kills if <15 RAM (nerfed from 20)
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

	// === MORE WEAPONS ===
	"pipe_wrench": {
		ID:          "pipe_wrench",
		Name:        "pipe |",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityUncommon,
		Description: "Chain your attacks together",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 5, NICE: -1}, // Faster attacks
	},
	"vim_blade": {
		ID:          "vim_blade",
		Name:        ":wq!",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityRare,
		Description: "Force write and quit - devastating",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 12, FD: 4},
	},
	"sed_saber": {
		ID:          "sed_saber",
		Name:        "sed s/.*/dead/",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityUncommon,
		Description: "Stream edit your enemies",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 7},
	},
	"awk_axe": {
		ID:          "awk_axe",
		Name:        "awk '{kill}'",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityUncommon,
		Description: "Pattern match and destroy",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 6, FD: 2},
	},
	"grep_glaive": {
		ID:          "grep_glaive",
		Name:        "grep -r pain",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityCommon,
		Description: "Recursively search and destroy",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 4},
	},
	"fork_bomb": {
		ID:          "fork_bomb",
		Name:        ":(){:|:&};:",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityLegendary,
		Description: "The forbidden weapon",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 15, NICE: -3},
	},
	"rm_rf": {
		ID:          "rm_rf",
		Name:        "rm -rf /",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityLegendary,
		Description: "Delete everything in your path",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 20},
		Effects: []ItemEffect{
			{Type: EffectInstantKill, Value: 20, Target: TargetEnemy}, // Nerfed from 30 for multiplayer balance
		},
	},
	"cron_claw": {
		ID:          "cron_claw",
		Name:        "* * * * *",
		Glyph:       ')',
		ItemType:    ItemTypeWeapon,
		Rarity:      RarityUncommon,
		Description: "Scheduled strikes every turn",
		Stackable:   false,
		EquipSlot:   SlotWeapon,
		StatBonus:   types.Stats{CPU: 4, NICE: -2},
	},

	// === MORE ARMOR ===
	"firewall": {
		ID:          "firewall",
		Name:        "iptables",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityUncommon,
		Description: "Blocks incoming attacks",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 25},
	},
	"selinux_shield": {
		ID:          "selinux_shield",
		Name:        "SELinux",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityRare,
		Description: "Mandatory access control",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 30, UID: -200},
	},
	"sudo_armor": {
		ID:          "sudo_armor",
		Name:        "sudoers.d/",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityLegendary,
		Description: "Root-level protection",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 50, UID: -500},
	},
	"container": {
		ID:          "container",
		Name:        "docker run",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityRare,
		Description: "Isolated namespace protection",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 35, FD: 4},
	},
	"sandbox": {
		ID:          "sandbox",
		Name:        "chroot jail",
		Glyph:       '[',
		ItemType:    ItemTypeArmor,
		Rarity:      RarityUncommon,
		Description: "Restricted environment",
		Stackable:   false,
		EquipSlot:   SlotArmor,
		StatBonus:   types.Stats{RAM: 15, FD: 2},
	},

	// === UTILITY ITEMS ===
	"ssh_key": {
		ID:          "ssh_key",
		Name:        "id_rsa",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityUncommon,
		Description: "SSH private key - grants access",
		Stackable:   false,
		EquipSlot:   SlotUtility1,
		StatBonus:   types.Stats{UID: -200, FD: 2},
	},
	"gpg_ring": {
		ID:          "gpg_ring",
		Name:        "GPG keyring",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityRare,
		Description: "Cryptographic identity",
		Stackable:   false,
		EquipSlot:   SlotUtility1,
		StatBonus:   types.Stats{UID: -300, RAM: 10},
	},
	"env_vars": {
		ID:          "env_vars",
		Name:        "$PATH",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityCommon,
		Description: "Environment variables for efficiency",
		Stackable:   false,
		EquipSlot:   SlotUtility1,
		StatBonus:   types.Stats{NICE: -2, CPU: 2},
	},
	"cron_tab": {
		ID:          "cron_tab",
		Name:        "crontab",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityUncommon,
		Description: "Scheduled task automation",
		Stackable:   false,
		EquipSlot:   SlotUtility2,
		StatBonus:   types.Stats{NICE: -3, FD: 3},
	},
	"alias_file": {
		ID:          "alias_file",
		Name:        ".bashrc",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityCommon,
		Description: "Shell aliases for quick commands",
		Stackable:   false,
		EquipSlot:   SlotUtility2,
		StatBonus:   types.Stats{CPU: 3, NICE: -1},
	},
	"config_file": {
		ID:          "config_file",
		Name:        ".vimrc",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityUncommon,
		Description: "Optimized editor configuration",
		Stackable:   false,
		EquipSlot:   SlotUtility2,
		StatBonus:   types.Stats{FD: 4, CPU: 2},
	},
	"tmux_session": {
		ID:          "tmux_session",
		Name:        "tmux attach",
		Glyph:       '~',
		ItemType:    ItemTypeUtility,
		Rarity:      RarityRare,
		Description: "Persistent terminal multiplexer",
		Stackable:   false,
		EquipSlot:   SlotUtility1,
		StatBonus:   types.Stats{FD: 6, RAM: 15, NICE: -1},
	},

	// === MORE CONSUMABLES ===
	"realloc": {
		ID:          "realloc",
		Name:        "realloc()",
		Glyph:       '+',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityUncommon,
		Description: "Allocates 50 RAM",
		Stackable:   true,
		MaxStack:    5,
		Effects: []ItemEffect{
			{Type: EffectHeal, Value: 50, Target: TargetSelf},
		},
	},
	"mmap": {
		ID:          "mmap",
		Name:        "mmap()",
		Glyph:       '+',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityRare,
		Description: "Memory map: major heal",
		Stackable:   true,
		MaxStack:    3,
		Effects: []ItemEffect{
			{Type: EffectHeal, Value: 80, Target: TargetSelf}, // Nerfed from 999 for multiplayer balance
		},
	},
	"nice_boost": {
		ID:          "nice_boost",
		Name:        "nice -n -20",
		Glyph:       '>',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityUncommon,
		Description: "Highest priority: +5 speed for 5 turns",
		Stackable:   true,
		MaxStack:    5,
		Effects: []ItemEffect{
			{Type: EffectBuff, Value: 5, Target: TargetSelf},
		},
	},
	"cpu_boost": {
		ID:          "cpu_boost",
		Name:        "turbo boost",
		Glyph:       '>',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityUncommon,
		Description: "+10 CPU for 5 turns",
		Stackable:   true,
		MaxStack:    5,
		Effects: []ItemEffect{
			{Type: EffectBuff, Value: 10, Target: TargetSelf},
		},
	},
	"segfault_bomb": {
		ID:          "segfault_bomb",
		Name:        "SIGSEGV",
		Glyph:       '#',
		ItemType:    ItemTypeConsumable,
		Rarity:      RarityRare,
		Description: "Throws 80 damage at an enemy",
		Stackable:   true,
		MaxStack:    3,
		Effects: []ItemEffect{
			{Type: EffectDamage, Value: 80, Target: TargetEnemy},
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
		TemplateID:  templateID,
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
	// Try to stack with existing item (by TemplateID, not instance ID)
	if item.Stackable {
		for _, existing := range inv.Items {
			if existing.TemplateID == item.TemplateID && existing.Quantity < existing.MaxStack {
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

// Clear removes all items from inventory.
func (inv *Inventory) Clear() {
	inv.Items = make([]*Item, 0, inv.MaxSlots)
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
// For utility items, auto-fills Utility2 if Utility1 is occupied.
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
		// Auto-fill: if Utility1 is occupied, try Utility2
		if eq.Utility1 != nil && eq.Utility2 == nil {
			eq.Utility2 = item
			return nil
		}
		old = eq.Utility1
		eq.Utility1 = item
	case SlotUtility2:
		// Auto-fill: if Utility2 is occupied, try Utility1
		if eq.Utility2 != nil && eq.Utility1 == nil {
			eq.Utility1 = item
			return nil
		}
		old = eq.Utility2
		eq.Utility2 = item
	}
	return old
}

// Unequip removes an item from the specified slot and returns it.
func (eq *Equipment) Unequip(slot EquipSlot) *Item {
	var item *Item
	switch slot {
	case SlotWeapon:
		item = eq.Weapon
		eq.Weapon = nil
	case SlotArmor:
		item = eq.Armor
		eq.Armor = nil
	case SlotUtility1:
		item = eq.Utility1
		eq.Utility1 = nil
	case SlotUtility2:
		item = eq.Utility2
		eq.Utility2 = nil
	}
	return item
}

// GetAll returns all equipped items.
func (eq *Equipment) GetAll() []*Item {
	var items []*Item
	if eq.Weapon != nil {
		items = append(items, eq.Weapon)
	}
	if eq.Armor != nil {
		items = append(items, eq.Armor)
	}
	if eq.Utility1 != nil {
		items = append(items, eq.Utility1)
	}
	if eq.Utility2 != nil {
		items = append(items, eq.Utility2)
	}
	return items
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
