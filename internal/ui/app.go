// Package ui provides the Bubble Tea UI for /dev/dungeon.
package ui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/dungeon"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/types"
)

// ViewType represents different UI views.
type ViewType int

const (
	ViewMainMenu ViewType = iota
	ViewClassSelect
	ViewGame
	ViewCombat
	ViewInventory
	ViewPause
	ViewGameOver
	ViewVictory
	ViewAdmin
	ViewHelp
	ViewMessageHistory
	ViewIntro
	ViewShop
)

// introTickMsg is sent to advance intro animation.
type introTickMsg struct{}

// introFrames contains the animated intro sequence.
var introFrames = []string{
	`
      ╔══════════════════════════════════════════════════════════════╗
      ║                                                              ║
      ║    ██████╗ ███████╗██╗   ██╗   ██████╗ ██╗   ██╗███╗   ██╗   ║
      ║    ██╔══██╗██╔════╝██║   ██║   ██╔══██╗██║   ██║████╗  ██║   ║
      ║    ██║  ██║█████╗  ██║   ██║   ██║  ██║██║   ██║██╔██╗ ██║   ║
      ║    ██║  ██║██╔══╝  ╚██╗ ██╔╝   ██║  ██║██║   ██║██║╚██╗██║   ║
      ║    ██████╔╝███████╗ ╚████╔╝    ██████╔╝╚██████╔╝██║ ╚████║   ║
      ║    ╚═════╝ ╚══════╝  ╚═══╝     ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝   ║
      ║                                                              ║
      ╚══════════════════════════════════════════════════════════════╝
`,
	`
      ╔══════════════════════════════════════════════════════════════╗
      ║                                                              ║
      ║                    >>> SYSTEM ALERT <<<                      ║
      ║                                                              ║
      ║                    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                        ║
      ║                    ▓  KERNEL PANIC  ▓                        ║
      ║                    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                        ║
      ║                                                              ║
      ║           The system has been compromised.                   ║
      ║                                                              ║
      ╚══════════════════════════════════════════════════════════════╝
`,
	`
      ╔══════════════════════════════════════════════════════════════╗
      ║                                                              ║
      ║  Something has gone wrong deep in /dev/null.                 ║
      ║                                                              ║
      ║  Rogue processes have escaped.                               ║
      ║  Zombie processes walk the directories.                      ║
      ║  Daemons have turned hostile.                                ║
      ║  Fork bombs multiply unchecked.                              ║
      ║                                                              ║
      ║  And at the heart of it all...                               ║
      ║  The KERNEL PANIC awaits.                                    ║
      ║                                                              ║
      ╚══════════════════════════════════════════════════════════════╝
`,
	`
      ╔══════════════════════════════════════════════════════════════╗
      ║                                                              ║
      ║  You are a newly spawned process.                            ║
      ║                                                              ║
      ║  Your mission: Navigate from /home through the filesystem,   ║
      ║  descending ever deeper into the system.                     ║
      ║                                                              ║
      ║  /home → /tmp → /var → /etc → /usr → /bin → /sys → /dev/null ║
      ║                                                              ║
      ║  Fight. Survive. Find the KERNEL PANIC.                      ║
      ║  And end this madness.                                       ║
      ║                                                              ║
      ╚══════════════════════════════════════════════════════════════╝
`,
	`
      ╔══════════════════════════════════════════════════════════════╗
      ║                                                              ║
      ║               YOUR STATS EXPLAINED:                          ║
      ║                                                              ║
      ║    RAM   - Health. Reach 0 = OOM killed.                     ║
      ║    CPU   - Attack power. Kill processes faster.              ║
      ║    FD    - File descriptors. Fuel your abilities.            ║
      ║    NICE  - Priority. Lower = faster + more crits.            ║
      ║    UID   - User ID. 0 = root = ultimate power.               ║
      ║                                                              ║
      ║               Good luck, process.                            ║
      ║               The system depends on you.                     ║
      ║                                                              ║
      ╚══════════════════════════════════════════════════════════════╝
`,
}

// Model is the main Bubble Tea model for the game.
type Model struct {
	// Core game state
	config    *config.Config
	engine    *game.Engine
	player    *entity.Player
	gameState types.GameState

	// UI state
	currentView  ViewType
	width        int
	height       int

	// Menu state
	menuCursor   int
	menuOptions  []string

	// Combat state
	combatCursor    int
	combatLog       []string
	enemies         []*entity.Enemy
	combat          *game.CombatState
	selectingSkill  bool
	skillCursor     int
	targetCursor    int // Which enemy is targeted

	// Inventory state
	invCursor    int

	// Class selection state
	classCursor  int
	classOptions []entity.PlayerClass

	// Admin console state
	adminCursor  int
	adminOptions []string
	godMode      bool
	prevView     ViewType // View to return to after admin

	// Messages
	statusMsg         string
	messageHistory    []string // Full message history
	messageScrollIdx  int      // Current scroll position (0 = most recent)
	showingHistory    bool     // Whether message history view is active

	// Intro animation state
	introFrame     int  // Current frame of intro
	introSkipped   bool // User skipped intro
	pendingClass   entity.PlayerClass // Class to use after intro

	// Shop state
	shopCursor     int
	shopItems      []ShopItem

	// Styles
	styles       *Styles
}

// ShopItem represents an item for sale.
type ShopItem struct {
	TemplateID string
	Name       string
	Price      int
	InStock    bool
}

// Styles holds all UI styles.
type Styles struct {
	// Layout
	Container    lipgloss.Style
	Header       lipgloss.Style
	Footer       lipgloss.Style

	// Game view
	MapBorder    lipgloss.Style
	StatPanel    lipgloss.Style
	LogPanel     lipgloss.Style

	// Text
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Normal       lipgloss.Style
	Highlight    lipgloss.Style
	Danger       lipgloss.Style
	Success      lipgloss.Style
	Muted        lipgloss.Style

	// Menu
	MenuItem     lipgloss.Style
	MenuSelected lipgloss.Style

	// Tiles
	Wall         lipgloss.Style
	Floor        lipgloss.Style
	Player       lipgloss.Style
	Enemy        lipgloss.Style
	Item         lipgloss.Style
	Stairs       lipgloss.Style
}

// NewStyles creates the default styles.
func NewStyles() *Styles {
	return &Styles{
		Container: lipgloss.NewStyle().
			Padding(1, 2),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("240")),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),

		MapBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")),
		StatPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1),
		LogPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		Highlight: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		Danger: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),
		Success: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46")),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),

		MenuItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		MenuSelected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236")),

		Wall: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		Floor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		Player: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		Enemy: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")),
		Stairs: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("45")),
	}
}

// New creates a new UI model.
func New(cfg *config.Config) *Model {
	return &Model{
		config:      cfg,
		currentView: ViewMainMenu,
		gameState:   types.StateMainMenu,
		menuOptions: []string{
			"New Game",
			"Continue",
			"Settings",
			"Quit",
		},
		menuCursor: 0,
		classOptions: []entity.PlayerClass{
			entity.ClassInit,
			entity.ClassCron,
			entity.ClassBash,
			entity.ClassVim,
			entity.ClassSudo,
		},
		classCursor: 0,
		adminOptions: []string{
			"Toggle God Mode",
			"Full Heal",
			"Give XP (+100)",
			"Spawn Item",
			"Teleport to Stairs",
			"Skip to Next Floor",
			"Show Debug Info",
			"Close",
		},
		adminCursor:     0,
		combatLog:       make([]string, 0),
		messageHistory:  make([]string, 0),
		styles:          NewStyles(),
	}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	// Request initial window size
	return tea.EnterAltScreen
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case introTickMsg:
		// Handle intro animation tick
		if m.currentView == ViewIntro && !m.introSkipped {
			m.introFrame++
			if m.introFrame >= len(introFrames) {
				// Intro complete, start the game
				m.startNewGame(m.pendingClass)
				return m, nil
			}
			// Continue to next frame
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return introTickMsg{}
			})
		}
		return m, nil
	}
	return m, nil
}

// handleKeyPress processes keyboard input.
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C ALWAYS quits - this is the universal escape hatch
	if msg.String() == "ctrl+c" {
		m.shutdown()
		return m, tea.Quit
	}

	// View-specific input
	switch m.currentView {
	case ViewMainMenu:
		return m.updateMainMenu(msg)
	case ViewClassSelect:
		return m.updateClassSelect(msg)
	case ViewGame:
		return m.updateGame(msg)
	case ViewCombat:
		return m.updateCombat(msg)
	case ViewInventory:
		return m.updateInventory(msg)
	case ViewPause:
		return m.updatePause(msg)
	case ViewGameOver:
		return m.updateGameOver(msg)
	case ViewAdmin:
		return m.updateAdmin(msg)
	case ViewHelp:
		return m.updateHelp(msg)
	case ViewMessageHistory:
		return m.updateMessageHistory(msg)
	case ViewIntro:
		return m.updateIntro(msg)
	case ViewShop:
		return m.updateShop(msg)
	}

	return m, nil
}

