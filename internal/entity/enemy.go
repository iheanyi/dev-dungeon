package entity

import "github.com/iheanyi/devdungeon/internal/types"

// EnemyType represents different enemy types.
type EnemyType string

const (
	EnemyZombie      EnemyType = "zombie"
	EnemyDaemon      EnemyType = "daemon"
	EnemyForkBomb    EnemyType = "fork_bomb"
	EnemySegfault    EnemyType = "segfault"
	EnemyRootkit     EnemyType = "rootkit"
	EnemyKernelPanic EnemyType = "kernel_panic"
)

// Enemy represents an enemy entity.
type Enemy struct {
	*BaseEntity
	Type      EnemyType
	Stats     types.Stats
	MaxStats  types.MaxStats
	XPReward  int
	Behavior  EnemyBehavior
	LootTable []string // Item IDs that can drop
}

// EnemyBehavior defines how an enemy acts in combat.
type EnemyBehavior int

const (
	BehaviorAggressive EnemyBehavior = iota // Always attacks
	BehaviorDefensive                       // Heals when low HP
	BehaviorErratic                         // Random actions
	BehaviorSwarm                           // Summons more enemies
	BehaviorStealth                         // Ambush attacks
)

// EnemyTemplate defines an enemy type's base stats.
type EnemyTemplate struct {
	Type      EnemyType
	Name      string
	Glyph     rune
	BaseStats types.Stats
	XPReward  int
	Behavior  EnemyBehavior
	LootTable []string
}

// EnemyTemplates holds all enemy definitions.
var EnemyTemplates = map[EnemyType]EnemyTemplate{
	EnemyZombie: {
		Type:  EnemyZombie,
		Name:  "zombie process",
		Glyph: 'z',
		BaseStats: types.Stats{
			PID:  30,
			CPU:  5,
			MEM:  10,
			NICE: 15, // Slow
			UID:  0,
		},
		XPReward:  10,
		Behavior:  BehaviorAggressive,
		LootTable: []string{"memory_fragment"},
	},
	EnemyDaemon: {
		Type:  EnemyDaemon,
		Name:  "daemon",
		Glyph: 'd',
		BaseStats: types.Stats{
			PID:  50,
			CPU:  8,
			MEM:  20,
			NICE: 10,
			UID:  0,
		},
		XPReward:  20,
		Behavior:  BehaviorDefensive,
		LootTable: []string{"service_token"},
	},
	EnemyForkBomb: {
		Type:  EnemyForkBomb,
		Name:  "fork bomb",
		Glyph: 'f',
		BaseStats: types.Stats{
			PID:  20,
			CPU:  3,
			MEM:  5,
			NICE: 5, // Fast
			UID:  0,
		},
		XPReward:  15,
		Behavior:  BehaviorSwarm,
		LootTable: []string{"cpu_cycle"},
	},
	EnemySegfault: {
		Type:  EnemySegfault,
		Name:  "segfault",
		Glyph: 's',
		BaseStats: types.Stats{
			PID:  40,
			CPU:  12,
			MEM:  15,
			NICE: 8,
			UID:  0,
		},
		XPReward:  25,
		Behavior:  BehaviorErratic,
		LootTable: []string{"core_dump"},
	},
	EnemyRootkit: {
		Type:  EnemyRootkit,
		Name:  "rootkit",
		Glyph: 'r',
		BaseStats: types.Stats{
			PID:  80,
			CPU:  15,
			MEM:  30,
			NICE: 12,
			UID:  1,
		},
		XPReward:  50,
		Behavior:  BehaviorStealth,
		LootTable: []string{"root_shard"},
	},
	EnemyKernelPanic: {
		Type:  EnemyKernelPanic,
		Name:  "KERNEL PANIC",
		Glyph: 'K',
		BaseStats: types.Stats{
			PID:  500,
			CPU:  30,
			MEM:  100,
			NICE: 5,
			UID:  0,
		},
		XPReward:  500,
		Behavior:  BehaviorAggressive,
		LootTable: []string{"victory"},
	},
}

// NewEnemy creates a new enemy from a template.
func NewEnemy(enemyType EnemyType, id string, pos types.Position, floorLevel int) *Enemy {
	template, ok := EnemyTemplates[enemyType]
	if !ok {
		template = EnemyTemplates[EnemyZombie] // Default
	}

	// Scale stats based on floor level
	scaledStats := scaleStats(template.BaseStats, floorLevel)

	return &Enemy{
		BaseEntity: NewBaseEntity(
			id,
			template.Name,
			pos,
			template.Glyph,
			true,
		),
		Type:      template.Type,
		Stats:     scaledStats,
		MaxStats:  types.MaxStats{MaxPID: scaledStats.PID, MaxMEM: scaledStats.MEM},
		XPReward:  template.XPReward * (1 + floorLevel/2),
		Behavior:  template.Behavior,
		LootTable: template.LootTable,
	}
}

// scaleStats scales enemy stats based on floor level.
func scaleStats(base types.Stats, level int) types.Stats {
	multiplier := 1.0 + float64(level)*0.15
	return types.Stats{
		PID:  int(float64(base.PID) * multiplier),
		CPU:  int(float64(base.CPU) * multiplier),
		MEM:  int(float64(base.MEM) * multiplier),
		NICE: base.NICE, // Speed doesn't scale
		UID:  base.UID,
	}
}

// GetStats returns the enemy's current stats.
func (e *Enemy) GetStats() types.Stats {
	return e.Stats
}

// TakeDamage reduces the enemy's PID and returns true if dead.
func (e *Enemy) TakeDamage(amount int) bool {
	e.Stats.PID -= amount
	if e.Stats.PID <= 0 {
		e.Stats.PID = 0
		return true
	}
	return false
}

// IsAlive returns whether the enemy is alive.
func (e *Enemy) IsAlive() bool {
	return e.Stats.PID > 0
}

// GetDamage returns the damage this enemy deals.
func (e *Enemy) GetDamage() int {
	return e.Stats.CPU
}
