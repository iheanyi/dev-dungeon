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
	baseStats := types.Stats{
		PID:  100,
		CPU:  10,
		MEM:  50,
		NICE: 10,
		UID:  0,
	}
	maxStats := types.MaxStats{
		MaxPID: 100,
		MaxMEM: 50,
	}

	switch class {
	case ClassCron:
		baseStats.NICE = 5 // Faster
		baseStats.CPU = 8
	case ClassBash:
		baseStats.CPU = 15 // More attack
		baseStats.MEM = 40
	case ClassVim:
		baseStats.MEM = 80 // More ability capacity
		maxStats.MaxMEM = 80
		baseStats.CPU = 8
	case ClassSudo:
		baseStats.UID = 1 // Higher permissions
		baseStats.PID = 80
		maxStats.MaxPID = 80
	}

	return baseStats, maxStats
}

// getClassSkills returns starting skills for a class.
func getClassSkills(class PlayerClass) []Skill {
	// All classes start with basic attack
	skills := []Skill{
		{ID: "attack", Name: "Attack", MEMCost: 0, BaseDamage: 10},
	}

	switch class {
	case ClassInit:
		skills = append(skills, Skill{
			ID: "fork", Name: "Fork", MEMCost: 20, BaseDamage: 15,
			Description: "Create a child process to attack",
		})
	case ClassCron:
		skills = append(skills, Skill{
			ID: "schedule", Name: "Schedule", MEMCost: 15, BaseDamage: 0,
			Description: "Schedule damage for next turn (2x)",
		})
	case ClassBash:
		skills = append(skills, Skill{
			ID: "pipe", Name: "Pipe", MEMCost: 25, BaseDamage: 20,
			Description: "Chain attacks together",
		})
	case ClassVim:
		skills = append(skills, Skill{
			ID: "macro", Name: "Macro", MEMCost: 30, BaseDamage: 25,
			Description: "Record and replay attacks",
		})
	case ClassSudo:
		skills = append(skills, Skill{
			ID: "escalate", Name: "Escalate", MEMCost: 35, BaseDamage: 30,
			Description: "Bypass enemy defenses",
		})
	}

	return skills
}

// Skill represents a player ability.
type Skill struct {
	ID          string
	Name        string
	Description string
	MEMCost     int
	BaseDamage  int
	Cooldown    int
	CurrentCD   int
}

// GetStats returns the player's current stats.
func (p *Player) GetStats() types.Stats {
	return p.Stats
}

// TakeDamage reduces the player's PID.
func (p *Player) TakeDamage(amount int) bool {
	p.Stats.PID -= amount
	if p.Stats.PID <= 0 {
		p.Stats.PID = 0
		return true // Dead
	}
	return false
}

// Heal restores the player's PID.
func (p *Player) Heal(amount int) {
	p.Stats.PID += amount
	if p.Stats.PID > p.MaxStats.MaxPID {
		p.Stats.PID = p.MaxStats.MaxPID
	}
}

// UseMEM consumes MEM for abilities.
func (p *Player) UseMEM(amount int) bool {
	if p.Stats.MEM < amount {
		return false
	}
	p.Stats.MEM -= amount
	return true
}

// RestoreMEM restores the player's MEM.
func (p *Player) RestoreMEM(amount int) {
	p.Stats.MEM += amount
	if p.Stats.MEM > p.MaxStats.MaxMEM {
		p.Stats.MEM = p.MaxStats.MaxMEM
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
	p.MaxStats.MaxPID += 10
	p.MaxStats.MaxMEM += 5
	p.Stats.PID = p.MaxStats.MaxPID // Full heal on level up
	p.Stats.MEM = p.MaxStats.MaxMEM
	p.Stats.CPU += 2
}

// IsAlive returns whether the player is alive.
func (p *Player) IsAlive() bool {
	return p.Stats.PID > 0
}
