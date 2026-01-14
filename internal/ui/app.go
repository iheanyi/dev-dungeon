// Package ui provides the Bubble Tea UI for /dev/dungeon.
package ui

import (
	"fmt"

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
	ViewGame
	ViewCombat
	ViewInventory
	ViewPause
	ViewGameOver
	ViewVictory
)

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
	combatCursor int
	combatLog    []string
	enemies      []*entity.Enemy

	// Inventory state
	invCursor    int

	// Messages
	statusMsg    string

	// Styles
	styles       *Styles
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
		combatLog:  make([]string, 0),
		styles:     NewStyles(),
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
		m.startNewGame()
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

// startNewGame initializes a new game.
func (m *Model) startNewGame() {
	// Create the game engine with seed 0 (random)
	m.engine = game.NewEngine(m.config, 0)

	// Set up the dungeon generator
	dungeonCfg := dungeon.DefaultConfig()
	dungeonCfg.Width = m.config.Display.MapWidth
	dungeonCfg.Height = m.config.Display.MapHeight
	m.engine.SetGenerator(dungeon.NewGenerator(dungeonCfg))

	// Start a new game with the configured class
	playerClass := entity.PlayerClass(m.config.Game.StartingClass)
	if err := m.engine.StartNewGame(playerClass); err != nil {
		m.statusMsg = fmt.Sprintf("Failed to start game: %v", err)
		return
	}

	// Get the player from the engine
	m.player = m.engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring
	m.statusMsg = fmt.Sprintf("Welcome to %s. Navigate the filesystem. Survive.", m.engine.CurrentFloorType().FloorName())
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
	}
	return m, nil
}

// movePlayer moves the player in a direction using the game engine.
func (m *Model) movePlayer(dir types.Direction) {
	if m.engine == nil || m.player == nil {
		return
	}

	result := m.engine.MovePlayer(dir)

	// Update status message
	if result.Message != "" {
		m.statusMsg = result.Message
	}

	// Check for combat initiation
	if result.Combat != nil {
		m.enemies = []*entity.Enemy{result.Combat}
		m.currentView = ViewCombat
		m.gameState = types.StateCombat
		m.combatCursor = 0
		m.combatLog = []string{fmt.Sprintf("A wild %s appears!", result.Combat.Name())}
	}
}

// updateCombat handles combat view input.
func (m *Model) updateCombat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "enter", " ", "1", "2", "3", "4":
		return m.executeCombatAction(msg.String())
	case "esc", "q":
		// Temporary: flee from combat (TODO: implement proper flee mechanic)
		m.currentView = ViewGame
		m.gameState = types.StateExploring
		m.statusMsg = "You fled from combat! (DEBUG: combat not yet implemented)"
	}
	return m, nil
}

// executeCombatAction executes the selected combat action.
func (m *Model) executeCombatAction(key string) (tea.Model, tea.Cmd) {
	// Map keys to actions
	actionIndex := m.combatCursor
	if key >= "1" && key <= "4" {
		actionIndex = int(key[0] - '1')
	}

	switch actionIndex {
	case 0: // Attack
		m.combatLog = append(m.combatLog, "You attack!")
	case 1: // Hack
		m.combatLog = append(m.combatLog, "You attempt to hack...")
	case 2: // Use Item
		m.currentView = ViewInventory
	case 3: // Flee
		m.combatLog = append(m.combatLog, "You attempt to flee...")
		// TODO: Flee logic
	}

	return m, nil
}

// updateInventory handles inventory view input.
func (m *Model) updateInventory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			m.invCursor = len(m.player.Inventory.Items) - 1
			if m.invCursor < 0 {
				m.invCursor = 0
			}
		}
	case "down", "j", "s":
		m.invCursor++
		if m.invCursor >= len(m.player.Inventory.Items) {
			m.invCursor = 0
		}
	case "enter", " ":
		// TODO: Use/equip item
	}
	return m, nil
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

