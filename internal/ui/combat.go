package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// startCombat initializes combat with the given enemies.
func (m *Model) startCombat(enemies []*entity.Enemy) {
	m.enemies = enemies
	m.combat = m.engine.StartCombat(enemies)
	m.currentView = ViewCombat
	m.gameState = types.StateCombat
	m.combatCursor = 0
	m.targetCursor = 0 // Target first enemy
	m.selectingSkill = false
	m.combatLog = []string{}
	for _, enemy := range enemies {
		m.combatLog = append(m.combatLog, fmt.Sprintf("A wild %s appears!", enemy.Name()))
	}
}

// updateCombat handles combat view input.
func (m *Model) updateCombat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If selecting a skill, handle skill selection
	if m.selectingSkill {
		return m.updateSkillSelect(msg)
	}

	switch msg.String() {
	case "up", "k", "w":
		m.combatCursor--
		if m.combatCursor < 0 {
			m.combatCursor = 3 // 4 combat options
		}
	case "down", "j", "s":
		m.combatCursor++
		if m.combatCursor > 3 {
			m.combatCursor = 0
		}
	case "left", "h", "right", "l", "tab":
		// Cycle through targets
		m.cycleTarget(msg.String() == "left" || msg.String() == "h")
	case "enter", " ", "1", "2", "3", "4":
		return m.executeCombatAction(msg.String())
	case "esc":
		// Debug: instant flee
		m.endCombat(false)
		m.statusMsg = "[DEBUG] Escaped combat"
	}
	return m, nil
}

// getValidTargetIndex returns the actual enemy index for the current target.
// It maps the targetCursor (index into alive enemies) to the full enemy list.
func (m *Model) getValidTargetIndex() int {
	if m.combat == nil {
		return 0
	}

	aliveEnemies := m.combat.GetAliveEnemies()
	if len(aliveEnemies) == 0 {
		return 0
	}

	// Clamp target cursor
	if m.targetCursor >= len(aliveEnemies) {
		m.targetCursor = 0
	}

	// Find the actual index in the full enemies list
	targetEnemy := aliveEnemies[m.targetCursor]
	for i, enemy := range m.combat.Enemies {
		if enemy == targetEnemy {
			return i
		}
	}

	return 0
}

// cycleTarget cycles through alive enemies.
func (m *Model) cycleTarget(backward bool) {
	if m.combat == nil {
		return
	}

	aliveEnemies := m.combat.GetAliveEnemies()
	if len(aliveEnemies) <= 1 {
		return
	}

	// Find current target's index in alive list
	currentIdx := m.targetCursor
	if currentIdx >= len(aliveEnemies) {
		currentIdx = 0
	}

	// Cycle
	if backward {
		currentIdx--
		if currentIdx < 0 {
			currentIdx = len(aliveEnemies) - 1
		}
	} else {
		currentIdx++
		if currentIdx >= len(aliveEnemies) {
			currentIdx = 0
		}
	}

	// Find the actual index in the full enemies list
	m.targetCursor = currentIdx
}

// updateSkillSelect handles skill selection within combat.
func (m *Model) updateSkillSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	numSkills := len(m.player.Skills)
	if numSkills == 0 {
		m.selectingSkill = false
		return m, nil
	}

	switch msg.String() {
	case "up", "k", "w":
		m.skillCursor--
		if m.skillCursor < 0 {
			m.skillCursor = numSkills - 1
		}
	case "down", "j", "s":
		m.skillCursor++
		if m.skillCursor >= numSkills {
			m.skillCursor = 0
		}
	case "enter", " ":
		// Execute the selected skill
		m.selectingSkill = false
		return m.executeSkill(m.skillCursor)
	case "esc", "q":
		// Cancel skill selection
		m.selectingSkill = false
	case "1", "2", "3", "4", "5":
		// Quick select skill by number
		idx := int(msg.String()[0] - '1')
		if idx >= 0 && idx < numSkills {
			m.selectingSkill = false
			return m.executeSkill(idx)
		}
	}
	return m, nil
}

