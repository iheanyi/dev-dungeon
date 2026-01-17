// Package save provides save/load functionality for /dev/dungeon.
package save

import (
	"time"

	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// Version is the current save file format version.
const Version = 1

// SaveData represents a complete game save.
type SaveData struct {
	Version        int          `json:"version"`
	MasterSeed     int64        `json:"master_seed"`
	Timestamp      time.Time    `json:"timestamp"`
	Player         PlayerData   `json:"player"`
	CurrentDepth   int          `json:"current_depth"`
	FloorStates    []FloorState `json:"floor_states"`
	MetaProgress   MetaProgress `json:"meta_progress"`
	RunType        string       `json:"run_type,omitempty"`        // "standard", "daily", or "seeded"
	RunStartTime   time.Time    `json:"run_start_time,omitempty"`  // When this run first started
	ElapsedSeconds int          `json:"elapsed_seconds,omitempty"` // Total play time accumulated across sessions
}

// PlayerData represents saved player state.
type PlayerData struct {
	Class       entity.PlayerClass `json:"class"`
	Stats       types.Stats        `json:"stats"`
	MaxStats    types.MaxStats     `json:"max_stats"`
	Level       int                `json:"level"`
	XP          int                `json:"xp"`
	XPToLevel   int                `json:"xp_to_level"`
	Position    types.Position     `json:"position"`
	Inventory   []ItemData         `json:"inventory"`
	Equipment   EquipmentData      `json:"equipment"`
	Skills      []string           `json:"skills"`       // Deprecated: use SkillStates
	SkillStates []SkillState       `json:"skill_states"` // Skill IDs with cooldown state
	ActiveBuffs []BuffState        `json:"active_buffs"` // Active buff effects
	ExitCodes   int                `json:"exit_codes"`
}

// SkillState represents a saved skill with cooldown state.
type SkillState struct {
	ID        string `json:"id"`
	CurrentCD int    `json:"current_cd"`
}

// BuffState represents a saved active buff.
type BuffState struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Value    int    `json:"value"`
}

// ItemData represents a saved item.
type ItemData struct {
	TemplateID string `json:"template_id"`
	Quantity   int    `json:"quantity"`
}

// EquipmentData represents saved equipment slots.
type EquipmentData struct {
	Weapon   string `json:"weapon,omitempty"` // Template ID
	Armor    string `json:"armor,omitempty"`
	Utility1 string `json:"utility1,omitempty"`
	Utility2 string `json:"utility2,omitempty"`
}

// FloorState represents the delta state for a visited floor.
// We don't save the full floor - it regenerates from seed.
// We only save what changed.
type FloorState struct {
	Depth         int              `json:"depth"`
	ExploredTiles []types.Position `json:"explored_tiles"`
	DeadEnemies   []string         `json:"dead_enemies"` // Enemy IDs
	LootedItems   []string         `json:"looted_items"` // Item IDs
}

// MetaProgress represents permanent unlocks that persist across runs.
type MetaProgress struct {
	TotalExitCodes   int         `json:"total_exit_codes"`
	UnlockedClasses  []string    `json:"unlocked_classes"`
	PermanentBonuses StatBonuses `json:"permanent_bonuses"`
	UnlockedItems    []string    `json:"unlocked_items"` // Items added to loot pool
	RunsCompleted    int         `json:"runs_completed"`
	DeepestFloor     int         `json:"deepest_floor"`
	TotalDeaths      int         `json:"total_deaths"`
}

// StatBonuses represents permanent stat bonuses from meta-progression.
type StatBonuses struct {
	PID  int `json:"pid"`
	CPU  int `json:"cpu"`
	MEM  int `json:"mem"`
	NICE int `json:"nice"`
	UID  int `json:"uid"`
}

// NewMetaProgress creates a fresh meta-progress with defaults.
func NewMetaProgress() MetaProgress {
	return MetaProgress{
		UnlockedClasses: []string{"init"}, // Start with init class
		UnlockedItems:   []string{},
	}
}

// SaveTrigger represents what triggered a save.
type SaveTrigger int

const (
	TriggerFloorTransition SaveTrigger = iota
	TriggerCombatVictory
	TriggerRareItemPickup
	TriggerAutoSave
	TriggerManual
	TriggerQuit
)

func (t SaveTrigger) String() string {
	names := []string{
		"floor_transition",
		"combat_victory",
		"rare_item_pickup",
		"auto_save",
		"manual",
		"quit",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "unknown"
}
