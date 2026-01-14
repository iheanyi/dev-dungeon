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
	BehaviorDefensive                       // Heals when low
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
			RAM:  30,  // Low health, easy to kill
			CPU:  5,   // Low damage
			FD:   4,   // Few abilities
			NICE: 15,  // Slow
			UID:  1000,
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
			RAM:  50,
			CPU:  8,
			FD:   8,
			NICE: 10,
			UID:  1, // High privilege (system daemon)
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
			RAM:  20,  // Weak individually
			CPU:  3,
			FD:   2,   // Uses FDs fast (forking!)
			NICE: 5,   // Fast
			UID:  1000,
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
			RAM:  40,
			CPU:  12,  // High damage (memory corruption)
			FD:   6,
			NICE: 8,
			UID:  1000,
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
			RAM:  80,
			CPU:  15,
			FD:   12,
			NICE: 12,  // Slower but stealthy
			UID:  0,   // Root access!
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
			RAM:  500,  // Boss health
			CPU:  30,   // High damage
			FD:   32,   // Many abilities
			NICE: 5,    // Fast
			UID:  0,    // Kernel-level
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
		MaxStats:  types.MaxStats{MaxRAM: scaledStats.RAM, MaxFD: scaledStats.FD},
		XPReward:  template.XPReward * (1 + floorLevel/2),
		Behavior:  template.Behavior,
		LootTable: template.LootTable,
	}
}

// scaleStats scales enemy stats based on floor level.
func scaleStats(base types.Stats, level int) types.Stats {
	multiplier := 1.0 + float64(level)*0.15
	return types.Stats{
		RAM:  int(float64(base.RAM) * multiplier),
		CPU:  int(float64(base.CPU) * multiplier),
		FD:   int(float64(base.FD) * multiplier),
		NICE: base.NICE, // Speed doesn't scale
		UID:  base.UID,  // Permissions don't scale
	}
}

// GetStats returns the enemy's current stats.
func (e *Enemy) GetStats() types.Stats {
	return e.Stats
}

// TakeDamage reduces the enemy's RAM and returns true if dead (OOM).
func (e *Enemy) TakeDamage(amount int) bool {
	e.Stats.RAM -= amount
	if e.Stats.RAM <= 0 {
		e.Stats.RAM = 0
		return true // OOM killed
	}
	return false
}

// IsAlive returns whether the enemy is alive.
func (e *Enemy) IsAlive() bool {
	return e.Stats.RAM > 0
}

// GetDamage returns the damage this enemy deals.
func (e *Enemy) GetDamage() int {
	return e.Stats.CPU
}