// shutdown gracefully shuts down the game.
func (m *Model) shutdown() {
	if m.engine != nil {
		m.engine.Shutdown()
	}
}

// updateMainMenu handles main menu input.
func (m *Model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "w":
		m.menuCursor--
		if m.menuCursor < 0 {
			m.menuCursor = len(m.menuOptions) - 1
		}
	case "down", "j", "s":
		m.menuCursor++
		if m.menuCursor >= len(m.menuOptions) {
			m.menuCursor = 0
		}
	case "enter", " ":
		return m.selectMenuItem()
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

// selectMenuItem handles menu selection.
func (m *Model) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.menuOptions[m.menuCursor] {
	case "New Game":
		m.classCursor = 0
		m.currentView = ViewClassSelect
		return m, nil
	case "Continue":
		m.continueGame()
		return m, nil
	case "Settings":
		// TODO: Settings screen
		m.statusMsg = "Settings not yet implemented"
		return m, nil
	case "Quit":
		m.shutdown()
		return m, tea.Quit
	}
	return m, nil
}

// updateClassSelect handles class selection input.
func (m *Model) updateClassSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "w":
		m.classCursor--
		if m.classCursor < 0 {
			m.classCursor = len(m.classOptions) - 1
		}
	case "down", "j", "s":
		m.classCursor++
		if m.classCursor >= len(m.classOptions) {
			m.classCursor = 0
		}
	case "enter", " ":
		// Start intro sequence, then game
		m.pendingClass = m.classOptions[m.classCursor]
		m.introFrame = 0
		m.introSkipped = false
		m.currentView = ViewIntro
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return introTickMsg{}
		})
	case "esc", "q":
		m.currentView = ViewMainMenu
	}
	return m, nil
}

// continueGame loads and continues a saved game.
func (m *Model) continueGame() {
	// Create a temporary engine to check for saves
	if m.engine == nil {
		m.engine = game.NewEngine(m.config, 0)

		// Set up the dungeon generator
		dungeonCfg := dungeon.DefaultConfig()
		dungeonCfg.Width = m.config.Display.MapWidth
		dungeonCfg.Height = m.config.Display.MapHeight
		m.engine.SetGenerator(dungeon.NewGenerator(dungeonCfg))
	}

	// Try to load the latest save
	if err := m.engine.LoadLatestSave(); err != nil {
		m.statusMsg = err.Error()
		return
	}

	// Get the player from the engine
	m.player = m.engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring
	m.statusMsg = fmt.Sprintf("Welcome back to %s.", m.engine.CurrentFloorType().FloorName())
}

// startNewGame initializes a new game with the selected class.
func (m *Model) startNewGame(playerClass entity.PlayerClass) {
	// Create the game engine with seed 0 (random)
	m.engine = game.NewEngine(m.config, 0)

	// Set up the dungeon generator
	dungeonCfg := dungeon.DefaultConfig()
	dungeonCfg.Width = m.config.Display.MapWidth
	dungeonCfg.Height = m.config.Display.MapHeight
	m.engine.SetGenerator(dungeon.NewGenerator(dungeonCfg))

	// Start a new game with the selected class
	if err := m.engine.StartNewGame(playerClass); err != nil {
		m.statusMsg = fmt.Sprintf("Failed to start game: %v", err)
		m.currentView = ViewMainMenu
		return
	}

	// Get the player from the engine
	m.player = m.engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring
	m.statusMsg = fmt.Sprintf("Spawned as %s. Welcome to %s.", playerClass, m.engine.CurrentFloorType().FloorName())
}

// updateGame handles game view input.
func (m *Model) updateGame(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "w", "up", "k":
		m.movePlayer(types.DirUp)
	case "s", "down", "j":
		m.movePlayer(types.DirDown)
	case "a", "left", "h":
		m.movePlayer(types.DirLeft)
	case "d", "right", "l":
		m.movePlayer(types.DirRight)
	case ">", ".":
		// Descend stairs
		if m.engine != nil {
			if err := m.engine.DescendStairs(); err != nil {
				m.statusMsg = err.Error()
			} else {
				m.statusMsg = fmt.Sprintf("Descended to %s (depth %d).", m.engine.CurrentFloorType().FloorName(), m.engine.CurrentDepth())
			}
		}
	case "<", ",":
		// Ascend stairs
		if m.engine != nil {
			if err := m.engine.AscendStairs(); err != nil {
				m.statusMsg = err.Error()
			} else {
				m.statusMsg = fmt.Sprintf("Ascended to %s (depth %d).", m.engine.CurrentFloorType().FloorName(), m.engine.CurrentDepth())
			}
		}
	case "i":
		m.currentView = ViewInventory
	case "p", "esc":
		m.currentView = ViewPause
	case "q":
		// Save and return to main menu
		if m.engine != nil {
			m.engine.Shutdown()
			m.engine = nil
		}
		m.player = nil
		m.currentView = ViewMainMenu
		m.gameState = types.StateMainMenu
		m.statusMsg = "Game saved."
	case "`":
		// Open admin console - requires root (UID 0) OR actual sudo
		hasGameRoot := m.player != nil && m.player.Stats.UID == 0
		hasRealRoot := os.Geteuid() == 0
		if hasGameRoot || hasRealRoot {
			m.prevView = ViewGame
			m.adminCursor = 0
			m.currentView = ViewAdmin
			if hasRealRoot && !hasGameRoot {
				m.statusMsg = "[REAL SUDO DETECTED] Admin access granted."
			}
		} else {
			m.statusMsg = "Permission denied: requires root. Try 'sudo' class, find root_shard, or run with actual sudo."
		}
	case "?":
		m.prevView = ViewGame
		m.currentView = ViewHelp
	case "m":
		m.prevView = ViewGame
		m.messageScrollIdx = 0
		m.currentView = ViewMessageHistory
	case "$":
		// Open shop
		m.openShop()
	}
	return m, nil
}

// movePlayer moves the player in a direction using the game engine.
func (m *Model) movePlayer(dir types.Direction) {
	if m.engine == nil || m.player == nil {
		return
	}

	result := m.engine.MovePlayer(dir)

	// Sync any messages from engine
	m.syncEngineMessages()

	// Update status message
	if result.Message != "" {
		m.addToHistory(result.Message)
	}

	// Check for combat initiation
	if result.Combat != nil {
		m.startCombat([]*entity.Enemy{result.Combat})
	}
}

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
					m.currentView = ViewGameOver
					m.gameState = types.StateGameOver
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
					m.currentView = ViewGameOver
					m.gameState = types.StateGameOver
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
		m.currentView = ViewVictory
		m.gameState = types.StateVictory
		m.statusMsg = "KERNEL PANIC DEFEATED! You saved the system!"
		return
	}

	if victory {
		m.statusMsg = "Victory! Enemies defeated."
	}

	m.currentView = ViewGame
	m.gameState = types.StateExploring
}

// updateInventory handles inventory view input.
func (m *Model) updateInventory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.player == nil {
		return m, nil
	}

	// Total items = inventory + 4 equipment slots
	invLen := len(m.player.Inventory.Items)
	totalItems := invLen + 4 // weapon, armor, util1, util2
	if totalItems == 4 {
		totalItems = 4 // At minimum, show equipment slots
	}

	switch msg.String() {
	case "esc", "i":
		if m.gameState == types.StateCombat {
			m.currentView = ViewCombat
		} else {
			m.currentView = ViewGame
		}
	case "up", "k", "w":
		m.invCursor--
		if m.invCursor < 0 {
			m.invCursor = totalItems - 1
		}
	case "down", "j", "s":
		m.invCursor++
		if m.invCursor >= totalItems {
			m.invCursor = 0
		}
	case "enter", " ":
		m.useOrEquipItem()
	case "e":
		m.equipItem()
	case "d":
		m.dropItem()
	case "u":
		m.unequipItem()
	case "?":
		m.showItemDetails()
	}
	return m, nil
}

// useOrEquipItem uses consumables or equips equipment.
func (m *Model) useOrEquipItem() {
	if m.player == nil || m.invCursor >= len(m.player.Inventory.Items) {
		return
	}

	item := m.player.Inventory.Items[m.invCursor]

	switch item.ItemType {
	case entity.ItemTypeConsumable:
		m.useConsumable(item)
	case entity.ItemTypeWeapon, entity.ItemTypeArmor, entity.ItemTypeUtility:
		m.equipItem()
	default:
		m.statusMsg = fmt.Sprintf("Cannot use %s.", item.Name())
	}
}

