package entity

import "github.com/iheanyi/devdungeon/internal/types"

// PlayerClass represents the player's class.
type PlayerClass string

const (
	ClassInit PlayerClass = "init"
	ClassCron PlayerClass = "cron"
	ClassBash PlayerClass = "bash"
	ClassVim  PlayerClass = "vim"
	ClassSudo PlayerClass = "sudo"
)

// Player represents the player character.
type Player struct {
	*BaseEntity
	Class     PlayerClass
	Stats     types.Stats
	MaxStats  types.MaxStats
	Level     int
	XP        int
	XPToLevel int
	Inventory *Inventory
	Equipment *Equipment
	Skills    []Skill
	ExitCodes int // Current run currency
}

// NewPlayer creates a new player with the given class.
func NewPlayer(class PlayerClass) *Player {
	stats, maxStats := getClassStats(class)

	return &Player{
		BaseEntity: NewBaseEntity(
			"player",
			string(class),
			types.Position{X: 0, Y: 0},
			'@',
			true,
		),
		Class:     class,
		Stats:     stats,
		MaxStats:  maxStats,
		Level:     1,
		XP:        0,
		XPToLevel: 100,
		Inventory: NewInventory(10),
		Equipment: NewEquipment(),
		Skills:    getClassSkills(class),
		ExitCodes: 0,
	}
}

// getClassStats returns starting stats for a class.
func getClassStats(class PlayerClass) (types.Stats, types.MaxStats) {
	// Base stats for all classes
	baseStats := types.Stats{
		RAM:  100, // Health (memory allocation)
		CPU:  10,  // Attack power
		FD:   16,  // Ability resource (file descriptors)
		NICE: 10,  // Speed (lower = faster)
		UID:  1000, // Access level (lower = more power, 0 = root)
	}
	maxStats := types.MaxStats{
		MaxRAM: 100,
		MaxFD:  16,
	}

	switch class {
	case ClassCron:
		// Cron: Fast scheduler, good at timing
		baseStats.NICE = 5 // Faster
		baseStats.CPU = 8
	case ClassBash:
		// Bash: Powerful shell, high attack
		baseStats.CPU = 15  // More attack
		baseStats.FD = 12   // Fewer abilities
		maxStats.MaxFD = 12
	case ClassVim:
		// Vim: Complex editor, many abilities
		baseStats.FD = 24   // More ability capacity
		maxStats.MaxFD = 24
		baseStats.CPU = 8
	case ClassSudo:
		// Sudo: Privilege escalation, access-focused
		baseStats.UID = 0    // Root access!
		baseStats.RAM = 80   // Less health (great power = great risk)
		maxStats.MaxRAM = 80
	}

	return baseStats, maxStats
}

// getClassSkills returns starting skills for a class.
func getClassSkills(class PlayerClass) []Skill {
	// All classes start with basic attack
	skills := []Skill{
		{ID: "kill", Name: "kill -TERM", FDCost: 0, BaseDamage: 10,
			Description: "Send SIGTERM to target"},
	}

	switch class {
	case ClassInit:
		skills = append(skills, Skill{
			ID: "fork", Name: "fork()", FDCost: 4, BaseDamage: 15,
			Description: "Create a child process to attack",
		})
	case ClassCron:
		skills = append(skills, Skill{
			ID: "schedule", Name: "crontab", FDCost: 3, BaseDamage: 0,
			Description: "Schedule damage for next turn (2x)",
		})
	case ClassBash:
		skills = append(skills, Skill{
			ID: "pipe", Name: "pipe |", FDCost: 5, BaseDamage: 20,
			Description: "Chain attacks together",
		})
	case ClassVim:
		skills = append(skills, Skill{
			ID: "macro", Name: ":normal", FDCost: 6, BaseDamage: 25,
			Description: "Record and replay attacks",
		})
	case ClassSudo:
		skills = append(skills, Skill{
			ID: "escalate", Name: "sudo !!", FDCost: 8, BaseDamage: 30,
			Description: "Bypass enemy defenses with root",
		})
	}

	return skills
}

// Skill represents a player ability.
type Skill struct {
	ID          string
	Name        string
	Description string
	FDCost      int // File descriptors consumed
	BaseDamage  int
	Cooldown    int
	CurrentCD   int
}

// GetStats returns the player's current stats.
func (p *Player) GetStats() types.Stats {
	return p.Stats
}

// TakeDamage reduces the player's RAM (health).
func (p *Player) TakeDamage(amount int) bool {
	p.Stats.RAM -= amount
	if p.Stats.RAM <= 0 {
		p.Stats.RAM = 0
		return true // OOM killed
	}
	return false
}

// Heal restores the player's RAM.
func (p *Player) Heal(amount int) {
	p.Stats.RAM += amount
	if p.Stats.RAM > p.MaxStats.MaxRAM {
		p.Stats.RAM = p.MaxStats.MaxRAM
	}
}

// UseFD consumes file descriptors for abilities.
func (p *Player) UseFD(amount int) bool {
	if p.Stats.FD < amount {
		return false
	}
	p.Stats.FD -= amount
	return true
}

// RestoreFD restores the player's file descriptors.
func (p *Player) RestoreFD(amount int) {
	p.Stats.FD += amount
	if p.Stats.FD > p.MaxStats.MaxFD {
		p.Stats.FD = p.MaxStats.MaxFD
	}
}

// GainXP adds experience and handles level ups.
func (p *Player) GainXP(amount int) bool {
	p.XP += amount
	if p.XP >= p.XPToLevel {
		p.LevelUp()
		return true
	}
	return false
}

// LevelUp increases the player's level and stats.
func (p *Player) LevelUp() {
	p.Level++
	p.XP -= p.XPToLevel
	p.XPToLevel = int(float64(p.XPToLevel) * 1.5)

	// Stat increases
	p.MaxStats.MaxRAM += 10
	p.MaxStats.MaxFD += 2
	p.Stats.RAM = p.MaxStats.MaxRAM // Full heal on level up
	p.Stats.FD = p.MaxStats.MaxFD
	p.Stats.CPU += 2
}

// IsAlive returns whether the player is alive.
func (p *Player) IsAlive() bool {
	return p.Stats.RAM > 0
}
