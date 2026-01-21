package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

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
	case ViewUnlockShop:
		return m.viewUnlockShop()
	case ViewLeaderboard:
		return m.viewLeaderboard()
	case ViewDailyLeaderboard:
		return m.viewDailyLeaderboard()
	case ViewConfirmDialog:
		return m.viewConfirmDialog()
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
	var title string
	if m.dailyRunMode {
		dateStr := time.Now().UTC().Format("January 2, 2006")
		title = m.styles.Title.Render(fmt.Sprintf(`
    ╔═══════════════════════════════════════════╗
    ║          DAILY RUN - %s          ║
    ║           SELECT YOUR PROCESS             ║
    ╚═══════════════════════════════════════════╝
	`, dateStr))
	} else {
		title = m.styles.Title.Render(`
    ╔═══════════════════════════════════════════╗
    ║           SELECT YOUR PROCESS             ║
    ╚═══════════════════════════════════════════╝
	`)
	}

	var menu string
	for i, class := range m.classOptions {
		cursor := "  "
		style := m.styles.MenuItem
		isUnlocked := m.isClassUnlocked(class)

		if i == m.classCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}

		classStr := string(class)
		if !isUnlocked {
			// Show locked class with price
			price := m.getClassUnlockPrice(class)
			classStr = fmt.Sprintf("%s [LOCKED - %d exit codes]", class, price)
			if i != m.classCursor {
				style = m.styles.Muted
			}
		}
		menu += style.Render(fmt.Sprintf("%s%s", cursor, classStr)) + "\n"
	}

	// Show description of selected class
	selectedClass := m.classOptions[m.classCursor]
	desc, stats := m.getClassDescription(selectedClass)
	details := "\n" + m.styles.Muted.Render("─── Class Info ───") + "\n"
	details += m.styles.Normal.Render(desc) + "\n\n"
	details += m.styles.Highlight.Render(stats) + "\n"

	if !m.isClassUnlocked(selectedClass) {
		details += "\n" + m.styles.Danger.Render("This class is locked. Visit the Unlocks shop to purchase it.") + "\n"
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Navigate  [Enter] Select  [Esc] Back")

	content := m.styles.Container.Render(title + "\n" + menu + details + footer)
	return m.centerContent(content)
}

