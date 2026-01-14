// Package ui provides the Bubble Tea UI for /dev/dungeon.
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
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
	return nil
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
	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if m.currentView == ViewMainMenu {
			return m, tea.Quit
		}
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
		// TODO: Load saved game
		m.statusMsg = "No save file found"
		return m, nil
	case "Settings":
		// TODO: Settings screen
		return m, nil
	case "Quit":
		return m, tea.Quit
	}
	return m, nil
}

// startNewGame initializes a new game.
func (m *Model) startNewGame() {
	m.player = entity.NewPlayer(entity.PlayerClass(m.config.Game.StartingClass))
	m.currentView = ViewGame
	m.gameState = types.StateExploring
	m.statusMsg = "Welcome to /dev/dungeon. Navigate the filesystem. Survive."
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
	case "i":
		m.currentView = ViewInventory
	case "p", "esc":
		m.currentView = ViewPause
	}
	return m, nil
}

// movePlayer moves the player in a direction.
func (m *Model) movePlayer(dir types.Direction) {
	if m.player == nil {
		return
	}

	pos := m.player.Position()
	switch dir {
	case types.DirUp:
		pos.Y--
	case types.DirDown:
		pos.Y++
	case types.DirLeft:
		pos.X--
	case types.DirRight:
		pos.X++
	}

	// TODO: Collision detection with world
	m.player.SetPosition(pos)
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
		m.currentView = ViewMainMenu
		m.gameState = types.StateMainMenu
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
	// Stats panel
	stats := m.renderStats()

	// Map (placeholder for now)
	mapView := m.renderMap()

	// Log/status
	log := m.renderLog()

	// Layout: stats on right, map in center, log at bottom
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, mapView, "  ", stats)

	return m.styles.Container.Render(topRow + "\n" + log)
}

// renderStats renders the stats panel.
func (m *Model) renderStats() string {
	if m.player == nil {
		return ""
	}

	p := m.player
	content := fmt.Sprintf(
		"%s\n%s\n\n"+
			"RAM: %s/%d\n"+
			"CPU: %d\n"+
			"FD:  %s/%d\n"+
			"NICE: %d\n"+
			"UID: %d\n\n"+
			"Level: %d\n"+
			"XP: %d/%d\n"+
			"Floor: /home",
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

// renderMap renders the game map (placeholder).
func (m *Model) renderMap() string {
	// Placeholder map
	width := 40
	height := 15

	var mapStr string
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if m.player != nil && x == m.player.Position().X && y == m.player.Position().Y {
				mapStr += m.styles.Player.Render("@")
			} else if x == 0 || x == width-1 || y == 0 || y == height-1 {
				mapStr += m.styles.Wall.Render("#")
			} else {
				mapStr += m.styles.Floor.Render(".")
			}
		}
		mapStr += "\n"
	}

	return m.styles.MapBorder.Render(mapStr)
}

// renderLog renders the message log.
func (m *Model) renderLog() string {
	content := m.statusMsg
	if content == "" {
		content = "Ready."
	}

	footer := m.styles.Muted.Render("[WASD] Move  [I] Inventory  [P] Pause")

	return m.styles.LogPanel.Width(60).Render(content + "\n" + footer)
}

// viewCombat renders the combat view.
func (m *Model) viewCombat() string {
	// TODO: Full combat UI
	return m.styles.Container.Render("Combat View (TODO)")
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