// executeCombatAction executes the selected combat action.
func (m *Model) executeCombatAction(key string) (tea.Model, tea.Cmd) {
	if m.combat == nil {
		return m, nil
	}

	// Map keys to actions
	actionIndex := m.combatCursor
	if key >= "1" && key <= "4" {
		actionIndex = int(key[0] - '1')
	}

	var action types.ActionType
	switch actionIndex {
	case 0: // Attack
		action = types.ActionAttack
	case 1: // Hack (skill) - open skill selection
		if m.player != nil && len(m.player.Skills) > 0 {
			m.selectingSkill = true
			m.skillCursor = 0
			return m, nil
		}
		m.combatLog = append(m.combatLog, "No skills available.")
		return m, nil
	case 2: // Use Item
		m.currentView = ViewInventory
		return m, nil
	case 3: // Flee
		action = types.ActionFlee
	}

	// Execute player action on selected target
	targetIdx := m.getValidTargetIndex()

	result := m.combat.ExecutePlayerAction(action, targetIdx, 0)
	m.combatLog = append(m.combatLog, result.Message)

	// Check for combat end conditions
	if result.Fled {
		m.endCombat(false)
		m.statusMsg = "You fled from combat!"
		return m, nil
	}

	if result.Victory {
		m.endCombat(true)
		return m, nil
	}

	// Enemy turns
	if !m.combat.IsOver {
		// Remember health before enemy turns for god mode
		preHP := m.player.Stats.RAM

		enemyResults := m.combat.ExecuteEnemyTurns()
		for _, er := range enemyResults {
			m.combatLog = append(m.combatLog, er.Message)
			if er.Defeat {
				if m.godMode {
					// God mode: restore health, ignore defeat
					m.player.Stats.RAM = preHP
					m.combatLog = append(m.combatLog, "[GOD MODE] Damage negated!")
				} else {
					m.endCombat(false)
					m.finishRun(false) // Handle all death logic
					return m, nil
				}
			}
		}

		// God mode: restore any damage taken
		if m.godMode && m.player.Stats.RAM < preHP {
			m.player.Stats.RAM = preHP
		}
	}

	// Tick cooldowns at start of new turn
	m.combat.TickCooldowns()

	return m, nil
}

// executeSkill executes a specific skill and handles enemy turns.
func (m *Model) executeSkill(skillIdx int) (tea.Model, tea.Cmd) {
	if m.combat == nil {
		return m, nil
	}

	// Use selected target
	targetIdx := m.getValidTargetIndex()

	// Execute skill
	result := m.combat.ExecutePlayerAction(types.ActionHack, targetIdx, skillIdx)
	m.combatLog = append(m.combatLog, result.Message)

	// Check for victory
	if result.Victory {
		m.endCombat(true)
		return m, nil
	}

	// Enemy turns
	if !m.combat.IsOver {
		preHP := m.player.Stats.RAM

		enemyResults := m.combat.ExecuteEnemyTurns()
		for _, er := range enemyResults {
			m.combatLog = append(m.combatLog, er.Message)
			if er.Defeat {
				if m.godMode {
					m.player.Stats.RAM = preHP
					m.combatLog = append(m.combatLog, "[GOD MODE] Damage negated!")
				} else {
					m.endCombat(false)
					m.finishRun(false) // Handle all death logic
					return m, nil
				}
			}
		}

		if m.godMode && m.player.Stats.RAM < preHP {
			m.player.Stats.RAM = preHP
		}
	}

	m.combat.TickCooldowns()
	return m, nil
}

// endCombat handles combat ending.
func (m *Model) endCombat(victory bool) {
	bossKilled := false
	if m.combat != nil && m.engine != nil {
		bossKilled = m.engine.EndCombat(m.combat)
	}

	m.combat = nil
	m.enemies = nil
	m.selectingSkill = false

	// Check for game victory (boss killed)
	if bossKilled {
		m.finishRun(true) // Handle all victory logic
		return
	}

	if victory {
		m.statusMsg = "Victory! Enemies defeated."
	}

	m.currentView = ViewGame
	m.gameState = types.StateExploring
}