// viewMainMenu renders the main menu.
func (m *Model) viewMainMenu() string {
	title := m.styles.Title.Render(`
    ╔═══════════════════════════════════════════╗
    ║         /dev/dungeon                      ║
    ║    Navigate the filesystem. Survive.      ║
    ╚═══════════════════════════════════════════╝
	`)

	// Show greeting for multiplayer users (centered to match title box width)
	greeting := ""
	if m.isMultiplayer && m.username != "" {
		greetingText := fmt.Sprintf("Welcome back, %s", m.username)
		// Center within ~47 char width to match title box
		greeting = lipgloss.NewStyle().Width(47).Align(lipgloss.Center).Foreground(m.styles.Highlight.GetForeground()).Render(greetingText) + "\n\n"
	}

	// Get daily seed for display
	dailySeed := getDailySeed()
	dailySeedStr := fmt.Sprintf("%d", dailySeed)
	if len(dailySeedStr) > 8 {
		dailySeedStr = dailySeedStr[:8] + "..."
	}

	var menu string
	for i, option := range m.menuOptions {
		cursor := "  "
		style := m.styles.MenuItem

		// Gray out options that aren't available
		isDisabled := false
		disabledReason := ""
		switch option {
		case "Continue":
			if !m.hasValidSave {
				isDisabled = true
				disabledReason = "no save"
			}
		case "Daily Run", "Daily Leaderboard":
			if !m.isMultiplayer {
				isDisabled = true
				disabledReason = "SSH only"
			}
		}

		if isDisabled {
			style = m.styles.Muted
		} else if i == m.menuCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}

		displayOption := option
		if isDisabled {
			displayOption = fmt.Sprintf("%s (%s)", option, disabledReason)
		} else if option == "Daily Run" {
			displayOption = fmt.Sprintf("Daily Run (%s)", dailySeedStr)
		}

		menu += style.Render(cursor+displayOption) + "\n"
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Navigate  [Enter] Select  [q] Quit")

	if m.statusMsg != "" {
		footer = m.styles.Danger.Render(m.statusMsg) + "\n" + footer
	}

	attribution := m.styles.Muted.Render("\n\nBuilt by @kwuchu  •  dev-dungeon.com")

	content := m.styles.Container.Render(title + "\n" + greeting + menu + "\n" + footer + attribution)
	return m.centerContent(content)
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
	var seed int64
	if m.engine != nil {
		floorName = m.engine.CurrentFloorType().FloorName()
		floorDepth = m.engine.CurrentDepth()
		seed = m.engine.MasterSeed()
	}

	// Show username if in multiplayer mode
	userLine := ""
	if m.isMultiplayer && m.username != "" {
		userLine = m.styles.Highlight.Render("@"+m.username) + "\n"
	}

	// Format seed display (truncate for readability)
	seedStr := fmt.Sprintf("%d", seed)
	if len(seedStr) > 10 {
		seedStr = seedStr[:10] + "..."
	}

	content := fmt.Sprintf(
		"%s%s\n%s\n\n"+
			"RAM: %s/%d\n"+
			"CPU: %d\n"+
			"FD:  %s/%d\n"+
			"NICE: %d\n"+
			"UID: %d\n\n"+
			"Level: %d\n"+
			"XP: %d/%d\n"+
			"Floor: %s\n"+
			"Depth: %d\n\n"+
			"Seed: %s",
		userLine,
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
		seedStr,
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

// colorizeEnemyRAM colors enemy RAM based on percentage.
func (m *Model) colorizeEnemyRAM(current, max int) string {
	pct := float64(current) / float64(max)
	str := fmt.Sprintf("%d", current)
	if pct > 0.6 {
		return str
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
	} else if pct > 0.25 {
		return m.styles.Muted.Render(str)
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

// centerContent centers content both horizontally and vertically in the terminal.
func (m *Model) centerContent(content string) string {
	w, h := m.width, m.height
	if w == 0 {
		w = 120
	}
	if h == 0 {
		h = 40
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// renderMap renders the game map from the engine with a viewport centered on the player.
func (m *Model) renderMap() string {
	if m.engine == nil {
		return m.styles.MapBorder.Render("No map loaded")
	}

	tiles := m.engine.GetVisibleTiles()
	if len(tiles) == 0 {
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

	footer := m.styles.Muted.Render("[WASD/hjkl] Move  [</>] Stairs  [I] Inv  [$] Shop  [M] Log  [P] Pause  [?] Help")

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

			// Colorize enemy RAM with accessibility indicators (like player RAM)
			enemyRAMStr := m.colorizeEnemyRAM(enemy.Stats.RAM, enemy.MaxStats.MaxRAM)

			enemyInfo += fmt.Sprintf("%s%s %-14s RAM: %s/%d  CPU: %d%s\n",
				targetIndicator,
				m.styles.Enemy.Render(string(enemy.Glyph())),
				enemy.Name(),
				enemyRAMStr,
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

	content := m.styles.Container.Render(title + playerInfo + enemyInfo + menu + log + footer)
	return m.centerContent(content)
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
		detailsPanel = "\n" + m.styles.Muted.Render("─── Details ───\n") + selectedDesc
	}

	footer := m.styles.Muted.Render("\n[↑/↓] Select  [Enter] Use/Equip  [U] Unequip  [D] Drop  [I/Esc] Close")

	content := m.styles.Container.Render(title + "\n" + items + equipment + detailsPanel + footer)
	return m.centerContent(content)
}

// getItemDetails returns formatted item details.
func (m *Model) getItemDetails(item *entity.Item) string {
	if item == nil {
		return ""
	}
	// Type and rarity
	typeRarity := m.styles.Muted.Render(fmt.Sprintf("[%s] ", item.ItemType))
	typeRarity += m.formatRarity(item.Rarity) + "\n"
	details := typeRarity
	details += m.styles.Normal.Render(item.Description) + "\n"
	statStr := m.formatStatBonus(item)
	if statStr != "" {
		details += m.styles.Highlight.Render(statStr) + "\n"
	}
	return details
}

// formatRarity returns a styled rarity string with visual prefix for accessibility.
func (m *Model) formatRarity(rarity entity.ItemRarity) string {
	name := rarity.String()
	switch rarity {
	case entity.RarityCommon:
		return m.styles.Muted.Render(name)
	case entity.RarityUncommon:
		return m.styles.Normal.Render("+ " + name)
	case entity.RarityRare:
		return m.styles.Highlight.Render("++ " + name)
	case entity.RarityEpic:
		return m.styles.Title.Render("+++ " + name)
	case entity.RarityLegendary:
		return m.styles.Danger.Render("*** " + name)
	default:
		return name
	}
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

// viewPause renders the pause menu.
func (m *Model) viewPause() string {
	c := m.styles.Title.Render("═══ PAUSED ═══\n\n")
	c += "[P/Esc] Resume\n"
	c += "[Q] Quit to Menu\n"

	content := m.styles.Container.Render(c)
	return m.centerContent(content)
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

	// Show exit codes earned
	content += "\n" + m.styles.Highlight.Render("─── Exit Codes Earned ───") + "\n"
	content += m.styles.Success.Render(fmt.Sprintf("  +%d exit codes", m.runExitCodesEarned)) + "\n"
	for _, breakdown := range m.runExitCodesBreakdown {
		content += m.styles.Muted.Render(fmt.Sprintf("    %s", breakdown)) + "\n"
	}
	if m.metaProgress != nil {
		content += m.styles.Normal.Render(fmt.Sprintf("\n  Total saved: %d", m.metaProgress.TotalExitCodes)) + "\n"
	}

	content += m.styles.Muted.Render("\n[Enter] Continue  [Q] Quit")

	result := m.styles.Container.Render(content)
	return m.centerContent(result)
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

	// Show exit codes earned
	content += "\n" + m.styles.Highlight.Render("─── Exit Codes Earned ───") + "\n"
	content += m.styles.Success.Render(fmt.Sprintf("  +%d exit codes", m.runExitCodesEarned)) + "\n"
	for _, breakdown := range m.runExitCodesBreakdown {
		content += m.styles.Muted.Render(fmt.Sprintf("    %s", breakdown)) + "\n"
	}
	if m.metaProgress != nil {
		content += m.styles.Normal.Render(fmt.Sprintf("\n  Total saved: %d", m.metaProgress.TotalExitCodes)) + "\n"
	}

	content += m.styles.Muted.Render("\n[Enter] Continue  [Q] Quit")

	result := m.styles.Container.Render(content)
	return m.centerContent(result)
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

	content := m.styles.Container.Render(title + menu + status + footer)
	return m.centerContent(content)
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
	actions += "  $                         - Open shop (ls -la)\n"
	actions += "  M                         - Message history\n"
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
	inventory += "  U                        - Unequip item\n"
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
	tips += "  - Find root_shard items to lower your UID (local play)\n"

	footer := m.styles.Muted.Render("\n[Esc/?/Enter] Close")

	content := m.styles.Container.Render(title + movement + actions + combat + inventory + stats + tips + footer)
	return m.centerContent(content)
}

// viewMessageHistory renders the scrollable message history.
func (m *Model) viewMessageHistory() string {
	title := m.styles.Title.Render("═══ MESSAGE LOG ═══") + "\n\n"

	if len(m.messageHistory) == 0 {
		c := m.styles.Muted.Render("No messages yet.")
		footer := m.styles.Muted.Render("\n[Esc/M/Enter] Close")
		content := m.styles.Container.Render(title + c + footer)
		return m.centerContent(content)
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

	result := m.styles.Container.Render(title + content + footer)
	return m.centerContent(result)
}

// viewIntro renders the animated intro sequence.
func (m *Model) viewIntro() string {
	if m.introFrame < 0 || m.introFrame >= len(introFrames) {
		return ""
	}

	frame := introFrames[m.introFrame]

	// Progress indicator
	progress := fmt.Sprintf("[%d/%d]", m.introFrame+1, len(introFrames))

	footer := m.styles.Muted.Render("\n\n" + progress + "   [Space/→] Next   [←] Back   [Esc] Skip")

	content := m.styles.Container.Render(m.styles.Title.Render(frame) + footer)
	return m.centerContent(content)
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

		// Price with accessibility indicator (not just color)
		priceStr := fmt.Sprintf("%4d", item.Price)
		if !item.InStock {
			priceStr = m.styles.Muted.Render(priceStr + "  ")
		} else if m.player.ExitCodes < item.Price {
			// Can't afford - red with X indicator
			priceStr = m.styles.Danger.Render(priceStr + " x")
		} else {
			// Can afford - green with checkmark
			priceStr = m.styles.Success.Render(priceStr + " +")
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

	content := m.styles.Container.Render(title + balance + items + details + footer)
	return m.centerContent(content)
}

// viewUnlockShop renders the unlock shop interface.
func (m *Model) viewUnlockShop() string {
	title := m.styles.Title.Render(`
    ╔═══════════════════════════════════════════╗
    ║             PERMANENT UNLOCKS             ║
    ╚═══════════════════════════════════════════╝
	`) + "\n"

	// Show total exit codes
	balance := m.styles.Muted.Render("Exit Codes: ") + m.styles.Highlight.Render(fmt.Sprintf("%d", m.metaProgress.TotalExitCodes)) + "\n\n"

	// Category tabs
	categories := []string{"Classes", "Bonuses", "Items"}
	var tabs string
	for i, cat := range categories {
		if i == m.unlockCategory {
			tabs += m.styles.MenuSelected.Render(fmt.Sprintf(" [%s] ", cat))
		} else {
			tabs += m.styles.Muted.Render(fmt.Sprintf("  %s  ", cat))
		}
	}
	tabs += "\n" + m.styles.Muted.Render("─────────────────────────────────────────────") + "\n\n"

	// Category content
	var content string
	switch m.unlockCategory {
	case 0:
		content = m.renderUnlockClasses()
	case 1:
		content = m.renderUnlockBonuses()
	case 2:
		content = m.renderUnlockItems()
	}

	// Run statistics if available
	stats := ""
	if m.metaProgress != nil {
		stats = "\n" + m.styles.Muted.Render("─── Statistics ───") + "\n"
		stats += fmt.Sprintf("  Runs Completed: %d\n", m.metaProgress.RunsCompleted)
		stats += fmt.Sprintf("  Deepest Floor:  %d\n", m.metaProgress.DeepestFloor)
		stats += fmt.Sprintf("  Total Deaths:   %d\n", m.metaProgress.TotalDeaths)
	}

	// Always reserve space for status message to prevent layout shift
	statusLine := "\n"
	if m.statusMsg != "" {
		statusLine = "\n" + m.styles.Highlight.Render(m.statusMsg)
	}

	footer := statusLine + m.styles.Muted.Render("\n[←/→] Category  [↑/↓] Navigate  [Enter] Purchase  [Esc] Back")

	result := m.styles.Container.Render(title + balance + tabs + content + stats + footer)
	return m.centerContent(result)
}

// renderUnlockClasses renders the classes category in the unlock shop.
func (m *Model) renderUnlockClasses() string {
	classes := []struct {
		class entity.PlayerClass
		desc  string
	}{
		{entity.ClassCron, "Scheduler daemon. Fast and precise."},
		{entity.ClassBash, "Powerful shell. High attack output."},
		{entity.ClassVim, "Complex editor. Many abilities."},
		{entity.ClassSudo, "Root access. High risk, high power."},
	}

	var content string
	for i, c := range classes {
		cursor := "  "
		style := m.styles.MenuItem
		if i == m.unlockCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}

		unlocked := m.isClassUnlocked(c.class)
		price := m.getClassUnlockPrice(c.class)

		// Status with accessibility indicators (not just color)
		var statusStr string
		if unlocked {
			statusStr = m.styles.Success.Render(" [UNLOCKED]")
		} else if m.metaProgress.TotalExitCodes >= price {
			statusStr = m.styles.Highlight.Render(fmt.Sprintf(" + %d exit codes", price))
		} else {
			statusStr = m.styles.Danger.Render(fmt.Sprintf(" x %d exit codes", price))
		}

		content += style.Render(fmt.Sprintf("%s%s%s", cursor, c.class, statusStr)) + "\n"
		if i == m.unlockCursor {
			content += m.styles.Muted.Render(fmt.Sprintf("     %s", c.desc)) + "\n"
		}
	}

	return content
}

// renderUnlockBonuses renders the bonuses category in the unlock shop.
func (m *Model) renderUnlockBonuses() string {
	bonuses := m.getUnlockableBonuses()

	var content string
	for i, bonus := range bonuses {
		cursor := "  "
		style := m.styles.MenuItem
		if i == m.unlockCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}

		price := m.getBonusPrice(bonus)
		levelStr := fmt.Sprintf("[%d/%d]", bonus.CurrentLevel, bonus.MaxLevel)

		// Status with accessibility indicators (not just color)
		var statusStr string
		if bonus.CurrentLevel >= bonus.MaxLevel {
			statusStr = m.styles.Success.Render(" [MAX]")
		} else if m.metaProgress.TotalExitCodes >= price {
			statusStr = m.styles.Highlight.Render(fmt.Sprintf(" + %d exit codes", price))
		} else {
			statusStr = m.styles.Danger.Render(fmt.Sprintf(" x %d exit codes", price))
		}

		content += style.Render(fmt.Sprintf("%s%s %s%s", cursor, bonus.Name, levelStr, statusStr)) + "\n"
		if i == m.unlockCursor {
			content += m.styles.Muted.Render(fmt.Sprintf("     %s", bonus.Description)) + "\n"
		}
	}

	return content
}

// renderUnlockItems renders the items category in the unlock shop.
func (m *Model) renderUnlockItems() string {
	items := m.getUnlockableItems()

	var content string
	for i, item := range items {
		cursor := "  "
		style := m.styles.MenuItem
		if i == m.unlockCursor {
			cursor = "> "
			style = m.styles.MenuSelected
		}

		// Status with accessibility indicators (not just color)
		var statusStr string
		if item.Unlocked {
			statusStr = m.styles.Success.Render(" [UNLOCKED]")
		} else if m.metaProgress.TotalExitCodes >= item.Price {
			statusStr = m.styles.Highlight.Render(fmt.Sprintf(" + %d exit codes", item.Price))
		} else {
			statusStr = m.styles.Danger.Render(fmt.Sprintf(" x %d exit codes", item.Price))
		}

		content += style.Render(fmt.Sprintf("%s%s%s", cursor, item.Name, statusStr)) + "\n"
		if i == m.unlockCursor {
			content += m.styles.Muted.Render(fmt.Sprintf("     %s", item.Description)) + "\n"
		}
	}

	return content
}

// viewLeaderboard renders the leaderboard interface.
func (m *Model) viewLeaderboard() string {
	title := m.styles.Title.Render(`
    ╔═══════════════════════════════════════════╗
    ║              LEADERBOARD                  ║
    ╚═══════════════════════════════════════════╝
	`) + "\n"

	// Run type tabs
	runTypes := []struct {
		id    string
		label string
	}{
		{"all", "All Runs"},
		{"standard", "Standard"},
		{"daily", "Daily"},
	}
	var tabs string
	for _, rt := range runTypes {
		if rt.id == m.leaderboardRunType {
			tabs += m.styles.MenuSelected.Render(fmt.Sprintf(" [%s] ", rt.label))
		} else {
			tabs += m.styles.Muted.Render(fmt.Sprintf("  %s  ", rt.label))
		}
	}
	tabs += "\n" + m.styles.Muted.Render("─────────────────────────────────────────────") + "\n\n"

	// Content
	var content string
	if !m.isMultiplayer {
		// Local mode - no leaderboard access
		content = m.styles.Muted.Render("  Leaderboards are only available via SSH.") + "\n\n"
		content += "  " + m.styles.Muted.Render("Connect with:") + "\n"
		content += "  " + m.styles.Highlight.Render("ssh player@dev-dungeon.com") + "\n\n"
	} else if m.leaderboardError != "" {
		content = m.styles.Danger.Render("  " + m.leaderboardError + "\n\n")
		content += m.styles.Muted.Render("  Press [R] to retry\n")
	} else if len(m.leaderboardEntries) == 0 {
		content = m.styles.Muted.Render("  No entries yet. Be the first to make the board!\n")
	} else {
		// Header - format columns consistently
		header := fmt.Sprintf("   %-4s %-12s %8s %7s %-8s", "RANK", "PLAYER", "SCORE", "FLOORS", "CLASS")
		content = m.styles.Muted.Render(header) + "\n"
		content += m.styles.Muted.Render("   ─────────────────────────────────────────") + "\n"

		// Entries
		for i, entry := range m.leaderboardEntries {
			cursor := "   "
			if i == m.leaderboardCursor {
				cursor = " > "
			}

			// Truncate username if too long
			username := entry.Username
			if len(username) > 10 {
				username = username[:9] + "…"
			}

			// Build the row with consistent formatting
			line := fmt.Sprintf("%-4d %-12s %8d %7d %-8s",
				entry.Rank,
				username,
				entry.Score,
				entry.FloorsCleared,
				entry.Class,
			)

			// Apply style based on selection/highlight
			var styledLine string
			if m.username != "" && entry.Username == m.username {
				styledLine = m.styles.Highlight.Render(cursor + line)
			} else if i == m.leaderboardCursor {
				styledLine = m.styles.MenuSelected.Render(cursor + line)
			} else {
				styledLine = cursor + line
			}
			content += styledLine + "\n"
		}
	}

	footer := "\n" + m.styles.Muted.Render("[←/→] Filter  [↑/↓] Navigate  [R] Refresh  [Esc] Back")

	result := m.styles.Container.Render(title + tabs + content + footer)
	return m.centerContent(result)
}

// viewDailyLeaderboard renders the date-navigable daily leaderboard.
func (m *Model) viewDailyLeaderboard() string {
	// Format the date for display
	dateStr := m.dailyLeaderboardDate.Format("Jan 2, 2006")
	today := time.Now().UTC().Truncate(24 * time.Hour)
	minDate := today.AddDate(0, 0, -6)

	// Navigation arrows - check if navigation is allowed
	canGoBack := !m.dailyLeaderboardDate.Equal(minDate) && !m.dailyLeaderboardDate.Before(minDate)
	canGoForward := !m.dailyLeaderboardDate.Equal(today) && !m.dailyLeaderboardDate.After(today)

	// Build navigation arrows - show muted when disabled for accessibility
	leftArrow := "←"
	rightArrow := "→"
	if !canGoBack {
		leftArrow = m.styles.Muted.Render("←")
	}
	if !canGoForward {
		rightArrow = m.styles.Muted.Render("→")
	}

	// Build date navigation line (43 chars inner width to match box)
	// Layout: 11 spaces + arrow + 3 spaces + date(12) + 3 spaces + arrow + 12 spaces = 43
	dateLine := fmt.Sprintf("           %s   %-12s   %s            ", leftArrow, dateStr, rightArrow)

	title := m.styles.Title.Render(
		"╔═══════════════════════════════════════════╗\n"+
			"║            DAILY LEADERBOARD              ║\n"+
			"║"+dateLine+"║\n"+
			"╚═══════════════════════════════════════════╝") + "\n"

	// Content
	var content string
	if !m.isMultiplayer {
		// Local mode - no leaderboard access
		content = m.styles.Muted.Render("  Daily leaderboards are only available via SSH.\n\n")
		content += m.styles.Normal.Render("  Connect with:\n")
		content += m.styles.Highlight.Render("  ssh player@dev-dungeon.com\n\n")
	} else if m.dailyLeaderboardError != "" {
		content = m.styles.Danger.Render("  " + m.dailyLeaderboardError + "\n\n")
		content += m.styles.Muted.Render("  Press [R] to retry\n")
	} else if len(m.dailyLeaderboardEntries) == 0 {
		content = m.styles.Muted.Render("  No entries for this day.\n")
		if m.dailyLeaderboardDate.Equal(today) {
			content += m.styles.Normal.Render("\n  Be the first to complete today's daily run!\n")
		}
	} else {
		// Header - format columns consistently
		header := fmt.Sprintf("   %-4s %-12s %8s %7s %-8s", "RANK", "PLAYER", "SCORE", "FLOORS", "CLASS")
		content = m.styles.Muted.Render(header) + "\n"
		content += m.styles.Muted.Render("   ─────────────────────────────────────────") + "\n"

		// Top entries
		for _, entry := range m.dailyLeaderboardEntries {
			// Truncate username if too long
			username := entry.Username
			if len(username) > 10 {
				username = username[:9] + "…"
			}

			line := fmt.Sprintf("   %-4d %-12s %8d %7d %-8s",
				entry.Rank,
				username,
				entry.Score,
				entry.FloorsCleared,
				entry.Class,
			)

			// Highlight current user
			if m.username != "" && entry.Username == m.username {
				content += m.styles.Highlight.Render(line) + "\n"
			} else {
				content += line + "\n"
			}
		}

		// Show player's position if not in top N
		if m.dailyPlayerEntry != nil && m.dailyPlayerRank > len(m.dailyLeaderboardEntries) {
			content += m.styles.Muted.Render("   ···") + "\n"

			username := m.dailyPlayerEntry.Username
			if len(username) > 10 {
				username = username[:9] + "…"
			}

			line := fmt.Sprintf("   %-4d %-12s %8d %7d %-8s  ← You",
				m.dailyPlayerRank,
				username,
				m.dailyPlayerEntry.Score,
				m.dailyPlayerEntry.FloorsCleared,
				m.dailyPlayerEntry.Class,
			)
			content += m.styles.Highlight.Render(line) + "\n"
		} else if m.dailyPlayerRank == 0 && m.dailyLeaderboardDate.Equal(today) {
			// Player hasn't done today's daily yet
			content += "\n" + m.styles.Muted.Render("   You haven't completed today's daily run yet.") + "\n"
		}
	}

	footer := "\n" + m.styles.Muted.Render("[←/→] Change Date  [R] Refresh  [Esc] Back")

	result := m.styles.Container.Render(title + content + footer)
	return m.centerContent(result)
}

// viewConfirmDialog renders a yes/no confirmation dialog.
func (m *Model) viewConfirmDialog() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("Confirm"))
	b.WriteString("\n\n")
	b.WriteString(m.confirmMessage)
	b.WriteString("\n\n")
	b.WriteString(m.styles.Muted.Render("[Y] Yes  [N] No"))
	return m.centerContent(m.styles.Container.Render(b.String()))
}