// useConsumable applies a consumable item's effects.
func (m *Model) useConsumable(item *entity.Item) {
	if item == nil {
		return
	}

	var effectMsg string
	for _, effect := range item.Effects {
		switch effect.Type {
		case entity.EffectHeal:
			oldRAM := m.player.Stats.RAM
			m.player.Heal(effect.Value)
			healed := m.player.Stats.RAM - oldRAM
			effectMsg = fmt.Sprintf("Allocated %d RAM.", healed)

		case entity.EffectRestoreFD:
			oldFD := m.player.Stats.FD
			m.player.RestoreFD(effect.Value)
			restored := m.player.Stats.FD - oldFD
			effectMsg = fmt.Sprintf("Restored %d FD.", restored)

		case entity.EffectDamage:
			// In combat, damage first enemy
			if m.combat != nil && len(m.combat.Enemies) > 0 {
				for _, enemy := range m.combat.Enemies {
					if enemy.IsAlive() {
						killed := enemy.TakeDamage(effect.Value)
						effectMsg = fmt.Sprintf("Dealt %d damage to %s!", effect.Value, enemy.Name())
						if killed {
							effectMsg += " OOM killed!"
						}
						break
					}
				}
			} else {
				effectMsg = "No target for damage item."
				return // Don't consume
			}

		case entity.EffectBuff:
			// Determine buff type based on item
			var buffType entity.BuffType
			var buffValue int
			var duration int

			switch item.TemplateID {
			case "sudo_potion":
				buffType = entity.BuffInvincible
				buffValue = 0
				duration = 3 // 3 turns of invincibility
				effectMsg = "ROOT ACCESS GRANTED! Immune to damage for 3 turns."
			case "nice_boost":
				buffType = entity.BuffHaste
				buffValue = 5 // -5 NICE
				duration = 5
				effectMsg = "NICE reduced! Acting faster for 5 turns."
			case "cpu_boost":
				buffType = entity.BuffStrength
				buffValue = 10 // +10 CPU
				duration = 5
				effectMsg = "CPU boosted! +10 attack for 5 turns."
			default:
				effectMsg = fmt.Sprintf("Gained buff from %s.", item.Name())
				buffType = entity.BuffStrength
				buffValue = 5
				duration = 3
			}

			m.player.AddBuff(entity.Buff{
				Type:     buffType,
				Name:     item.Name(),
				Duration: duration,
				Value:    buffValue,
			})

		case entity.EffectReveal:
			effectMsg = "Revealed floor contents."
			// TODO: Implement reveal

		default:
			effectMsg = fmt.Sprintf("Used %s.", item.Name())
		}
	}

	// Consume the item
	if item.Stackable && item.Quantity > 1 {
		item.Quantity--
	} else {
		m.player.Inventory.RemoveItem(item.ID())
		// Adjust cursor if needed
		if m.invCursor >= len(m.player.Inventory.Items) && m.invCursor > 0 {
			m.invCursor--
		}
	}

	m.addToHistory(effectMsg)

	// If in combat, add to combat log
	if m.combat != nil {
		m.combatLog = append(m.combatLog, effectMsg)
	}
}

// addToHistory adds a message to the status and history.
func (m *Model) addToHistory(msg string) {
	m.statusMsg = msg
	if msg != "" {
		m.messageHistory = append(m.messageHistory, msg)
		// Cap history at 500 messages
		if len(m.messageHistory) > 500 {
			m.messageHistory = m.messageHistory[1:]
		}
	}
}

// syncEngineMessages syncs messages from the engine into history.
func (m *Model) syncEngineMessages() {
	if m.engine == nil {
		return
	}
	messages := m.engine.Messages()
	for _, msg := range messages {
		m.addToHistory(msg)
	}
	m.engine.ClearMessages()
}

// equipItem equips the selected item.
func (m *Model) equipItem() {
	if m.player == nil || m.invCursor >= len(m.player.Inventory.Items) {
		return
	}

	item := m.player.Inventory.Items[m.invCursor]

	// Only equip weapons, armor, utility
	if item.EquipSlot == entity.SlotNone {
		m.statusMsg = fmt.Sprintf("Cannot equip %s.", item.Name())
		return
	}

	// Remove from inventory
	m.player.Inventory.RemoveItem(item.ID())

	// Equip (returns old item if any)
	oldItem := m.player.Equipment.Equip(item)

	// Put old item back in inventory
	if oldItem != nil {
		m.player.Inventory.AddItem(oldItem)
	}

	// Adjust cursor
	if m.invCursor >= len(m.player.Inventory.Items) && m.invCursor > 0 {
		m.invCursor--
	}

	m.statusMsg = fmt.Sprintf("Equipped %s.", item.Name())
}

// dropItem drops the selected item.
func (m *Model) dropItem() {
	if m.player == nil || m.invCursor >= len(m.player.Inventory.Items) {
		return
	}

	item := m.player.Inventory.Items[m.invCursor]
	m.player.Inventory.RemoveItem(item.ID())

	// Place item at player's position in the world
	if m.engine != nil {
		item.SetPosition(m.player.Position())
		m.engine.GetWorld().AddItem(item)
	}

	// Adjust cursor
	if m.invCursor >= len(m.player.Inventory.Items) && m.invCursor > 0 {
		m.invCursor--
	}

	m.statusMsg = fmt.Sprintf("Dropped %s.", item.Name())
}

// unequipItem removes equipped item and puts it back in inventory.
func (m *Model) unequipItem() {
	if m.player == nil {
		return
	}

	// Check if cursor is in equipment section (after inventory items)
	invLen := len(m.player.Inventory.Items)

	// Equipment slots: 0=weapon, 1=armor, 2=utility1, 3=utility2 (relative to after inventory)
	equipIdx := m.invCursor - invLen
	if equipIdx < 0 {
		m.statusMsg = "Select an equipped item to unequip."
		return
	}

	var item *entity.Item
	var slotName string

	switch equipIdx {
	case 0:
		item = m.player.Equipment.Weapon
		slotName = "weapon"
		if item != nil {
			m.player.Equipment.Weapon = nil
		}
	case 1:
		item = m.player.Equipment.Armor
		slotName = "armor"
		if item != nil {
			m.player.Equipment.Armor = nil
		}
	case 2:
		item = m.player.Equipment.Utility1
		slotName = "utility 1"
		if item != nil {
			m.player.Equipment.Utility1 = nil
		}
	case 3:
		item = m.player.Equipment.Utility2
		slotName = "utility 2"
		if item != nil {
			m.player.Equipment.Utility2 = nil
		}
	default:
		m.statusMsg = "Invalid equipment slot."
		return
	}

	if item == nil {
		m.statusMsg = fmt.Sprintf("No %s equipped.", slotName)
		return
	}

	// Add back to inventory
	if m.player.Inventory.AddItem(item) {
		m.statusMsg = fmt.Sprintf("Unequipped %s.", item.Name())
	} else {
		// Inventory full, re-equip
		switch equipIdx {
		case 0:
			m.player.Equipment.Weapon = item
		case 1:
			m.player.Equipment.Armor = item
		case 2:
			m.player.Equipment.Utility1 = item
		case 3:
			m.player.Equipment.Utility2 = item
		}
		m.statusMsg = "Inventory full, cannot unequip."
	}
}

// showItemDetails shows detailed info about the selected item.
func (m *Model) showItemDetails() {
	if m.player == nil {
		return
	}

	var item *entity.Item
	invLen := len(m.player.Inventory.Items)

	if m.invCursor < invLen {
		// Inventory item
		item = m.player.Inventory.Items[m.invCursor]
	} else {
		// Equipment slot
		equipIdx := m.invCursor - invLen
		switch equipIdx {
		case 0:
			item = m.player.Equipment.Weapon
		case 1:
			item = m.player.Equipment.Armor
		case 2:
			item = m.player.Equipment.Utility1
		case 3:
			item = m.player.Equipment.Utility2
		}
	}

	if item == nil {
		m.statusMsg = "Empty slot - nothing equipped."
		return
	}

	// Build detailed description
	details := fmt.Sprintf("=== %s ===\n", item.Name())
	details += fmt.Sprintf("Type: %s | Rarity: %s\n", string(item.ItemType), item.Rarity.String())
	details += fmt.Sprintf("%s\n", item.Description)

	// Show stat bonuses
	if item.StatBonus.CPU != 0 || item.StatBonus.RAM != 0 || item.StatBonus.FD != 0 || item.StatBonus.NICE != 0 || item.StatBonus.UID != 0 {
		details += "Stats: "
		bonuses := []string{}
		if item.StatBonus.CPU != 0 {
			bonuses = append(bonuses, fmt.Sprintf("CPU %+d", item.StatBonus.CPU))
		}
		if item.StatBonus.RAM != 0 {
			bonuses = append(bonuses, fmt.Sprintf("RAM %+d", item.StatBonus.RAM))
		}
		if item.StatBonus.FD != 0 {
			bonuses = append(bonuses, fmt.Sprintf("FD %+d", item.StatBonus.FD))
		}
		if item.StatBonus.NICE != 0 {
			bonuses = append(bonuses, fmt.Sprintf("NICE %+d", item.StatBonus.NICE))
		}
		if item.StatBonus.UID != 0 {
			bonuses = append(bonuses, fmt.Sprintf("UID %+d", item.StatBonus.UID))
		}
		for i, b := range bonuses {
			if i > 0 {
				details += ", "
			}
			details += b
		}
	}

	m.statusMsg = details
}

