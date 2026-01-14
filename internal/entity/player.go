package entity

import (
	"fmt"

	"github.com/iheanyi/devdungeon/internal/types"
)

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
	Class       PlayerClass
	Stats       types.Stats
	MaxStats    types.MaxStats
	Level       int
	XP          int
	XPToLevel   int
	Inventory   *Inventory
	Equipment   *Equipment
	Skills      []Skill
	ExitCodes   int // Current run currency
	ActiveBuffs []Buff
}

// PermanentBonuses represents stat bonuses from meta-progression.
type PermanentBonuses struct {
	RAM  int // Bonus to max RAM
	CPU  int // Bonus to CPU
	FD   int // Bonus to max FD
	NICE int // Bonus to NICE (negative is better)
}

// NewPlayer creates a new player with the given class.
func NewPlayer(class PlayerClass) *Player {
	return NewPlayerWithBonuses(class, PermanentBonuses{})
}

// NewPlayerWithBonuses creates a new player with the given class and permanent bonuses.
func NewPlayerWithBonuses(class PlayerClass, bonuses PermanentBonuses) *Player {
	stats, maxStats := getClassStats(class)

	// Apply permanent bonuses
	stats.RAM += bonuses.RAM
	maxStats.MaxRAM += bonuses.RAM
	stats.CPU += bonuses.CPU
	stats.FD += bonuses.FD
	maxStats.MaxFD += bonuses.FD
	stats.NICE -= bonuses.NICE // Lower NICE is better (faster)
	if stats.NICE < 1 {
		stats.NICE = 1 // Minimum NICE of 1
	}

	p := &Player{
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

	// Give class-specific starting equipment and items
	giveStartingGear(p, class)

	return p
}

// giveStartingGear gives class-specific starting equipment and items.
func giveStartingGear(p *Player, class PlayerClass) {
	// Everyone starts with some healing
	startingMalloc := NewItem("malloc", "start_malloc_1", types.Position{})
	startingMalloc2 := NewItem("malloc", "start_malloc_2", types.Position{})
	if startingMalloc != nil {
		p.Inventory.AddItem(startingMalloc)
	}
	if startingMalloc2 != nil {
		p.Inventory.AddItem(startingMalloc2)
	}

	// Class-specific gear
	switch class {
	case ClassInit:
		// Init: balanced starter, gets basic gear
		weapon := NewItem("basic_script", "init_weapon", types.Position{})
		armor := NewItem("basic_shell", "init_armor", types.Position{})
		if weapon != nil {
			p.Equipment.Equip(weapon)
		}
		if armor != nil {
			p.Equipment.Equip(armor)
		}

	case ClassCron:
		// Cron: scheduler, gets speed-focused gear
		weapon := NewItem("cron_claw", "cron_weapon", types.Position{})
		if weapon != nil {
			p.Equipment.Equip(weapon)
		}
		// Bonus: nice boost consumable
		niceBoost := NewItem("nice_boost", "cron_boost", types.Position{})
		if niceBoost != nil {
			p.Inventory.AddItem(niceBoost)
		}

	case ClassBash:
		// Bash: power hitter, gets offensive gear
		weapon := NewItem("pipe_wrench", "bash_weapon", types.Position{})
		armor := NewItem("basic_shell", "bash_armor", types.Position{})
		if weapon != nil {
			p.Equipment.Equip(weapon)
		}
		if armor != nil {
			p.Equipment.Equip(armor)
		}
		// Bonus: damage item
		coreDump := NewItem("core_dump", "bash_bomb", types.Position{})
		if coreDump != nil {
			p.Inventory.AddItem(coreDump)
		}

	case ClassVim:
		// Vim: complex class, gets rare weapon
		weapon := NewItem("vim_blade", "vim_weapon", types.Position{})
		if weapon != nil {
			p.Equipment.Equip(weapon)
		}
		// Bonus: FD restore
		fdRestore := NewItem("fd_restore", "vim_fd", types.Position{})
		if fdRestore != nil {
			p.Inventory.AddItem(fdRestore)
		}

	case ClassSudo:
		// Sudo: already has root, gets defensive gear
		armor := NewItem("firewall", "sudo_armor", types.Position{})
		if armor != nil {
			p.Equipment.Equip(armor)
		}
		// No weapon - starts unarmed but powerful
		// Bonus: sudo potion
		sudoPotion := NewItem("sudo_potion", "sudo_pot", types.Position{})
		if sudoPotion != nil {
			p.Inventory.AddItem(sudoPotion)
		}
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

// BuffType represents different buff effects.
type BuffType string

const (
	BuffInvincible BuffType = "invincible" // Immune to damage (sudo mode)
	BuffStrength   BuffType = "strength"   // Increased CPU
	BuffHaste      BuffType = "haste"      // Lower NICE (faster)
	BuffRegenRAM   BuffType = "regen_ram"  // Heal over time
	BuffRegenFD    BuffType = "regen_fd"   // FD restore over time
)

// Buff represents an active buff on the player.
type Buff struct {
	Type      BuffType
	Name      string
	Duration  int // Turns remaining
	Value     int // Effect magnitude (damage bonus, regen amount, etc.)
}

// GetStats returns the player's current stats.
func (p *Player) GetStats() types.Stats {
	return p.Stats
}

// TakeDamage reduces the player's RAM (health). Returns true if dead (OOM killed).
// Invincibility buff prevents all damage.
func (p *Player) TakeDamage(amount int) bool {
	if p.HasBuff(BuffInvincible) {
		return false // Immune to damage
	}
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

// AddBuff adds a buff to the player. If the buff type already exists, refreshes duration.
func (p *Player) AddBuff(buff Buff) {
	// Check if buff already exists
	for i, existing := range p.ActiveBuffs {
		if existing.Type == buff.Type {
			// Refresh duration (take the higher value)
			if buff.Duration > existing.Duration {
				p.ActiveBuffs[i].Duration = buff.Duration
			}
			// Stack value for some buffs
			if buff.Type == BuffStrength || buff.Type == BuffRegenRAM || buff.Type == BuffRegenFD {
				p.ActiveBuffs[i].Value += buff.Value
			}
			return
		}
	}
	p.ActiveBuffs = append(p.ActiveBuffs, buff)
}

// RemoveBuff removes a buff by type.
func (p *Player) RemoveBuff(buffType BuffType) {
	for i, buff := range p.ActiveBuffs {
		if buff.Type == buffType {
			p.ActiveBuffs = append(p.ActiveBuffs[:i], p.ActiveBuffs[i+1:]...)
			return
		}
	}
}

// HasBuff checks if the player has a specific buff type.
func (p *Player) HasBuff(buffType BuffType) bool {
	for _, buff := range p.ActiveBuffs {
		if buff.Type == buffType {
			return true
		}
	}
	return false
}

// GetBuff returns the buff of a specific type, or nil if not found.
func (p *Player) GetBuff(buffType BuffType) *Buff {
	for i := range p.ActiveBuffs {
		if p.ActiveBuffs[i].Type == buffType {
			return &p.ActiveBuffs[i]
		}
	}
	return nil
}

// TickBuffs decrements buff durations and applies effects. Call at end of each turn.
func (p *Player) TickBuffs() []string {
	var messages []string
	var remaining []Buff

	for i := range p.ActiveBuffs {
		buff := &p.ActiveBuffs[i]

		// Apply tick effects
		switch buff.Type {
		case BuffRegenRAM:
			healAmount := buff.Value
			p.Heal(healAmount)
			messages = append(messages, fmt.Sprintf("Regenerated %d RAM.", healAmount))
		case BuffRegenFD:
			restoreAmount := buff.Value
			p.RestoreFD(restoreAmount)
			messages = append(messages, fmt.Sprintf("Restored %d FD.", restoreAmount))
		}

		// Decrement duration
		buff.Duration--
		if buff.Duration > 0 {
			remaining = append(remaining, *buff)
		} else {
			messages = append(messages, fmt.Sprintf("%s wore off.", buff.Name))
		}
	}

	p.ActiveBuffs = remaining
	return messages
}

// GetEffectiveCPU returns CPU stat including buff bonuses.
func (p *Player) GetEffectiveCPU() int {
	cpu := p.Stats.CPU
	if buff := p.GetBuff(BuffStrength); buff != nil {
		cpu += buff.Value
	}
	return cpu
}

// GetEffectiveNICE returns NICE stat including buff bonuses.
func (p *Player) GetEffectiveNICE() int {
	nice := p.Stats.NICE
	if buff := p.GetBuff(BuffHaste); buff != nil {
		nice -= buff.Value // Lower NICE = faster
		if nice < 1 {
			nice = 1
		}
	}
	return nice
}