// View implements tea.Model.
func (m *Model) View() string {
	switch m.currentView {
	case ViewMainMenu:
		return m.viewMainMenu()
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
	default:
		return "Unknown view"
	}
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

	footer := m.styles.Muted.Render("[WASD/hjkl] Move  [</>] Stairs  [I] Inventory  [P] Pause  [Q] Quit")

	// Add 2 for border padding
	logWidth := width + 2
	if logWidth < 60 {
		logWidth = 60
	}

	return m.styles.LogPanel.Width(logWidth).Render(content + "\n" + footer)
}

// viewCombat renders the combat view.
func (m *Model) viewCombat() string {
	title := m.styles.Danger.Render("═══ COMBAT ═══\n")

	// Show enemy info
	var enemyInfo string
	if len(m.enemies) > 0 {
		enemy := m.enemies[0]
		enemyInfo = fmt.Sprintf("\n%s  %s\n",
			m.styles.Enemy.Render(string(enemy.Glyph())),
			m.styles.Title.Render(enemy.Name()))
		enemyInfo += fmt.Sprintf("RAM: %d/%d\n", enemy.Stats.RAM, enemy.MaxStats.MaxRAM)
		enemyInfo += fmt.Sprintf("CPU: %d\n\n", enemy.Stats.CPU)
	}

	// Combat options
	options := []string{"[1] Attack (kill -TERM)", "[2] Hack", "[3] Use Item", "[4] Flee"}
	var menu string
	for i, opt := range options {
		if i == m.combatCursor {
			menu += m.styles.MenuSelected.Render("> "+opt) + "\n"
		} else {
			menu += m.styles.MenuItem.Render("  "+opt) + "\n"
		}
	}

	// Combat log
	var log string
	if len(m.combatLog) > 0 {
		log = "\n" + m.styles.Muted.Render("─── Log ───") + "\n"
		start := 0
		if len(m.combatLog) > 5 {
			start = len(m.combatLog) - 5
		}
		for _, entry := range m.combatLog[start:] {
			log += m.styles.Normal.Render("  "+entry) + "\n"
		}
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Select  [Enter/1-4] Act  [Esc] Flee (debug)")
	footer += m.styles.Danger.Render("\n\n⚠ Combat system not yet implemented - press Esc to flee")

	return m.styles.Container.Render(title + enemyInfo + menu + log + footer)
}

// viewInventory renders the inventory view.
func (m *Model) viewInventory() string {
	if m.player == nil {
		return ""
	}

	title := m.styles.Title.Render("═══ INVENTORY ═══")

	var items string
	if len(m.player.Inventory.Items) == 0 {
		items = m.styles.Muted.Render("  (empty)")
	} else {
		for i, item := range m.player.Inventory.Items {
			cursor := "  "
			style := m.styles.MenuItem
			if i == m.invCursor {
				cursor = "> "
				style = m.styles.MenuSelected
			}
			itemStr := fmt.Sprintf("%s%c %s", cursor, item.Glyph(), item.Name())
			if item.Stackable && item.Quantity > 1 {
				itemStr += fmt.Sprintf(" x%d", item.Quantity)
			}
			items += style.Render(itemStr) + "\n"
		}
	}

	// Equipment
	equipment := m.styles.Title.Render("\n═══ EQUIPMENT ═══\n")
	eq := m.player.Equipment
	equipment += fmt.Sprintf("Weapon: %s\n", m.equipmentSlot(eq.Weapon))
	equipment += fmt.Sprintf("Armor:  %s\n", m.equipmentSlot(eq.Armor))
	equipment += fmt.Sprintf("Util 1: %s\n", m.equipmentSlot(eq.Utility1))
	equipment += fmt.Sprintf("Util 2: %s\n", m.equipmentSlot(eq.Utility2))

	footer := m.styles.Muted.Render("\n[↑/↓] Navigate  [Enter] Use/Equip  [I/Esc] Close")

	return m.styles.Container.Render(title + "\n" + items + equipment + footer)
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
	`)

	if m.player != nil {
		content += fmt.Sprintf("\n\nExit codes earned: %d\n", m.player.ExitCodes)
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
	`)

	content += m.styles.Muted.Render("\n[Enter] Continue  [Q] Quit")

	return m.styles.Container.Render(content)
}