// updatePause handles pause menu input.
func (m *Model) updatePause(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "p":
		m.currentView = ViewGame
	case "q":
		// Save before returning to main menu
		if m.engine != nil {
			m.engine.Shutdown()
			m.engine = nil
		}
		m.player = nil
		m.currentView = ViewMainMenu
		m.gameState = types.StateMainMenu
		m.statusMsg = "Game saved."
	}
	return m, nil
}

// updateGameOver handles game over screen input.
func (m *Model) updateGameOver(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		m.currentView = ViewMainMenu
		m.gameState = types.StateMainMenu
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

// updateHelp handles help screen input.
func (m *Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?", "q", "enter", " ":
		m.currentView = m.prevView
	}
	return m, nil
}

// updateMessageHistory handles message history view input.
func (m *Model) updateMessageHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	historyLen := len(m.messageHistory)
	maxScroll := historyLen - 20
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch msg.String() {
	case "esc", "m", "q", "enter", " ":
		m.currentView = m.prevView
	case "up", "k":
		m.messageScrollIdx++
		if m.messageScrollIdx > maxScroll {
			m.messageScrollIdx = maxScroll
		}
	case "down", "j":
		m.messageScrollIdx--
		if m.messageScrollIdx < 0 {
			m.messageScrollIdx = 0
		}
	case "pgup":
		m.messageScrollIdx += 10
		if m.messageScrollIdx > maxScroll {
			m.messageScrollIdx = maxScroll
		}
	case "pgdown":
		m.messageScrollIdx -= 10
		if m.messageScrollIdx < 0 {
			m.messageScrollIdx = 0
		}
	case "home":
		m.messageScrollIdx = maxScroll
	case "end":
		m.messageScrollIdx = 0
	}
	return m, nil
}

// updateIntro handles intro sequence input.
func (m *Model) updateIntro(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ", "esc", "q":
		// Skip intro, start game immediately
		m.introSkipped = true
		m.startNewGame(m.pendingClass)
		return m, nil
	case "right", "d", "l":
		// Advance to next frame manually
		m.introFrame++
		if m.introFrame >= len(introFrames) {
			m.startNewGame(m.pendingClass)
			return m, nil
		}
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return introTickMsg{}
		})
	}
	return m, nil
}

// openShop initializes and opens the shop.
func (m *Model) openShop() {
	m.shopCursor = 0
	m.shopItems = m.generateShopInventory()
	m.prevView = ViewGame
	m.currentView = ViewShop
}

// generateShopInventory creates the shop's item list based on current depth.
func (m *Model) generateShopInventory() []ShopItem {
	depth := 1
	if m.engine != nil {
		depth = m.engine.CurrentDepth()
	}

	// Base items always available
	items := []ShopItem{
		{TemplateID: "malloc", Name: "malloc()", Price: 10, InStock: true},
		{TemplateID: "fd_restore", Name: "close()", Price: 10, InStock: true},
		{TemplateID: "realloc", Name: "realloc()", Price: 25, InStock: true},
	}

	// Add gear based on depth
	if depth >= 2 {
		items = append(items, ShopItem{TemplateID: "basic_script", Name: "bash script", Price: 30, InStock: true})
		items = append(items, ShopItem{TemplateID: "basic_shell", Name: "/bin/sh", Price: 30, InStock: true})
		items = append(items, ShopItem{TemplateID: "env_vars", Name: "$PATH", Price: 20, InStock: true})
	}
	if depth >= 3 {
		items = append(items, ShopItem{TemplateID: "pipe_wrench", Name: "pipe |", Price: 50, InStock: true})
		items = append(items, ShopItem{TemplateID: "firewall", Name: "iptables", Price: 60, InStock: true})
		items = append(items, ShopItem{TemplateID: "ssh_key", Name: "id_rsa", Price: 40, InStock: true})
	}
	if depth >= 5 {
		items = append(items, ShopItem{TemplateID: "vim_blade", Name: ":wq!", Price: 100, InStock: true})
		items = append(items, ShopItem{TemplateID: "selinux_shield", Name: "SELinux", Price: 120, InStock: true})
		items = append(items, ShopItem{TemplateID: "sudo_potion", Name: "sudo potion", Price: 80, InStock: true})
	}
	if depth >= 7 {
		items = append(items, ShopItem{TemplateID: "kill_9", Name: "kill -9", Price: 150, InStock: true})
		items = append(items, ShopItem{TemplateID: "mmap", Name: "mmap()", Price: 100, InStock: true})
	}

	return items
}

// updateShop handles shop input.
func (m *Model) updateShop(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "$", "q":
		m.currentView = m.prevView
	case "up", "k", "w":
		m.shopCursor--
		if m.shopCursor < 0 {
			m.shopCursor = len(m.shopItems) - 1
		}
	case "down", "j", "s":
		m.shopCursor++
		if m.shopCursor >= len(m.shopItems) {
			m.shopCursor = 0
		}
	case "enter", " ":
		m.buyItem()
	}
	return m, nil
}

// buyItem attempts to purchase the selected shop item.
func (m *Model) buyItem() {
	if m.player == nil || m.shopCursor >= len(m.shopItems) {
		return
	}

	item := &m.shopItems[m.shopCursor]
	if !item.InStock {
		m.statusMsg = "Item out of stock!"
		return
	}

	if m.player.ExitCodes < item.Price {
		m.statusMsg = fmt.Sprintf("Not enough exit codes! Need %d, have %d.", item.Price, m.player.ExitCodes)
		return
	}

	// Create the item
	newItem := entity.NewItem(item.TemplateID, fmt.Sprintf("shop_%s_%d", item.TemplateID, m.player.ExitCodes), types.Position{})
	if newItem == nil {
		m.statusMsg = "Error creating item."
		return
	}

	// Try to add to inventory
	if !m.player.Inventory.AddItem(newItem) {
		m.statusMsg = "Inventory full!"
		return
	}

	// Deduct cost
	m.player.ExitCodes -= item.Price
	item.InStock = false // Sold out
	m.statusMsg = fmt.Sprintf("Purchased %s for %d exit codes!", item.Name, item.Price)
}

// updateAdmin handles admin console input.
func (m *Model) updateAdmin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "w":
		m.adminCursor--
		if m.adminCursor < 0 {
			m.adminCursor = len(m.adminOptions) - 1
		}
	case "down", "j", "s":
		m.adminCursor++
		if m.adminCursor >= len(m.adminOptions) {
			m.adminCursor = 0
		}
	case "enter", " ":
		return m.executeAdminAction()
	case "esc", "`", "q":
		m.currentView = m.prevView
	}
	return m, nil
}

