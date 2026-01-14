package game

import (
	"fmt"
	"math/rand"

	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/save"
	"github.com/iheanyi/devdungeon/internal/types"
)

// CombatState represents the current state of combat.
type CombatState struct {
	Player      *entity.Player
	Enemies     []*entity.Enemy
	CurrentTurn int // 0 = player, 1+ = enemy index
	TurnNumber  int
	Log         []string
	IsOver      bool
	Victory     bool
	FleeChance  float64
	rng         *rand.Rand
}

// CombatAction is an alias for types.ActionType.
type CombatAction = types.ActionType

// CombatResult represents the result of a combat action.
type CombatResult struct {
	Message    string
	Damage     int
	TargetName string
	IsCritical bool
	Missed     bool
	Fled       bool
	Victory    bool
	Defeat     bool
}

// NewCombatState creates a new combat state.
func NewCombatState(player *entity.Player, enemies []*entity.Enemy, seed int64) *CombatState {
	return &CombatState{
		Player:      player,
		Enemies:     enemies,
		CurrentTurn: 0,
		TurnNumber:  1,
		Log:         make([]string, 0),
		FleeChance:  0.5, // 50% base flee chance
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// ExecutePlayerAction executes the player's chosen action.
func (cs *CombatState) ExecutePlayerAction(action CombatAction, targetIdx int, skillIdx int) CombatResult {
	if cs.IsOver {
		return CombatResult{Message: "Combat is over."}
	}

	var result CombatResult

	switch action {
	case types.ActionAttack:
		result = cs.playerAttack(targetIdx)
	case types.ActionHack:
		result = cs.playerUseSkill(targetIdx, skillIdx)
	case types.ActionFlee:
		result = cs.playerFlee()
	case types.ActionUseItem:
		result = CombatResult{Message: "Select an item from inventory."}
	}

	// Check for victory
	if cs.allEnemiesDead() {
		cs.IsOver = true
		cs.Victory = true
		result.Victory = true
	}

	// If player didn't flee and combat isn't over, enemies take turn
	if !result.Fled && !cs.IsOver {
		cs.CurrentTurn = 1
	}

	return result
}

// ExecuteEnemyTurns executes all enemy turns.
func (cs *CombatState) ExecuteEnemyTurns() []CombatResult {
	var results []CombatResult

	if cs.IsOver {
		return results
	}

	for i, enemy := range cs.Enemies {
		if !enemy.IsAlive() {
			continue
		}

		result := cs.enemyTurn(i)
		results = append(results, result)
		cs.Log = append(cs.Log, result.Message)

		// Check for player defeat
		if !cs.Player.IsAlive() {
			cs.IsOver = true
			cs.Victory = false
			result.Defeat = true
			break
		}
	}

	// Back to player turn
	cs.CurrentTurn = 0
	cs.TurnNumber++

	return results
}

// playerAttack performs a basic attack.
func (cs *CombatState) playerAttack(targetIdx int) CombatResult {
	if targetIdx < 0 || targetIdx >= len(cs.Enemies) {
		return CombatResult{Message: "Invalid target."}
	}

	enemy := cs.Enemies[targetIdx]
	if !enemy.IsAlive() {
		return CombatResult{Message: fmt.Sprintf("%s is already dead.", enemy.Name())}
	}

	// Calculate damage
	baseDamage := cs.Player.Stats.CPU

	// Add weapon bonus
	if cs.Player.Equipment.Weapon != nil {
		baseDamage += cs.Player.Equipment.Weapon.StatBonus.CPU
	}

	// Variance: 80-120% of base damage
	variance := 0.8 + cs.rng.Float64()*0.4
	damage := int(float64(baseDamage) * variance)

	// Critical hit chance based on NICE (lower = faster = more crits)
	critChance := 0.05 + float64(20-cs.Player.Stats.NICE)/100.0
	isCritical := cs.rng.Float64() < critChance
	if isCritical {
		damage = int(float64(damage) * 1.5)
	}

	// Apply damage
	killed := enemy.TakeDamage(damage)

	result := CombatResult{
		Damage:     damage,
		TargetName: enemy.Name(),
		IsCritical: isCritical,
	}

	if isCritical {
		result.Message = fmt.Sprintf("CRITICAL! kill -TERM %s for %d damage!", enemy.Name(), damage)
	} else {
		result.Message = fmt.Sprintf("kill -TERM %s for %d damage.", enemy.Name(), damage)
	}

	if killed {
		result.Message += fmt.Sprintf(" %s was OOM killed!", enemy.Name())
	}

	cs.Log = append(cs.Log, result.Message)
	return result
}

// playerUseSkill uses a player skill.
func (cs *CombatState) playerUseSkill(targetIdx int, skillIdx int) CombatResult {
	if skillIdx < 0 || skillIdx >= len(cs.Player.Skills) {
		return CombatResult{Message: "Invalid skill."}
	}

	skill := &cs.Player.Skills[skillIdx]

	// Check cooldown
	if skill.CurrentCD > 0 {
		return CombatResult{Message: fmt.Sprintf("%s is on cooldown (%d turns).", skill.Name, skill.CurrentCD)}
	}

	// Check FD cost
	if cs.Player.Stats.FD < skill.FDCost {
		return CombatResult{Message: fmt.Sprintf("Not enough FD for %s (need %d, have %d).", skill.Name, skill.FDCost, cs.Player.Stats.FD)}
	}

	// Consume FD
	cs.Player.UseFD(skill.FDCost)

	// Set cooldown
	skill.CurrentCD = skill.Cooldown

	// Calculate damage
	damage := skill.BaseDamage + cs.Player.Stats.CPU/2

	// Apply to target
	if targetIdx >= 0 && targetIdx < len(cs.Enemies) {
		enemy := cs.Enemies[targetIdx]
		if enemy.IsAlive() {
			killed := enemy.TakeDamage(damage)
			msg := fmt.Sprintf("%s hits %s for %d damage!", skill.Name, enemy.Name(), damage)
			if killed {
				msg += fmt.Sprintf(" %s was OOM killed!", enemy.Name())
			}
			cs.Log = append(cs.Log, msg)
			return CombatResult{Message: msg, Damage: damage, TargetName: enemy.Name()}
		}
	}

	return CombatResult{Message: fmt.Sprintf("Used %s.", skill.Name)}
}

// playerFlee attempts to flee from combat.
func (cs *CombatState) playerFlee() CombatResult {
	// Flee chance affected by NICE (lower = faster = easier flee)
	fleeChance := cs.FleeChance + float64(10-cs.Player.Stats.NICE)/100.0

	if cs.rng.Float64() < fleeChance {
		cs.IsOver = true
		cs.Victory = false
		msg := "You successfully fled from combat!"
		cs.Log = append(cs.Log, msg)
		return CombatResult{Message: msg, Fled: true}
	}

	msg := "Failed to flee!"
	cs.Log = append(cs.Log, msg)
	return CombatResult{Message: msg, Fled: false}
}

// enemyTurn performs an enemy's turn based on their behavior.
func (cs *CombatState) enemyTurn(enemyIdx int) CombatResult {
	enemy := cs.Enemies[enemyIdx]

	// Choose action based on behavior
	switch enemy.Behavior {
	case entity.BehaviorAggressive:
		return cs.enemyAttack(enemy)

	case entity.BehaviorDefensive:
		// Heal when below 30% health
		healthPct := float64(enemy.Stats.RAM) / float64(enemy.MaxStats.MaxRAM)
		if healthPct < 0.3 && cs.rng.Float64() < 0.6 {
			return cs.enemyHeal(enemy)
		}
		return cs.enemyAttack(enemy)

	case entity.BehaviorErratic:
		// Random action
		roll := cs.rng.Float64()
		if roll < 0.6 {
			return cs.enemyAttack(enemy)
		} else if roll < 0.8 {
			return cs.enemyWildSwing(enemy) // High damage, might miss
		} else {
			return cs.enemyConfused(enemy) // Does nothing
		}

	case entity.BehaviorSwarm:
		// Fork bomb: chance to spawn another enemy
		if len(cs.GetAliveEnemies()) < 4 && cs.rng.Float64() < 0.3 {
			return cs.enemyFork(enemy)
		}
		return cs.enemyAttack(enemy)

	case entity.BehaviorStealth:
		// Ambush: higher crit chance, lower base damage
		return cs.enemyAmbush(enemy)

	default:
		return cs.enemyAttack(enemy)
	}
}

// enemyAttack performs a basic enemy attack.
func (cs *CombatState) enemyAttack(enemy *entity.Enemy) CombatResult {
	baseDamage := enemy.Stats.CPU

	// Variance: 80-120%
	variance := 0.8 + cs.rng.Float64()*0.4
	damage := int(float64(baseDamage) * variance)

	// Apply armor reduction
	if cs.Player.Equipment.Armor != nil {
		reduction := cs.Player.Equipment.Armor.StatBonus.RAM / 5
		damage -= reduction
		if damage < 1 {
			damage = 1
		}
	}

	cs.Player.TakeDamage(damage)

	return CombatResult{
		Damage:     damage,
		TargetName: "you",
		Message:    fmt.Sprintf("%s attacks for %d damage!", enemy.Name(), damage),
	}
}

// enemyHeal heals the enemy (defensive behavior).
func (cs *CombatState) enemyHeal(enemy *entity.Enemy) CombatResult {
	healAmount := enemy.MaxStats.MaxRAM / 4
	enemy.Stats.RAM += healAmount
	if enemy.Stats.RAM > enemy.MaxStats.MaxRAM {
		enemy.Stats.RAM = enemy.MaxStats.MaxRAM
	}

	return CombatResult{
		Message: fmt.Sprintf("%s allocates memory, healing %d RAM!", enemy.Name(), healAmount),
	}
}

// enemyWildSwing is an erratic high-risk attack.
func (cs *CombatState) enemyWildSwing(enemy *entity.Enemy) CombatResult {
	// 40% miss chance
	if cs.rng.Float64() < 0.4 {
		return CombatResult{
			Missed:  true,
			Message: fmt.Sprintf("%s swings wildly and misses!", enemy.Name()),
		}
	}

	// 1.5x damage if it hits
	damage := int(float64(enemy.Stats.CPU) * 1.5)

	if cs.Player.Equipment.Armor != nil {
		reduction := cs.Player.Equipment.Armor.StatBonus.RAM / 5
		damage -= reduction
		if damage < 1 {
			damage = 1
		}
	}

	cs.Player.TakeDamage(damage)

	return CombatResult{
		Damage:     damage,
		TargetName: "you",
		Message:    fmt.Sprintf("%s lands a wild swing for %d damage!", enemy.Name(), damage),
	}
}

// enemyConfused does nothing (erratic behavior).
func (cs *CombatState) enemyConfused(enemy *entity.Enemy) CombatResult {
	messages := []string{
		fmt.Sprintf("%s is confused and does nothing.", enemy.Name()),
		fmt.Sprintf("%s segfaults momentarily.", enemy.Name()),
		fmt.Sprintf("%s throws a null pointer at nothing.", enemy.Name()),
	}
	return CombatResult{
		Message: messages[cs.rng.Intn(len(messages))],
	}
}

// enemyFork spawns a copy (swarm behavior).
func (cs *CombatState) enemyFork(enemy *entity.Enemy) CombatResult {
	// Create a weaker copy
	fork := entity.NewEnemy(enemy.Type, fmt.Sprintf("%s_fork_%d", enemy.ID(), cs.rng.Int()), types.Position{}, 1)
	fork.Stats.RAM = fork.Stats.RAM / 2 // Half health
	fork.Stats.CPU = fork.Stats.CPU / 2 // Half damage

	cs.Enemies = append(cs.Enemies, fork)

	return CombatResult{
		Message: fmt.Sprintf("%s forks! A new %s spawns!", enemy.Name(), fork.Name()),
	}
}

// enemyAmbush is a stealth attack with high crit chance.
func (cs *CombatState) enemyAmbush(enemy *entity.Enemy) CombatResult {
	baseDamage := enemy.Stats.CPU

	// 50% crit chance for stealth enemies
	isCrit := cs.rng.Float64() < 0.5
	if isCrit {
		baseDamage = int(float64(baseDamage) * 1.8)
	}

	variance := 0.9 + cs.rng.Float64()*0.2
	damage := int(float64(baseDamage) * variance)

	if cs.Player.Equipment.Armor != nil {
		reduction := cs.Player.Equipment.Armor.StatBonus.RAM / 5
		damage -= reduction
		if damage < 1 {
			damage = 1
		}
	}

	cs.Player.TakeDamage(damage)

	msg := fmt.Sprintf("%s strikes from the shadows for %d damage!", enemy.Name(), damage)
	if isCrit {
		msg = fmt.Sprintf("AMBUSH! %s", msg)
	}

	return CombatResult{
		Damage:     damage,
		TargetName: "you",
		IsCritical: isCrit,
		Message:    msg,
	}
}

// allEnemiesDead returns true if all enemies are dead.
func (cs *CombatState) allEnemiesDead() bool {
	for _, enemy := range cs.Enemies {
		if enemy.IsAlive() {
			return false
		}
	}
	return true
}

// GetAliveEnemies returns a list of alive enemies.
func (cs *CombatState) GetAliveEnemies() []*entity.Enemy {
	var alive []*entity.Enemy
	for _, enemy := range cs.Enemies {
		if enemy.IsAlive() {
			alive = append(alive, enemy)
		}
	}
	return alive
}

// TickCooldowns reduces skill cooldowns at the start of player turn.
func (cs *CombatState) TickCooldowns() {
	for i := range cs.Player.Skills {
		if cs.Player.Skills[i].CurrentCD > 0 {
			cs.Player.Skills[i].CurrentCD--
		}
	}
}

// CalculateRewards calculates XP and loot from defeated enemies.
func (cs *CombatState) CalculateRewards() (xp int, loot []*entity.Item) {
	for _, enemy := range cs.Enemies {
		if !enemy.IsAlive() {
			xp += enemy.XPReward

			// Roll for loot drops
			for _, itemID := range enemy.LootTable {
				// 50% drop chance per item in loot table
				if rand.Float64() < 0.5 {
					item := entity.NewItem(itemID, fmt.Sprintf("loot_%s_%d", itemID, rand.Int()), types.Position{})
					if item != nil {
						loot = append(loot, item)
					}
				}
			}
		}
	}
	return xp, loot
}

// --- Engine integration ---

// StartCombat initiates combat with enemies.
func (e *Engine) StartCombat(enemies []*entity.Enemy) *CombatState {
	e.state = types.StateCombat
	return NewCombatState(e.player, enemies, e.rng.Int63())
}

// EndCombat handles combat conclusion.
func (e *Engine) EndCombat(combat *CombatState) {
	if combat.Victory {
		// Calculate rewards
		xp, loot := combat.CalculateRewards()

		// Grant XP
		if e.player.GainXP(xp) {
			e.addMessage("LEVEL UP! You are now level %d!", e.player.Level)
		} else {
			e.addMessage("Gained %d XP.", xp)
		}

		// Add loot to inventory
		for _, item := range loot {
			if e.player.Inventory.AddItem(item) {
				e.addMessage("Looted: %s", item.Name())
			} else {
				e.addMessage("Inventory full, couldn't pick up %s.", item.Name())
			}
		}

		// Remove dead enemies from world
		for _, enemy := range combat.Enemies {
			if !enemy.IsAlive() {
				e.world.RemoveEnemy(enemy.ID())
			}
		}

		// Save after combat victory
		e.Save(save.TriggerCombatVictory)
	}

	e.state = types.StateExploring
}