// executeAdminAction executes the selected admin command.
func (m *Model) executeAdminAction() (tea.Model, tea.Cmd) {
	if m.player == nil || m.engine == nil {
		m.statusMsg = "No active game."
		m.currentView = m.prevView
		return m, nil
	}

	switch m.adminOptions[m.adminCursor] {
	case "Toggle God Mode":
		m.godMode = !m.godMode
		if m.godMode {
			m.statusMsg = "[ADMIN] God mode ENABLED - invincible"
		} else {
			m.statusMsg = "[ADMIN] God mode DISABLED"
		}

	case "Full Heal":
		m.player.Stats.RAM = m.player.MaxStats.MaxRAM
		m.player.Stats.FD = m.player.MaxStats.MaxFD
		m.statusMsg = "[ADMIN] Fully healed (RAM + FD restored)"

	case "Give XP (+100)":
		if m.player.GainXP(100) {
			m.statusMsg = fmt.Sprintf("[ADMIN] +100 XP - LEVEL UP! Now level %d", m.player.Level)
		} else {
			m.statusMsg = fmt.Sprintf("[ADMIN] +100 XP (%d/%d)", m.player.XP, m.player.XPToLevel)
		}

	case "Spawn Item":
		// Spawn a random useful item
		items := []string{"malloc", "kill_9", "chmod_x", "sudo_potion", "core_dump"}
		itemID := items[m.engine.MasterSeed()%int64(len(items))]
		item := entity.NewItem(itemID, fmt.Sprintf("admin_%d", m.engine.MasterSeed()), m.player.Position())
		if item != nil {
			if m.player.Inventory.AddItem(item) {
				m.statusMsg = fmt.Sprintf("[ADMIN] Spawned %s in inventory", item.Name())
			} else {
				m.statusMsg = "[ADMIN] Inventory full!"
			}
		}

	case "Teleport to Stairs":
		// Find stairs down position
		tiles := m.engine.GetVisibleTiles()
		for y, row := range tiles {
			for x, tile := range row {
				if tile.Type == types.TileStairsDown {
					m.player.SetPosition(types.Position{X: x, Y: y})
					m.statusMsg = "[ADMIN] Teleported to stairs"
					m.currentView = m.prevView
					return m, nil
				}
			}
		}
		m.statusMsg = "[ADMIN] No stairs found on this floor"

	case "Skip to Next Floor":
		if err := m.engine.DescendStairs(); err != nil {
			// Force descent even if not on stairs
			m.engine.ForceDescend()
			m.statusMsg = fmt.Sprintf("[ADMIN] Forced descent to depth %d", m.engine.CurrentDepth())
		} else {
			m.statusMsg = fmt.Sprintf("[ADMIN] Descended to depth %d", m.engine.CurrentDepth())
		}

	case "Show Debug Info":
		pos := m.player.Position()
		m.statusMsg = fmt.Sprintf("[DEBUG] Pos: (%d,%d) Floor: %s Depth: %d Seed: %d God: %v",
			pos.X, pos.Y,
			m.engine.CurrentFloorType().FloorName(),
			m.engine.CurrentDepth(),
			m.engine.MasterSeed(),
			m.godMode)

	case "Close":
		m.currentView = m.prevView
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	switch m.currentView {
	case ViewMainMenu:
		return m.viewMainMenu()
	case ViewClassSelect:
		return m.viewClassSelect()
	case ViewGame:
		return m.viewGame()
	case ViewCombat:
		return m.viewCombat()
	case ViewInventory:
		return m.viewInventory()
	case ViewPause:
		return m.viewPause()
	case ViewGameOver:
		return m.viewGameOver()
	case ViewVictory:
		return m.viewVictory()
	case ViewAdmin:
		return m.viewAdmin()
	case ViewHelp:
		return m.viewHelp()
	case ViewMessageHistory:
		return m.viewMessageHistory()
	case ViewIntro:
		return m.viewIntro()
	case ViewShop:
		return m.viewShop()
	default:
		return "Unknown view"
	}
}

// getClassDescription returns a description and stats preview for a class.
func (m *Model) getClassDescription(class entity.PlayerClass) (string, string) {
	switch class {
	case entity.ClassInit:
		return "The first process. Balanced starter.",
			"RAM: 100  CPU: 10  FD: 16  NICE: 10\nSkill: fork() - spawn child process"
	case entity.ClassCron:
		return "Scheduler daemon. Fast and precise.",
			"RAM: 100  CPU: 8   FD: 16  NICE: 5 (fast!)\nSkill: crontab - schedule 2x damage"
	case entity.ClassBash:
		return "Powerful shell. High attack output.",
			"RAM: 100  CPU: 15  FD: 12  NICE: 10\nSkill: pipe | - chain attacks"
	case entity.ClassVim:
		return "Complex editor. Many abilities.",
			"RAM: 100  CPU: 8   FD: 24  NICE: 10\nSkill: :normal - macro replay attack"
	case entity.ClassSudo:
		return "Root access. High risk, high power.",
			"RAM: 80   CPU: 10  FD: 16  UID: 0 (root!)\nSkill: sudo !! - bypass all defenses"
	default:
		return "Unknown class.", ""
	}
}

// viewClassSelect renders the class selection screen.
func (m *Model) viewClassSelect() string {
	title := m.styles.Title.Render(`
    ╔═══════════════════════════════════════════╗
    ║           SELECT YOUR PROCESS             ║
    ╚═══════════════════════════════════════════╝
	`)

	var menu string
	for i, class := range m.classOptions {
		cursor := "  "
		style := m.styles.MenuItem
		if i == m.classCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}
		menu += style.Render(fmt.Sprintf("%s%s", cursor, class)) + "\n"
	}

	// Show description of selected class
	desc, stats := m.getClassDescription(m.classOptions[m.classCursor])
	details := "\n" + m.styles.Muted.Render("─── Class Info ───") + "\n"
	details += m.styles.Normal.Render(desc) + "\n\n"
	details += m.styles.Highlight.Render(stats) + "\n"

	footer := m.styles.Muted.Render("\n[↑/↓] Navigate  [Enter] Select  [Esc] Back")

	return m.styles.Container.Render(title + "\n" + menu + details + footer)
}

// viewMainMenu renders the main menu.
func (m *Model) viewMainMenu() string {
	title := m.styles.Title.Render(`
    ╔═══════════════════════════════════════════╗
    ║         /dev/dungeon                      ║
    ║    Navigate the filesystem. Survive.      ║
    ╚═══════════════════════════════════════════╝
	`)

	var menu string
	for i, option := range m.menuOptions {
		cursor := "  "
		style := m.styles.MenuItem
		if i == m.menuCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}
		menu += style.Render(cursor+option) + "\n"
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Navigate  [Enter] Select  [q] Quit")

	if m.statusMsg != "" {
		footer = m.styles.Danger.Render(m.statusMsg) + "\n" + footer
	}

	return m.styles.Container.Render(title + "\n\n" + menu + "\n" + footer)
}

// viewGame renders the main game view.
func (m *Model) viewGame() string {
	// Get viewport dimensions for consistent sizing
	vpWidth, _ := m.getViewportSize()

	// Stats panel
	stats := m.renderStats()

	// Map
	mapView := m.renderMap()

	// Log/status - match map width
	log := m.renderLog(vpWidth)

	// Layout: map on left, stats on right, log at bottom spanning full width
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, mapView, "  ", stats)

	return m.styles.Container.Render(topRow + "\n" + log)
}

// renderStats renders the stats panel.
func (m *Model) renderStats() string {
	if m.player == nil {
		return ""
	}

	p := m.player

	// Get floor info from engine
	floorName := "/home"
	floorDepth := 1
	if m.engine != nil {
		floorName = m.engine.CurrentFloorType().FloorName()
		floorDepth = m.engine.CurrentDepth()
	}

	content := fmt.Sprintf(
		"%s\n%s\n\n"+
			"RAM: %s/%d\n"+
			"CPU: %d\n"+
			"FD:  %s/%d\n"+
			"NICE: %d\n"+
			"UID: %d\n\n"+
			"Level: %d\n"+
			"XP: %d/%d\n"+
			"Floor: %s\n"+
			"Depth: %d",
		m.styles.Title.Render(string(p.Class)),
		m.styles.Muted.Render("Process Status"),
		m.colorizeRAM(p.Stats.RAM, p.MaxStats.MaxRAM),
		p.MaxStats.MaxRAM,
		p.Stats.CPU,
		m.colorizeFD(p.Stats.FD, p.MaxStats.MaxFD),
		p.MaxStats.MaxFD,
		p.Stats.NICE,
		p.Stats.UID,
		p.Level,
		p.XP,
		p.XPToLevel,
		floorName,
		floorDepth,
	)

	return m.styles.StatPanel.Width(20).Render(content)
}

// colorizeRAM colors RAM (health) based on percentage.
func (m *Model) colorizeRAM(current, max int) string {
	pct := float64(current) / float64(max)
	str := fmt.Sprintf("%d", current)
	if pct > 0.6 {
		return m.styles.Success.Render(str)
	} else if pct > 0.3 {
		return m.styles.Highlight.Render(str)
	}
	return m.styles.Danger.Render(str)
}

// colorizeFD colors FD (ability resource) based on percentage.
func (m *Model) colorizeFD(current, max int) string {
	pct := float64(current) / float64(max)
	str := fmt.Sprintf("%d", current)
	if pct > 0.5 {
		return m.styles.Normal.Render(str)
	}
	return m.styles.Muted.Render(str)
}

// getViewportSize calculates the map viewport size based on terminal dimensions.
func (m *Model) getViewportSize() (width, height int) {
	termWidth := m.width
	termHeight := m.height

	// Use sensible defaults if terminal size not yet received
	if termWidth == 0 {
		termWidth = 120
	}
	if termHeight == 0 {
		termHeight = 40
	}

	// Account for UI chrome: stats panel (26 chars), borders, padding, gaps
	statsPanelWidth := 26
	borderAndPadding := 10 // map border (2) + container padding (4) + gap (2) + some buffer

	// Calculate available space for map
	width = termWidth - statsPanelWidth - borderAndPadding
	height = termHeight - 8 // Leave room for log panel (4 lines) and container padding

	// Minimum bounds
	if width < 40 {
		width = 40
	}
	if height < 15 {
		height = 15
	}

	return width, height
}

// renderMap renders the game map from the engine with a viewport centered on the player.
func (m *Model) renderMap() string {
	if m.engine == nil {
		return m.styles.MapBorder.Render("No map loaded")
	}

	tiles := m.engine.GetVisibleTiles()
	if tiles == nil || len(tiles) == 0 {
		return m.styles.MapBorder.Render("No map loaded")
	}

	dungeonHeight := len(tiles)
	dungeonWidth := 0
	if dungeonHeight > 0 {
		dungeonWidth = len(tiles[0])
	}

	playerPos := types.Position{}
	if m.player != nil {
		playerPos = m.player.Position()
	}

	// Calculate viewport size based on terminal dimensions
	vpWidth, vpHeight := m.getViewportSize()

	// Calculate viewport origin (top-left corner), centered on player
	vpX := playerPos.X - vpWidth/2
	vpY := playerPos.Y - vpHeight/2

	// Clamp viewport to dungeon bounds
	if vpX < 0 {
		vpX = 0
	}
	if vpY < 0 {
		vpY = 0
	}
	if vpX+vpWidth > dungeonWidth {
		vpX = dungeonWidth - vpWidth
		if vpX < 0 {
			vpX = 0
		}
	}
	if vpY+vpHeight > dungeonHeight {
		vpY = dungeonHeight - vpHeight
		if vpY < 0 {
			vpY = 0
		}
	}

	// Get enemies and items for rendering
	enemies := m.engine.GetEnemies()
	items := m.engine.GetItems()

	var mapStr string
	for vy := 0; vy < vpHeight; vy++ {
		y := vpY + vy
		if y >= dungeonHeight {
			break
		}

		for vx := 0; vx < vpWidth; vx++ {
			x := vpX + vx
			if x >= dungeonWidth {
				mapStr += " "
				continue
			}

			pos := types.Position{X: x, Y: y}
			tile := tiles[y][x]

			// Player takes priority
			if pos == playerPos {
				mapStr += m.styles.Player.Render("@")
				continue
			}

			// Check for enemy at position
			enemyFound := false
			for _, enemy := range enemies {
				if enemy.Position() == pos && tile.Visible {
					mapStr += m.styles.Enemy.Render(string(enemy.Glyph()))
					enemyFound = true
					break
				}
			}
			if enemyFound {
				continue
			}

			// Check for item at position
			itemFound := false
			for _, item := range items {
				if item.Position() == pos && tile.Visible {
					mapStr += m.styles.Item.Render(string(item.Glyph()))
					itemFound = true
					break
				}
			}
			if itemFound {
				continue
			}

			// Render tile based on visibility
			if !tile.Explored && !tile.Visible {
				mapStr += " " // Unexplored
			} else if tile.Visible {
				// Fully visible tiles
				mapStr += m.renderTile(tile.Type)
			} else {
				// Explored but not visible (darker)
				mapStr += m.styles.Muted.Render(m.getTileGlyph(tile.Type))
			}
		}
		mapStr += "\n"
	}

	return m.styles.MapBorder.Render(mapStr)
}

// renderTile returns a styled string for a tile type.
func (m *Model) renderTile(tileType types.TileType) string {
	switch tileType {
	case types.TileWall:
		return m.styles.Wall.Render("#")
	case types.TileFloor:
		return m.styles.Floor.Render(".")
	case types.TileStairsUp:
		return m.styles.Stairs.Render("<")
	case types.TileStairsDown:
		return m.styles.Stairs.Render(">")
	case types.TileDoor:
		return m.styles.Highlight.Render("+")
	default:
		return " "
	}
}

// getTileGlyph returns the character for a tile type (for explored-but-not-visible).
func (m *Model) getTileGlyph(tileType types.TileType) string {
	switch tileType {
	case types.TileWall:
		return "#"
	case types.TileFloor:
		return "."
	case types.TileStairsUp:
		return "<"
	case types.TileStairsDown:
		return ">"
	case types.TileDoor:
		return "+"
	default:
		return " "
	}
}

// renderLog renders the message log with dynamic width.
func (m *Model) renderLog(width int) string {
	content := m.statusMsg
	if content == "" {
		content = "Ready."
	}

	footer := m.styles.Muted.Render("[WASD/hjkl] Move  [</>] Stairs  [I] Inv  [M] Log  [P] Pause  [?] Help")

	// Add 2 for border padding
	logWidth := width + 2
	if logWidth < 60 {
		logWidth = 60
	}

	return m.styles.LogPanel.Width(logWidth).Render(content + "\n" + footer)
}

// viewCombat renders the combat view.
func (m *Model) viewCombat() string {
	title := m.styles.Danger.Render("═══ COMBAT ═══") + "\n\n"

	// Show player stats
	playerInfo := ""
	if m.player != nil {
		playerInfo = fmt.Sprintf("%s %s\n",
			m.styles.Player.Render("@"),
			m.styles.Title.Render(string(m.player.Class)))
		playerInfo += fmt.Sprintf("RAM: %s/%d  FD: %d/%d\n",
			m.colorizeRAM(m.player.Stats.RAM, m.player.MaxStats.MaxRAM),
			m.player.MaxStats.MaxRAM,
			m.player.Stats.FD,
			m.player.MaxStats.MaxFD)

		// Show active buffs
		if len(m.player.ActiveBuffs) > 0 {
			buffStr := "Buffs: "
			for i, buff := range m.player.ActiveBuffs {
				if i > 0 {
					buffStr += ", "
				}
				var buffColor string
				switch buff.Type {
				case entity.BuffInvincible:
					buffColor = m.styles.Highlight.Render(fmt.Sprintf("★%s(%d)", buff.Name, buff.Duration))
				case entity.BuffStrength:
					buffColor = m.styles.Danger.Render(fmt.Sprintf("⚔%s(%d)", buff.Name, buff.Duration))
				case entity.BuffHaste:
					buffColor = m.styles.Muted.Render(fmt.Sprintf("»%s(%d)", buff.Name, buff.Duration))
				default:
					buffColor = fmt.Sprintf("%s(%d)", buff.Name, buff.Duration)
				}
				buffStr += buffColor
			}
			playerInfo += buffStr + "\n"
		}
		playerInfo += "\n"
	}

	// Show enemy info with target indicator
	var enemyInfo string
	if m.combat != nil {
		aliveEnemies := m.combat.GetAliveEnemies()
		for i, enemy := range aliveEnemies {
			// Show target indicator
			targetIndicator := "  "
			if i == m.targetCursor {
				targetIndicator = m.styles.Highlight.Render("► ")
			}

			// Show boss indicator
			bossTag := ""
			if enemy.IsBoss {
				bossTag = m.styles.Danger.Render(" [BOSS]")
			}

			enemyInfo += fmt.Sprintf("%s%s %-14s RAM: %d/%d  CPU: %d%s\n",
				targetIndicator,
				m.styles.Enemy.Render(string(enemy.Glyph())),
				enemy.Name(),
				enemy.Stats.RAM,
				enemy.MaxStats.MaxRAM,
				enemy.Stats.CPU,
				bossTag)
		}
		if enemyInfo != "" {
			header := "─── Enemies "
			if len(aliveEnemies) > 1 {
				header += "(←/→ to target) "
			}
			header += "───"
			enemyInfo = m.styles.Muted.Render(header) + "\n" + enemyInfo + "\n"
		}
	}

	var menu string
	var footer string

	if m.selectingSkill && m.player != nil {
		// Skill selection mode
		menu = m.styles.Highlight.Render("─── Select Skill ───") + "\n"
		for i, skill := range m.player.Skills {
			// Show cooldown and FD cost
			cdStr := ""
			if skill.CurrentCD > 0 {
				cdStr = m.styles.Danger.Render(fmt.Sprintf(" [CD:%d]", skill.CurrentCD))
			}
			fdStr := ""
			if skill.FDCost > 0 {
				fdStr = m.styles.Muted.Render(fmt.Sprintf(" (FD:%d)", skill.FDCost))
			}

			skillStr := fmt.Sprintf("[%d] %s%s%s", i+1, skill.Name, fdStr, cdStr)

			if i == m.skillCursor {
				menu += m.styles.MenuSelected.Render("  > "+skillStr) + "\n"
				// Show skill description
				menu += m.styles.Muted.Render("      "+skill.Description) + "\n"
			} else {
				menu += m.styles.MenuItem.Render("    "+skillStr) + "\n"
			}
		}
		footer = m.styles.Muted.Render("\n[↑/↓] Select  [Enter/1-5] Use  [Esc] Cancel")
	} else {
		// Normal combat menu
		options := []string{
			"[1] Attack (kill -TERM)",
			"[2] Hack (use skill)",
			"[3] Use Item",
			"[4] Flee",
		}
		menu = m.styles.Muted.Render("─── Actions ───") + "\n"
		for i, opt := range options {
			if i == m.combatCursor {
				menu += m.styles.MenuSelected.Render("  > "+opt) + "\n"
			} else {
				menu += m.styles.MenuItem.Render("    "+opt) + "\n"
			}
		}
		footer = m.styles.Muted.Render("\n[↑/↓] Select  [Enter/1-4] Act")
	}

	// Combat log
	var log string
	if len(m.combatLog) > 0 {
		log = "\n" + m.styles.Muted.Render("─── Combat Log ───") + "\n"
		start := 0
		if len(m.combatLog) > 6 {
			start = len(m.combatLog) - 6
		}
		for _, entry := range m.combatLog[start:] {
			log += m.styles.Normal.Render("  "+entry) + "\n"
		}
	}

	return m.styles.Container.Render(title + playerInfo + enemyInfo + menu + log + footer)
}

// viewInventory renders the inventory view.
func (m *Model) viewInventory() string {
	if m.player == nil {
		return ""
	}

	title := m.styles.Title.Render("═══ INVENTORY ═══")
	invLen := len(m.player.Inventory.Items)

	var items string
	var selectedDesc string

	if invLen == 0 {
		items = m.styles.Muted.Render("  (empty)\n")
	} else {
		for i, item := range m.player.Inventory.Items {
			cursor := "  "
			style := m.styles.MenuItem
			if i == m.invCursor {
				cursor = "> "
				style = m.styles.MenuSelected
				selectedDesc = m.getItemDetails(item)
			}
			itemStr := fmt.Sprintf("%s%c %s", cursor, item.Glyph(), item.Name())
			if item.Stackable && item.Quantity > 1 {
				itemStr += fmt.Sprintf(" x%d", item.Quantity)
			}
			// Show item type tag
			itemStr += m.styles.Muted.Render(fmt.Sprintf(" [%s]", item.ItemType))
			items += style.Render(itemStr) + "\n"
		}
	}

	// Equipment (selectable)
	equipment := m.styles.Title.Render("\n═══ EQUIPPED ═══\n")
	eq := m.player.Equipment

	equipSlots := []struct {
		name string
		item *entity.Item
	}{
		{"Weapon", eq.Weapon},
		{"Armor", eq.Armor},
		{"Util 1", eq.Utility1},
		{"Util 2", eq.Utility2},
	}

	for i, slot := range equipSlots {
		slotIdx := invLen + i
		cursor := "  "
		style := m.styles.MenuItem
		if slotIdx == m.invCursor {
			cursor = "> "
			style = m.styles.MenuSelected
			if slot.item != nil {
				selectedDesc = m.getItemDetails(slot.item)
			}
		}
		equipment += style.Render(fmt.Sprintf("%s%-7s %s", cursor, slot.name+":", m.equipmentSlotDisplay(slot.item))) + "\n"
	}

	// Details panel
	detailsPanel := ""
	if selectedDesc != "" {
		detailsPanel = "\n" + m.styles.Muted.Render("─── Details [?] ───\n") + selectedDesc
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Select  [Enter] Use/Equip  [U] Unequip  [D] Drop  [?] Info  [I/Esc] Close")

	return m.styles.Container.Render(title + "\n" + items + equipment + detailsPanel + footer)
}

// getItemDetails returns formatted item details.
func (m *Model) getItemDetails(item *entity.Item) string {
	if item == nil {
		return ""
	}
	details := m.styles.Normal.Render(item.Description) + "\n"
	statStr := m.formatStatBonus(item)
	if statStr != "" {
		details += m.styles.Highlight.Render(statStr) + "\n"
	}
	return details
}

// equipmentSlotDisplay formats an equipment slot for display.
func (m *Model) equipmentSlotDisplay(item *entity.Item) string {
	if item == nil {
		return m.styles.Muted.Render("(empty)")
	}
	return fmt.Sprintf("%c %s", item.Glyph(), item.Name())
}

// formatStatBonus formats stat bonuses for display.
func (m *Model) formatStatBonus(item *entity.Item) string {
	var bonuses []string
	if item.StatBonus.CPU != 0 {
		bonuses = append(bonuses, fmt.Sprintf("CPU %+d", item.StatBonus.CPU))
	}
	if item.StatBonus.RAM != 0 {
		bonuses = append(bonuses, fmt.Sprintf("RAM %+d", item.StatBonus.RAM))
	}
	if item.StatBonus.FD != 0 {
		bonuses = append(bonuses, fmt.Sprintf("FD %+d", item.StatBonus.FD))
	}
	if item.StatBonus.UID != 0 {
		bonuses = append(bonuses, fmt.Sprintf("UID %+d", item.StatBonus.UID))
	}
	if len(bonuses) == 0 {
		return ""
	}
	return "Stats: " + fmt.Sprintf("%v", bonuses)
}

// equipmentSlot formats an equipment slot.
func (m *Model) equipmentSlot(item *entity.Item) string {
	if item == nil {
		return m.styles.Muted.Render("(empty)")
	}
	return fmt.Sprintf("%c %s", item.Glyph(), item.Name())
}

// viewPause renders the pause menu.
func (m *Model) viewPause() string {
	content := m.styles.Title.Render("═══ PAUSED ═══\n\n")
	content += "[P/Esc] Resume\n"
	content += "[Q] Quit to Menu\n"

	return m.styles.Container.Render(content)
}

// viewGameOver renders the game over screen.
func (m *Model) viewGameOver() string {
	content := m.styles.Danger.Render(`
    ╔═══════════════════════════════════════════╗
    ║              PROCESS TERMINATED           ║
    ║                  exit(1)                  ║
    ╚═══════════════════════════════════════════╝
	`) + "\n"

	// Show run statistics
	if m.engine != nil {
		stats := m.engine.GetRunStats()
		if stats != nil {
			content += m.styles.Title.Render("─── Run Statistics ───") + "\n"
			content += fmt.Sprintf("  Total Kills:     %d\n", stats.TotalKills)
			content += fmt.Sprintf("  Max Depth:       %d\n", stats.MaxDepthReached)
			content += fmt.Sprintf("  Floors Explored: %d\n", stats.FloorsExplored)
			content += fmt.Sprintf("  Steps Walked:    %d\n", stats.StepsWalked)
			content += fmt.Sprintf("  Items Collected: %d\n", stats.ItemsCollected)

			// Show kill breakdown if any kills
			if stats.TotalKills > 0 && len(stats.EnemiesKilled) > 0 {
				content += "\n" + m.styles.Muted.Render("  Kill Log:") + "\n"
				for enemyType, count := range stats.EnemiesKilled {
					content += fmt.Sprintf("    %s: %d\n", enemyType, count)
				}
			}
		}
	}

	if m.player != nil {
		content += fmt.Sprintf("\nLevel Reached: %d\n", m.player.Level)
	}

	content += m.styles.Muted.Render("\n[Enter] Continue  [Q] Quit")

	return m.styles.Container.Render(content)
}

// viewVictory renders the victory screen.
func (m *Model) viewVictory() string {
	content := m.styles.Success.Render(`
    ╔═══════════════════════════════════════════╗
    ║              KERNEL DEFEATED              ║
    ║                  exit(0)                  ║
    ╚═══════════════════════════════════════════╝
	`) + "\n"

	content += m.styles.Title.Render("You have conquered /dev/dungeon!") + "\n\n"

	// Show run statistics
	if m.engine != nil {
		stats := m.engine.GetRunStats()
		if stats != nil {
			content += m.styles.Highlight.Render("─── Final Statistics ───") + "\n"
			content += fmt.Sprintf("  Total Kills:     %d\n", stats.TotalKills)
			content += fmt.Sprintf("  Max Depth:       %d\n", stats.MaxDepthReached)
			content += fmt.Sprintf("  Floors Explored: %d\n", stats.FloorsExplored)
			content += fmt.Sprintf("  Steps Walked:    %d\n", stats.StepsWalked)
			content += fmt.Sprintf("  Items Collected: %d\n", stats.ItemsCollected)
		}
	}

	if m.player != nil {
		content += fmt.Sprintf("\nFinal Level: %d\n", m.player.Level)
	}

	content += m.styles.Muted.Render("\n[Enter] Continue  [Q] Quit")

	return m.styles.Container.Render(content)
}

// viewAdmin renders the admin console.
func (m *Model) viewAdmin() string {
	title := m.styles.Danger.Render("═══ ADMIN CONSOLE ═══") + "\n"
	title += m.styles.Muted.Render("(debug commands - use at your own risk)") + "\n\n"

	var menu string
	for i, opt := range m.adminOptions {
		cursor := "  "
		style := m.styles.MenuItem
		if i == m.adminCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}
		menu += style.Render(cursor+opt) + "\n"
	}

	// Show current status
	status := "\n" + m.styles.Muted.Render("─── Status ───") + "\n"
	if m.godMode {
		status += m.styles.Success.Render("  God Mode: ENABLED") + "\n"
	} else {
		status += m.styles.Normal.Render("  God Mode: disabled") + "\n"
	}
	if m.player != nil {
		status += fmt.Sprintf("  RAM: %d/%d  Level: %d\n", m.player.Stats.RAM, m.player.MaxStats.MaxRAM, m.player.Level)
	}
	if m.engine != nil {
		status += fmt.Sprintf("  Depth: %d  Floor: %s\n", m.engine.CurrentDepth(), m.engine.CurrentFloorType().FloorName())
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Navigate  [Enter] Execute  [Esc/`] Close")

	return m.styles.Container.Render(title + menu + status + footer)
}

// viewHelp renders the help/keybindings screen.
func (m *Model) viewHelp() string {
	title := m.styles.Title.Render("═══ HELP / KEYBINDINGS ═══") + "\n\n"

	movement := m.styles.Highlight.Render("Movement:") + "\n"
	movement += "  WASD / Arrow Keys / hjkl  - Move\n"
	movement += "  > or .                    - Descend stairs\n"
	movement += "  < or ,                    - Ascend stairs\n\n"

	actions := m.styles.Highlight.Render("Actions:") + "\n"
	actions += "  I                         - Open inventory\n"
	actions += "  P or Esc                  - Pause menu\n"
	actions += "  Q                         - Save & quit to menu\n"
	actions += "  ?                         - This help screen\n\n"

	combat := m.styles.Highlight.Render("Combat:") + "\n"
	combat += "  1 or Enter                - Attack (kill -TERM)\n"
	combat += "  2                         - Hack (use skill)\n"
	combat += "  3                         - Use item\n"
	combat += "  4                         - Attempt to flee\n\n"

	inventory := m.styles.Highlight.Render("Inventory:") + "\n"
	inventory += "  Enter / Space            - Use or equip item\n"
	inventory += "  E                        - Equip item\n"
	inventory += "  D                        - Drop item\n\n"

	stats := m.styles.Highlight.Render("Stats:") + "\n"
	stats += "  RAM   - Health (memory). Reach 0 = OOM killed\n"
	stats += "  CPU   - Attack power\n"
	stats += "  FD    - File descriptors for skills\n"
	stats += "  NICE  - Speed (lower = faster, more crits)\n"
	stats += "  UID   - Access level (0 = root = admin access)\n\n"

	tips := m.styles.Muted.Render("Tips:") + "\n"
	tips += "  - Walk into enemies to start combat\n"
	tips += "  - Items are auto-picked up when walking over them\n"
	tips += "  - sudo class starts with UID 0 (root access)\n"
	tips += "  - Find root_shard items to lower your UID\n"

	footer := m.styles.Muted.Render("\n[Esc/?/Enter] Close")

	return m.styles.Container.Render(title + movement + actions + combat + inventory + stats + tips + footer)
}

// viewMessageHistory renders the scrollable message history.
func (m *Model) viewMessageHistory() string {
	title := m.styles.Title.Render("═══ MESSAGE LOG ═══") + "\n\n"

	if len(m.messageHistory) == 0 {
		content := m.styles.Muted.Render("No messages yet.")
		footer := m.styles.Muted.Render("\n[Esc/M/Enter] Close")
		return m.styles.Container.Render(title + content + footer)
	}

	// Calculate visible range
	visibleLines := 20
	historyLen := len(m.messageHistory)

	// scrollIdx 0 = show most recent, higher = scroll back in time
	endIdx := historyLen - m.messageScrollIdx
	startIdx := endIdx - visibleLines
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > historyLen {
		endIdx = historyLen
	}

	var content string
	for i := startIdx; i < endIdx; i++ {
		// Show line numbers relative to total history
		lineNum := i + 1
		msg := m.messageHistory[i]
		content += m.styles.Muted.Render(fmt.Sprintf("%3d ", lineNum)) + msg + "\n"
	}

	// Show scroll position
	scrollInfo := fmt.Sprintf("\n─── Showing %d-%d of %d messages ───",
		startIdx+1, endIdx, historyLen)
	if m.messageScrollIdx > 0 {
		scrollInfo += m.styles.Muted.Render(" [↑ for older]")
	}
	if m.messageScrollIdx < historyLen-visibleLines {
		scrollInfo += m.styles.Muted.Render(" [↓ for newer]")
	}
	content += m.styles.Muted.Render(scrollInfo)

	footer := m.styles.Muted.Render("\n\n[↑/↓] Scroll  [PgUp/PgDn] Fast scroll  [Home/End] Jump  [Esc/M] Close")

	return m.styles.Container.Render(title + content + footer)
}

// viewIntro renders the animated intro sequence.
func (m *Model) viewIntro() string {
	if m.introFrame < 0 || m.introFrame >= len(introFrames) {
		return ""
	}

	frame := introFrames[m.introFrame]

	// Progress indicator
	progress := fmt.Sprintf("[%d/%d]", m.introFrame+1, len(introFrames))

	footer := m.styles.Muted.Render("\n\n" + progress + "   [Enter/Space] Skip   [→] Next")

	return m.styles.Container.Render(m.styles.Title.Render(frame) + footer)
}

// viewShop renders the shop interface styled like ls -la output.
func (m *Model) viewShop() string {
	title := m.styles.Title.Render("$ ls -la /dev/store") + "\n"
	title += m.styles.Muted.Render("total 42\n")
	title += m.styles.Muted.Render("drwxr-xr-x  2 root  shop  4096 Jan 13 04:20 .\n")
	title += m.styles.Muted.Render("drwxr-xr-x 10 root  root  4096 Jan 13 04:20 ..\n\n")

	// Show player's exit codes like a shell variable
	balance := m.styles.Muted.Render("$ echo $EXIT_CODES\n")
	balance += m.styles.Highlight.Render(fmt.Sprintf("%d", m.player.ExitCodes)) + "\n\n"

	var items string
	for i, item := range m.shopItems {
		cursor := " "
		style := m.styles.MenuItem
		if i == m.shopCursor {
			cursor = ">"
			style = m.styles.MenuSelected
		}

		// Format like ls -la output: permissions, owner, size (price), name
		perms := "-rw-r--r--"
		if !item.InStock {
			perms = "----------"
		}

		priceStr := fmt.Sprintf("%4d", item.Price)
		if !item.InStock {
			priceStr = m.styles.Muted.Render(priceStr)
		} else if m.player.ExitCodes < item.Price {
			priceStr = m.styles.Danger.Render(priceStr)
		} else {
			priceStr = m.styles.Success.Render(priceStr)
		}

		nameStr := item.Name
		if !item.InStock {
			nameStr = m.styles.Muted.Render(item.Name + " (SOLD)")
		}

		// ls -la format: perms links owner group size date name
		itemLine := fmt.Sprintf("%s %s 1 shop shop %s Jan 13 %s", cursor, perms, priceStr, nameStr)
		items += style.Render(itemLine) + "\n"
	}

	// Show selected item details
	details := ""
	if m.shopCursor < len(m.shopItems) {
		item := m.shopItems[m.shopCursor]
		template := entity.NewItem(item.TemplateID, "preview", types.Position{})
		if template != nil {
			details = "\n" + m.styles.Muted.Render("$ cat README."+item.TemplateID+"\n")
			details += m.styles.Normal.Render(template.Description) + "\n"
			statStr := m.formatStatBonus(template)
			if statStr != "" {
				details += m.styles.Highlight.Render(statStr) + "\n"
			}
		}
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Browse  [Enter] Buy  [$/Esc] Exit")

	return m.styles.Container.Render(title + balance + items + details + footer)
}
