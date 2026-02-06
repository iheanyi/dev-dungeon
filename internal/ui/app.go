// Package ui provides the Bubble Tea UI for /dev/dungeon.
package ui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/content"
	"github.com/iheanyi/devdungeon/internal/dungeon"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/save"
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
	ViewUnlockShop
	ViewLeaderboard
	ViewDailyLeaderboard
	ViewConfirmDialog
)

// introTickMsg is sent to advance intro animation.
type introTickMsg struct{}

// introFrames contains the intro sequence frames (52 chars wide for compatibility).
var introFrames = []string{
	`
  ╔════════════════════════════════════════════════╗
  ║     /dev/dungeon                               ║
  ║                                                ║
  ║     ██████╗ ███████╗██╗   ██╗                  ║
  ║     ██╔══██╗██╔════╝██║   ██║                  ║
  ║     ██║  ██║█████╗  ██║   ██║                  ║
  ║     ██║  ██║██╔══╝  ╚██╗ ██╔╝                  ║
  ║     ██████╔╝███████╗ ╚████╔╝                   ║
  ║     ╚═════╝ ╚══════╝  ╚═══╝                    ║
  ║                                                ║
  ║     A Unix-themed terminal roguelike           ║
  ║                                                ║
  ╚════════════════════════════════════════════════╝
`,
	`
  ╔════════════════════════════════════════════════╗
  ║                                                ║
  ║          >>> SYSTEM ALERT <<<                  ║
  ║                                                ║
  ║          ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                    ║
  ║          ▓  KERNEL PANIC  ▓                    ║
  ║          ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                    ║
  ║                                                ║
  ║     The system has been compromised.           ║
  ║                                                ║
  ╚════════════════════════════════════════════════╝
`,
	`
  ╔════════════════════════════════════════════════╗
  ║                                                ║
  ║  Something went wrong deep in /dev/null.       ║
  ║                                                ║
  ║  Rogue processes have escaped.                 ║
  ║  Zombie processes walk the directories.        ║
  ║  Daemons have turned hostile.                  ║
  ║  Fork bombs multiply unchecked.                ║
  ║                                                ║
  ║  And at the heart of it all...                 ║
  ║  The KERNEL PANIC awaits.                      ║
  ║                                                ║
  ╚════════════════════════════════════════════════╝
`,
	`
  ╔════════════════════════════════════════════════╗
  ║                                                ║
  ║  You are a newly spawned process.              ║
  ║                                                ║
  ║  Navigate from /home through the filesystem,   ║
  ║  descending ever deeper into the system.       ║
  ║                                                ║
  ║  /home → /tmp → /var → /etc →                  ║
  ║  /usr → /bin → /sys → /dev/null                ║
  ║                                                ║
  ║  Fight. Survive. Find the KERNEL PANIC.        ║
  ║  And end this madness.                         ║
  ║                                                ║
  ╚════════════════════════════════════════════════╝
`,
	`
  ╔════════════════════════════════════════════════╗
  ║                                                ║
  ║         YOUR STATS EXPLAINED:                  ║
  ║                                                ║
  ║  RAM  - Health. 0 = OOM killed.                ║
  ║  CPU  - Attack power. Kill faster.             ║
  ║  FD   - File descriptors. Fuel abilities.      ║
  ║  NICE - Priority. Lower = faster + crits.      ║
  ║  UID  - User ID. 0 = root = god mode.          ║
  ║                                                ║
  ║         Good luck, process.                    ║
  ║         The system depends on you.             ║
  ║                                                ║
  ╚════════════════════════════════════════════════╝
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
	currentView ViewType
	width       int
	height      int

	// Menu state
	menuCursor  int
	menuOptions []string

	// Combat state
	combatCursor   int
	combatLog      []string
	enemies        []*entity.Enemy
	combat         *game.CombatState
	selectingSkill bool
	skillCursor    int
	targetCursor   int // Which enemy is targeted

	// Inventory state
	invCursor int

	// Class selection state
	classCursor  int
	classOptions []entity.PlayerClass

	// Admin console state
	adminCursor  int
	adminOptions []string
	godMode      bool
	prevView     ViewType // View to return to after admin

	// Confirmation dialog state
	confirmMessage    string   // Message to display
	confirmAction     func()   // Action to execute on "Yes"
	confirmReturnView ViewType // View to return to on "No"

	// Messages
	statusMsg        string
	messageHistory   []string // Full message history
	messageScrollIdx int      // Current scroll position (0 = most recent)

	// Intro animation state
	introFrame    int                // Current frame of intro
	introSkipped  bool               // User skipped intro
	introFromMenu bool               // Viewing intro from menu (returns to menu, not game)
	pendingClass  entity.PlayerClass // Class to use after intro

	// Shop state
	shopCursor int
	shopItems  []ShopItem

	// Meta-progression state
	metaProgress *save.MetaProgress
	saveManager  *save.Manager

	// Unlock shop state
	unlockCursor   int
	unlockCategory int // 0=classes, 1=bonuses, 2=items

	// Exit codes earned this run (calculated on death/victory)
	runExitCodesEarned    int
	runExitCodesBreakdown []string

	// Multiplayer session info
	isMultiplayer  bool   // True when connected via SSH (disables cheats)
	username       string // Authenticated username for display
	dailyRunMode   bool   // True when starting a daily run (uses fixed seed)
	currentRunType string // "standard", "daily", or "seeded" for current game

	// Save state
	hasValidSave        bool                // True if a valid save file exists
	pendingSave         *save.SaveData      // Pending save data from DB (multiplayer) to load on Continue
	quitRequested       bool                // True when user wants to quit (for sessionWrapper to intercept)
	saveCallback        SaveCallback        // Callback to save game (set by server for multiplayer)
	clearSaveCallback   ClearSaveCallback   // Callback to delete save (set by server for multiplayer)
	metaProgressUpdater MetaProgressUpdater // Callback to update meta progress (set by server for multiplayer)

	// Daily run status (for preventing duplicate daily runs)
	dailyRunCompleted   bool                // True if player has leaderboard entry for today's daily
	dailyRunInProgress  bool                // True if pendingSave.MasterSeed matches today's daily seed
	dailySeed           int64               // Today's daily seed (set by session)
	submitDailyCallback SubmitDailyCallback // Callback to submit abandoned daily run

	// Run time tracking
	runStartTime     time.Time // When this run first started (persisted across saves)
	sessionStartTime time.Time // When this session started (for calculating current session duration)
	elapsedSeconds   int       // Accumulated play time from previous sessions

	// Leaderboard state
	leaderboardEntries   []LeaderboardEntry
	leaderboardCursor    int
	leaderboardRunType   string               // "all", "standard", "daily"
	leaderboardFetcher   LeaderboardFetcher   // Callback to fetch data (set by server)
	leaderboardSubmitter LeaderboardSubmitter // Callback to submit score (set by server)
	leaderboardError     string               // Error message if fetch failed

	// Daily leaderboard state (date-navigable)
	dailyLeaderboardDate    time.Time               // Currently selected date
	dailyLeaderboardEntries []LeaderboardEntry      // Top N entries for selected date
	dailyPlayerRank         int                     // Player's rank (0 if not on board)
	dailyPlayerEntry        *LeaderboardEntry       // Player's entry (nil if not on board)
	dailyLeaderboardFetcher DailyLeaderboardFetcher // Callback to fetch daily data
	dailyLeaderboardError   string                  // Error message if fetch failed

	// Styles
	styles *Styles
}

// ShopItem represents an item for sale.
type ShopItem struct {
	TemplateID string
	Name       string
	Price      int
	InStock    bool
}

// UnlockableClass represents a class available for unlock.
type UnlockableClass struct {
	Class       entity.PlayerClass
	Name        string
	Description string
	Price       int
	Unlocked    bool
}

// UnlockableBonus represents a permanent stat bonus for purchase.
type UnlockableBonus struct {
	ID           string
	Name         string
	Description  string
	StatType     string // "RAM", "CPU", "MEM", "NICE"
	BonusAmount  int
	CurrentLevel int
	MaxLevel     int
	BasePrice    int
}

// UnlockableItem represents a loot pool item unlock.
type UnlockableItem struct {
	TemplateID  string
	Name        string
	Description string
	Price       int
	Unlocked    bool
}

// LeaderboardEntry represents a score entry for display in the leaderboard.
type LeaderboardEntry struct {
	Rank          int
	Username      string
	Score         int
	FloorsCleared int
	Class         string
	Seed          int64
	RunType       string // "standard", "daily", "seeded"
}

// LeaderboardFetcher is a callback function to fetch leaderboard data.
// Returns entries and any error. Set by server for multiplayer mode.
type LeaderboardFetcher func(runType string, limit int) ([]LeaderboardEntry, error)

// DailyLeaderboardFetcher is a callback to fetch daily leaderboard for a specific date.
// Returns top entries, player's rank, player's entry (if any), and any error.
type DailyLeaderboardFetcher func(date time.Time, limit int, userID int) ([]LeaderboardEntry, int, *LeaderboardEntry, error)

// SaveCallback is a callback function to save game state.
// Called before returning to main menu in multiplayer mode.
// Returns the save data that was persisted, for updating pendingSave.
type SaveCallback func() (*save.SaveData, error)

// LeaderboardSubmitter is a callback to submit a score to the leaderboard.
// Called on death or victory. Parameters: score, floorsCleared, timeSeconds, class, seed, runType, victory.
type LeaderboardSubmitter func(score, floorsCleared, timeSeconds int, class string, seed int64, runType string, victory bool) error

// SubmitDailyCallback submits an abandoned daily run to leaderboard.
// Called with the save data when player starts a new game while having an in-progress daily.
type SubmitDailyCallback func(saveData *save.SaveData) error

// ClearSaveCallback is a callback to delete the game save.
// Called on death or victory when the run ends and save should be deleted.
type ClearSaveCallback func() error

// MetaProgressUpdater is a callback to update meta progress in the database.
// Called on death or victory to persist exit codes earned.
// Parameters: exitCodesEarned, victory, maxDepthReached
type MetaProgressUpdater func(exitCodesEarned int, victory bool, maxDepthReached int) error

// Styles holds all UI styles.
type Styles struct {
	// Layout
	Container lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style

	// Game view
	MapBorder lipgloss.Style
	StatPanel lipgloss.Style
	LogPanel  lipgloss.Style

	// Text
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Normal    lipgloss.Style
	Highlight lipgloss.Style
	Danger    lipgloss.Style
	Success   lipgloss.Style
	Muted     lipgloss.Style

	// Menu
	MenuItem     lipgloss.Style
	MenuSelected lipgloss.Style

	// Tiles
	Wall   lipgloss.Style
	Floor  lipgloss.Style
	Player lipgloss.Style
	Enemy  lipgloss.Style
	Item   lipgloss.Style
	Stairs lipgloss.Style
}

// NewStyles creates the default styles.
func NewStyles() *Styles {
	return newStyles(nil)
}

// NewStylesWithRenderer creates styles using a custom renderer (for SSH sessions).
func NewStylesWithRenderer(renderer *lipgloss.Renderer) *Styles {
	return newStyles(renderer)
}

// newStyles is the internal constructor.
func newStyles(renderer *lipgloss.Renderer) *Styles {
	// Helper to create styles with optional renderer
	newStyle := func() lipgloss.Style {
		if renderer != nil {
			return renderer.NewStyle()
		}
		return lipgloss.NewStyle()
	}

	return &Styles{
		Container: newStyle().
			Padding(1, 2),
		Header: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("240")),
		Footer: newStyle().
			Foreground(lipgloss.Color("240")),

		MapBorder: newStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")),
		StatPanel: newStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1),
		LogPanel: newStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),

		Title: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		Subtitle: newStyle().
			Foreground(lipgloss.Color("243")),
		Normal: newStyle().
			Foreground(lipgloss.Color("252")),
		Highlight: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		Danger: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),
		Success: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("46")),
		Muted: newStyle().
			Foreground(lipgloss.Color("240")),

		MenuItem: newStyle().
			Foreground(lipgloss.Color("252")),
		MenuSelected: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236")),

		Wall: newStyle().
			Foreground(lipgloss.Color("240")),
		Floor: newStyle().
			Foreground(lipgloss.Color("238")),
		Player: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		Enemy: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),
		Item: newStyle().
			Foreground(lipgloss.Color("226")),
		Stairs: newStyle().
			Bold(true).
			Foreground(lipgloss.Color("45")),
	}
}

// New creates a new UI model.
func New(cfg *config.Config) *Model {
	return newModel(cfg, nil)
}

// NewWithRenderer creates a new UI model with a custom renderer (for SSH sessions).
func NewWithRenderer(cfg *config.Config, renderer *lipgloss.Renderer) *Model {
	return newModel(cfg, renderer)
}

// newModel is the internal constructor.
func newModel(cfg *config.Config, renderer *lipgloss.Renderer) *Model {
	// Initialize save manager for meta-progress
	saveMgr, _ := save.NewManager(save.DefaultConfig())

	// Load meta-progress
	var metaProg *save.MetaProgress
	if saveMgr != nil {
		metaProg, _ = saveMgr.LoadMetaProgress()
	}
	if metaProg == nil {
		defaultMeta := save.NewMetaProgress()
		metaProg = &defaultMeta
	}

	m := &Model{
		config:      cfg,
		currentView: ViewMainMenu,
		gameState:   types.StateMainMenu,
		menuOptions: []string{
			"New Game",
			"Daily Run",
			"Continue",
			"Leaderboard",
			"Daily Leaderboard",
			"Unlocks",
			"How to Play",
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
		adminCursor:    0,
		combatLog:      make([]string, 0),
		messageHistory: make([]string, 0),
		metaProgress:   metaProg,
		saveManager:    saveMgr,
	}

	// Use provided renderer or default styles
	if renderer != nil {
		m.styles = NewStylesWithRenderer(renderer)
	} else {
		m.styles = NewStyles()
	}

	// Check if a valid save exists
	m.hasValidSave = m.checkValidSave()

	return m
}

// checkValidSave checks if a valid save file exists.
func (m *Model) checkValidSave() bool {
	if m.saveManager == nil {
		return false
	}

	data, err := m.saveManager.LoadLatest()
	if err != nil || data == nil {
		return false
	}

	// Validate save data has valid player stats (same validation as engine.LoadGame)
	if data.Player.MaxStats.MaxRAM <= 0 || data.Player.Level <= 0 {
		return false
	}

	return true
}

// GetEngine returns the game engine (for external save/load).
func (m *Model) GetEngine() *game.Engine {
	return m.engine
}

// GetSaveData returns the current save data with run type included.
// This wraps the engine's GetSaveData and adds UI-level state like RunType.
func (m *Model) GetSaveData() *save.SaveData {
	if m.engine == nil {
		return nil
	}
	saveData := m.engine.GetSaveData()
	if saveData == nil {
		return nil
	}
	// Include run type in save data
	runType := m.currentRunType
	if runType == "" {
		runType = "standard"
	}
	saveData.RunType = runType

	// Include time tracking in save data
	saveData.RunStartTime = m.runStartTime
	// Calculate total elapsed: previous sessions + current session duration
	currentSessionDuration := int(time.Since(m.sessionStartTime).Seconds())
	saveData.ElapsedSeconds = m.elapsedSeconds + currentSessionDuration

	return saveData
}

// GetTotalPlayTime returns the total play time for the current run in seconds.
// This includes all previous sessions plus the current session.
func (m *Model) GetTotalPlayTime() int {
	if m.sessionStartTime.IsZero() {
		return m.elapsedSeconds
	}
	currentSessionDuration := int(time.Since(m.sessionStartTime).Seconds())
	return m.elapsedSeconds + currentSessionDuration
}

// SetMultiplayerMode configures the model for multiplayer (SSH) sessions.
// This disables admin console and godmode cheats.
func (m *Model) SetMultiplayerMode(username string) {
	m.isMultiplayer = true
	m.username = username
}

// SetLeaderboardFetcher sets the callback function for fetching leaderboard data.
// This is typically set by the server for multiplayer sessions.
func (m *Model) SetLeaderboardFetcher(fetcher LeaderboardFetcher) {
	m.leaderboardFetcher = fetcher
}

// SetDailyLeaderboardFetcher sets the callback function for fetching daily leaderboard data.
// This is typically set by the server for multiplayer sessions.
func (m *Model) SetDailyLeaderboardFetcher(fetcher DailyLeaderboardFetcher) {
	m.dailyLeaderboardFetcher = fetcher
}

// SetLeaderboardSubmitter sets the callback for submitting scores to the leaderboard.
// This is called on player death or victory.
func (m *Model) SetLeaderboardSubmitter(submitter LeaderboardSubmitter) {
	m.leaderboardSubmitter = submitter
}

// SetSaveCallback sets the callback function for saving game state.
// This is typically set by the server for multiplayer sessions.
func (m *Model) SetSaveCallback(callback SaveCallback) {
	m.saveCallback = callback
}

// SetClearSaveCallback sets the callback function for deleting the game save.
// This is called on death or victory when the run ends.
func (m *Model) SetClearSaveCallback(callback ClearSaveCallback) {
	m.clearSaveCallback = callback
}

// SetMetaProgressUpdater sets the callback function for updating meta progress.
// This is used by the server for multiplayer sessions to persist exit codes.
func (m *Model) SetMetaProgressUpdater(updater MetaProgressUpdater) {
	m.metaProgressUpdater = updater
}

// SetMetaProgress replaces the model's meta-progress state.
// This is used by the server to inject DB-loaded meta-progress in multiplayer,
// overriding the default (empty) meta-progress from the constructor.
func (m *Model) SetMetaProgress(mp *save.MetaProgress) {
	if mp != nil {
		m.metaProgress = mp
	}
}

// SetHasValidSave sets whether a valid save exists and optionally stores the save data.
// This is used by the server for multiplayer sessions where saves are stored in the database.
// The save data will be loaded when the user selects Continue.
func (m *Model) SetHasValidSave(hasValidSave bool, saveData ...*save.SaveData) {
	m.hasValidSave = hasValidSave
	if len(saveData) > 0 {
		m.pendingSave = saveData[0]
	}
}

// SetDailyRunStatus sets the daily run state for menu gating.
func (m *Model) SetDailyRunStatus(completed, inProgress bool, dailySeed int64) {
	m.dailyRunCompleted = completed
	m.dailyRunInProgress = inProgress
	m.dailySeed = dailySeed
}

// SetSubmitDailyCallback sets the callback for submitting abandoned daily runs.
func (m *Model) SetSubmitDailyCallback(cb SubmitDailyCallback) {
	m.submitDailyCallback = cb
}

// handleMenuSelection is a test helper that simulates selecting a menu option.
// It finds the option in menuOptions and triggers the selection logic.
func (m *Model) handleMenuSelection(option string) {
	// Find the option in menuOptions
	for i, opt := range m.menuOptions {
		if opt == option {
			m.menuCursor = i
			m.selectMenuItem()
			return
		}
	}
	// If option not found, just set up the menu and call selectMenuItem
	// This handles cases where menuOptions isn't fully set up
	m.menuOptions = []string{option}
	m.menuCursor = 0
	m.selectMenuItem()
}

// handleConfirmDialogKey is a test helper that simulates a key press in the confirm dialog.
func (m *Model) handleConfirmDialogKey(key string) {
	switch key {
	case "y", "Y", "enter":
		if m.confirmAction != nil {
			m.confirmAction()
		}
	case "n", "N", "esc", "q":
		m.currentView = m.confirmReturnView
	}
}

// showConfirmDialog displays a yes/no confirmation dialog.
func (m *Model) showConfirmDialog(message string, onConfirm func(), returnView ViewType) {
	m.confirmMessage = message
	m.confirmAction = onConfirm
	m.confirmReturnView = returnView
	m.currentView = ViewConfirmDialog
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
		// No longer used - intro is now user-paced
		return m, nil
	}
	return m, nil
}

// handleKeyPress processes keyboard input.
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C ALWAYS quits - this is the universal escape hatch
	if msg.String() == "ctrl+c" {
		return m, m.requestQuit()
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
	case ViewUnlockShop:
		return m.updateUnlockShop(msg)
	case ViewLeaderboard:
		return m.updateLeaderboard(msg)
	case ViewDailyLeaderboard:
		return m.updateDailyLeaderboard(msg)
	case ViewConfirmDialog:
		return m.updateConfirmDialog(msg)
	}

	return m, nil
}

// getDailySeed returns today's seed for daily runs.
// Uses UTC date to ensure the same seed worldwide for leaderboard fairness.
func getDailySeed() int64 {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	return today.UnixNano()
}

// shutdown gracefully shuts down the game.
func (m *Model) shutdown() {
	if m.engine != nil {
		m.engine.Shutdown()
	}
}

// WantsToQuit returns true if the user has requested to quit.
// This is checked by the session wrapper to save before quitting.
func (m *Model) WantsToQuit() bool {
	return m.quitRequested
}

// ClearQuitRequest clears the quit request flag.
func (m *Model) ClearQuitRequest() {
	m.quitRequested = false
}

// requestQuit sets the quit flag for the session wrapper to handle.
func (m *Model) requestQuit() tea.Cmd {
	m.quitRequested = true
	m.shutdown()
	return tea.Quit
}

// updateMainMenu handles main menu input.
func (m *Model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "w":
		m.menuCursor--
		if m.menuCursor < 0 {
			m.menuCursor = len(m.menuOptions) - 1
		}
		// Skip disabled "Continue" option
		if m.menuOptions[m.menuCursor] == "Continue" && !m.hasValidSave {
			m.menuCursor--
			if m.menuCursor < 0 {
				m.menuCursor = len(m.menuOptions) - 1
			}
		}
	case "down", "j", "s":
		m.menuCursor++
		if m.menuCursor >= len(m.menuOptions) {
			m.menuCursor = 0
		}
		// Skip disabled "Continue" option
		if m.menuOptions[m.menuCursor] == "Continue" && !m.hasValidSave {
			m.menuCursor++
			if m.menuCursor >= len(m.menuOptions) {
				m.menuCursor = 0
			}
		}
	case "enter", " ":
		return m.selectMenuItem()
	case "q":
		return m, m.requestQuit()
	}
	return m, nil
}

// selectMenuItem handles menu selection.
func (m *Model) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.menuOptions[m.menuCursor] {
	case "New Game":
		// Check if they have an in-progress daily run
		if m.dailyRunInProgress && m.pendingSave != nil {
			depth := m.pendingSave.CurrentDepth
			m.showConfirmDialog(
				fmt.Sprintf("You have a daily run in progress (Floor %d).\nStarting a new game will submit your current progress.\n\nContinue?", depth),
				func() {
					// Submit the abandoned daily run
					// Capture pendingSave before clearing to avoid race condition
					saveData := m.pendingSave
					if m.submitDailyCallback != nil && saveData != nil {
						go func() { _ = m.submitDailyCallback(saveData) }()
					}
					// Clear the in-progress state
					m.dailyRunInProgress = false
					m.dailyRunCompleted = true // They now have a score for today
					m.pendingSave = nil
					m.hasValidSave = false
					// Proceed to class select
					m.classCursor = 0
					m.currentView = ViewClassSelect
				},
				ViewMainMenu,
			)
			return m, nil
		}
		m.classCursor = 0
		m.currentView = ViewClassSelect
		return m, nil
	case "Daily Run":
		// Daily runs are SSH-only (competitive feature requiring leaderboard)
		if !m.isMultiplayer {
			m.statusMsg = "Daily runs require SSH connection"
			return m, nil
		}
		// Block if already completed today's daily
		if m.dailyRunCompleted {
			m.statusMsg = "You've already completed today's daily run"
			return m, nil
		}
		// Redirect if in-progress daily exists
		if m.dailyRunInProgress {
			m.statusMsg = "You have a daily run in progress - use Continue"
			return m, nil
		}
		m.classCursor = 0
		m.dailyRunMode = true
		m.currentView = ViewClassSelect
		return m, nil
	case "Continue":
		// Prevent selecting if no valid save
		if !m.hasValidSave {
			m.statusMsg = "No save file found"
			return m, nil
		}
		m.continueGame()
		return m, nil
	case "Leaderboard":
		m.openLeaderboard()
		return m, nil
	case "Daily Leaderboard":
		// Daily leaderboards are SSH-only
		if !m.isMultiplayer {
			m.statusMsg = "Daily leaderboards require SSH connection"
			return m, nil
		}
		m.showDailyLeaderboard()
		return m, nil
	case "Unlocks":
		m.openUnlockShop()
		return m, nil
	case "How to Play":
		// Show intro sequence without starting a game
		m.introFrame = 0
		m.introSkipped = false
		m.introFromMenu = true // Flag to return to menu instead of starting game
		m.currentView = ViewIntro
		return m, nil
	case "Quit":
		return m, m.requestQuit()
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
		// Check if class is unlocked
		selectedClass := m.classOptions[m.classCursor]
		if !m.isClassUnlocked(selectedClass) {
			m.statusMsg = fmt.Sprintf("Class '%s' is locked. Visit Unlocks from the main menu to purchase it.", selectedClass)
			return m, nil
		}
		// Start intro sequence, then game (user-paced, no auto-advance)
		m.pendingClass = selectedClass
		m.introFrame = 0
		m.introSkipped = false
		m.introFromMenu = false // This leads to a game, not back to menu
		m.currentView = ViewIntro
		return m, nil
	case "esc", "q":
		m.currentView = ViewMainMenu
	}
	return m, nil
}

// continueGame loads and continues a saved game.
func (m *Model) continueGame() {
	// Create engine if needed
	if m.engine == nil {
		m.engine = game.NewEngine(m.config, 0)

		// Set up the dungeon generator
		dungeonCfg := dungeon.DefaultConfig()
		dungeonCfg.Width = m.config.Display.MapWidth
		dungeonCfg.Height = m.config.Display.MapHeight
		m.engine.SetGenerator(dungeon.NewGenerator(dungeonCfg))
	}

	// Set unlocked items for loot pool (for newly generated floors when descending)
	if m.metaProgress != nil {
		m.engine.SetUnlockedItems(m.metaProgress.UnlockedItems)
	}

	// Load save data - prefer pending save from DB (multiplayer), fallback to local files
	if m.pendingSave != nil {
		// Load from database save (multiplayer mode)
		if err := m.engine.LoadGame(m.pendingSave); err != nil {
			m.statusMsg = err.Error()
			return
		}
		// Restore run type from save (so daily runs stay as daily runs after continue)
		if m.pendingSave.RunType != "" {
			m.currentRunType = m.pendingSave.RunType
		} else {
			m.currentRunType = "standard"
		}
		// Restore time tracking from save
		m.runStartTime = m.pendingSave.RunStartTime
		m.elapsedSeconds = m.pendingSave.ElapsedSeconds
		// If no run start time in old save, use now as fallback
		if m.runStartTime.IsZero() {
			m.runStartTime = time.Now().UTC()
		}
	} else {
		// Try to load from local files (single player mode)
		if err := m.engine.LoadLatestSave(); err != nil {
			m.statusMsg = err.Error()
			return
		}
		m.currentRunType = "standard" // Local saves don't track run type
		// Local saves don't have time tracking, start fresh
		m.runStartTime = time.Now().UTC()
		m.elapsedSeconds = 0
	}

	// Start a new session timer
	m.sessionStartTime = time.Now().UTC()

	// Get the player from the engine
	m.player = m.engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring
	m.statusMsg = fmt.Sprintf("Welcome back to %s.", m.engine.CurrentFloorType().FloorName())
}

// startNewGame initializes a new game with the selected class.
func (m *Model) startNewGame(playerClass entity.PlayerClass) {
	// Determine the seed: use daily seed if in daily run mode, otherwise random (0)
	var seed int64
	if m.dailyRunMode {
		seed = getDailySeed()
		m.dailyRunMode = false // Reset for next game
	}

	// Create the game engine
	m.engine = game.NewEngine(m.config, seed)

	// Set up the dungeon generator
	dungeonCfg := dungeon.DefaultConfig()
	dungeonCfg.Width = m.config.Display.MapWidth
	dungeonCfg.Height = m.config.Display.MapHeight
	m.engine.SetGenerator(dungeon.NewGenerator(dungeonCfg))

	// Build permanent bonuses from meta progress
	bonuses := entity.PermanentBonuses{}
	if m.metaProgress != nil {
		bonuses.RAM = m.metaProgress.PermanentBonuses.RAM
		bonuses.CPU = m.metaProgress.PermanentBonuses.CPU
		bonuses.FD = m.metaProgress.PermanentBonuses.FD
		bonuses.NICE = m.metaProgress.PermanentBonuses.NICE

		// Set unlocked items for loot pool (must be set before StartNewGame generates floors)
		m.engine.SetUnlockedItems(m.metaProgress.UnlockedItems)
	}

	// Start a new game with the selected class and bonuses
	if err := m.engine.StartNewGameWithBonuses(playerClass, bonuses); err != nil {
		m.statusMsg = fmt.Sprintf("Failed to start game: %v", err)
		m.currentView = ViewMainMenu
		return
	}

	// Reset run exit codes tracking
	m.runExitCodesEarned = 0
	m.runExitCodesBreakdown = nil

	// Initialize time tracking for this new run
	now := time.Now().UTC()
	m.runStartTime = now
	m.sessionStartTime = now
	m.elapsedSeconds = 0

	// Get the player from the engine
	m.player = m.engine.Player()
	m.currentView = ViewGame
	m.gameState = types.StateExploring

	// Set run type and status message
	if seed != 0 {
		m.currentRunType = "daily"
		dateStr := time.Now().UTC().Format("2006-01-02")
		m.statusMsg = fmt.Sprintf("DAILY RUN (%s) - Spawned as %s. Welcome to %s.", dateStr, playerClass, m.engine.CurrentFloorType().FloorName())
	} else {
		m.currentRunType = "standard"
		m.statusMsg = fmt.Sprintf("Spawned as %s. Welcome to %s.", playerClass, m.engine.CurrentFloorType().FloorName())
	}
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
		m.saveGameAndReturnToMenu()
	case "`":
		// Open admin console - disabled in multiplayer to prevent cheating
		if m.isMultiplayer {
			m.statusMsg = "Admin console disabled in multiplayer mode."
			return m, nil
		}
		// Requires root (UID 0) OR actual sudo
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

	// Wall bump feedback (replace generic message with thematic one)
	if !result.Moved && len(result.Combat) == 0 && result.Message == "You cannot walk there." {
		m.statusMsg = "You bump into a wall."
	}

	// Trigger exploration messages when the player actually moved
	if result.Moved && m.engine.RNG() != nil {
		rng := m.engine.RNG()
		roll := rng.Float64()

		// Single roll for all message types to avoid consuming extra RNG
		if roll < 0.01 {
			// Easter egg messages: ~1% chance
			msg := content.EasterEggs[rng.Intn(len(content.EasterEggs))]
			m.addToHistory(msg)
		} else if roll < 0.06 {
			// Exploration messages: ~5% chance
			msg := content.ExplorationMessages[rng.Intn(len(content.ExplorationMessages))]
			m.addToHistory(msg)
		} else if roll < 0.09 {
			// Random encounter messages: ~3% chance
			msg := content.RandomEncounterMessages[rng.Intn(len(content.RandomEncounterMessages))]
			m.addToHistory(msg)
		}
	}

	// Check for combat initiation
	if len(result.Combat) > 0 {
		m.startCombat(result.Combat)
	}
}

// updatePause handles pause menu input.
func (m *Model) updatePause(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "p":
		m.currentView = ViewGame
	case "q":
		m.saveGameAndReturnToMenu()
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
		return m, m.requestQuit()
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
	case "esc", "q":
		// Skip intro entirely
		m.introSkipped = true
		if m.introFromMenu {
			m.introFromMenu = false
			m.currentView = ViewMainMenu
			return m, nil
		}
		m.startNewGame(m.pendingClass)
		return m, nil
	case "enter", " ", "right", "down", "d", "l", "j":
		// Advance to next frame
		m.introFrame++
		if m.introFrame >= len(introFrames) {
			if m.introFromMenu {
				m.introFromMenu = false
				m.currentView = ViewMainMenu
				return m, nil
			}
			m.startNewGame(m.pendingClass)
			return m, nil
		}
		return m, nil
	case "left", "up", "a", "h", "k":
		// Go back to previous frame
		if m.introFrame > 0 {
			m.introFrame--
		}
		return m, nil
	}
	return m, nil
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
		// Extra safeguard: godMode disabled in multiplayer
		if m.isMultiplayer {
			m.statusMsg = "[ADMIN] God mode disabled in multiplayer."
			return m, nil
		}
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
			if forceErr := m.engine.ForceDescend(); forceErr != nil {
				m.statusMsg = fmt.Sprintf("[ADMIN] Force descent failed: %v", forceErr)
			} else {
				m.statusMsg = fmt.Sprintf("[ADMIN] Forced descent to depth %d", m.engine.CurrentDepth())
			}
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

// --- Exit Code Calculation ---

// calculateRunExitCodes calculates exit codes earned for a run.
func (m *Model) calculateRunExitCodes(victory bool) (int, []string) {
	total := 0
	var breakdown []string

	if m.engine == nil {
		return 0, breakdown
	}

	stats := m.engine.GetRunStats()
	if stats == nil {
		return 0, breakdown
	}

	// Floors cleared: 10 per floor
	floorsBonus := stats.MaxDepthReached * 10
	total += floorsBonus
	breakdown = append(breakdown, fmt.Sprintf("Floors cleared (%d): +%d", stats.MaxDepthReached, floorsBonus))

	// Enemies killed: 1-5 per enemy type
	for enemyType, count := range stats.EnemiesKilled {
		value := m.getEnemyExitCodeValue(enemyType)
		killBonus := value * count
		total += killBonus
		if killBonus > 0 {
			breakdown = append(breakdown, fmt.Sprintf("%s x%d: +%d", enemyType, count, killBonus))
		}
	}

	// Boss kill bonus (if killed kernel_panic)
	if _, killed := stats.EnemiesKilled["kernel_panic"]; killed {
		bossBonus := 50
		total += bossBonus
		breakdown = append(breakdown, fmt.Sprintf("Boss defeated: +%d", bossBonus))
	}

	// Victory bonus
	if victory {
		victoryBonus := 200
		total += victoryBonus
		breakdown = append(breakdown, fmt.Sprintf("Victory bonus: +%d", victoryBonus))
	}

	return total, breakdown
}

// getEnemyExitCodeValue returns the exit code value for killing an enemy type.
func (m *Model) getEnemyExitCodeValue(enemyType string) int {
	switch enemyType {
	case "zombie":
		return 1
	case "daemon":
		return 2
	case "fork_bomb":
		return 3
	case "segfault":
		return 4
	case "rootkit":
		return 5
	case "kernel_panic":
		return 10
	default:
		return 1
	}
}

// awardRunExitCodes awards exit codes at end of run and updates meta progress.
func (m *Model) awardRunExitCodes(victory bool) {
	earned, breakdown := m.calculateRunExitCodes(victory)
	m.runExitCodesEarned = earned
	m.runExitCodesBreakdown = breakdown

	// Get max depth for multiplayer callback
	maxDepth := 0
	if m.engine != nil {
		stats := m.engine.GetRunStats()
		if stats != nil {
			maxDepth = stats.MaxDepthReached
		}
	}

	// Use multiplayer callback if available (server handles persistence)
	if m.metaProgressUpdater != nil {
		// Error logged in server callback - exit codes still displayed this session
		_ = m.metaProgressUpdater(earned, victory, maxDepth)
		return
	}

	// Fallback to local save for single-player mode
	if m.metaProgress == nil || m.saveManager == nil {
		return
	}

	// Update meta progress
	m.metaProgress.TotalExitCodes += earned

	if victory {
		m.metaProgress.RunsCompleted++
	} else {
		m.metaProgress.TotalDeaths++
	}

	// Update deepest floor
	if maxDepth > m.metaProgress.DeepestFloor {
		m.metaProgress.DeepestFloor = maxDepth
	}

	// Save meta progress
	_ = m.saveManager.SaveMetaProgress(m.metaProgress)
}

// finishRun handles all end-of-run logic for both death and victory.
// This centralizes exit codes, leaderboard submission, and save cleanup.
func (m *Model) finishRun(victory bool) {
	m.awardRunExitCodes(victory)
	m.submitToLeaderboard(victory)
	m.clearSaveOnRunEnd()

	if victory {
		m.currentView = ViewVictory
		m.gameState = types.StateVictory
		m.statusMsg = "KERNEL PANIC DEFEATED! You saved the system!"
	} else {
		m.currentView = ViewGameOver
		m.gameState = types.StateGameOver
	}
}

// clearSaveOnRunEnd clears the save state when a run ends (death or victory).
// This prevents the Continue button from loading a dead game.
func (m *Model) clearSaveOnRunEnd() {
	// Clear in-memory state
	m.hasValidSave = false
	m.pendingSave = nil

	// Use multiplayer callback if available (server handles DB deletion)
	if m.clearSaveCallback != nil {
		// Error logged in server callback - the important thing is hasValidSave is false
		_ = m.clearSaveCallback()
		return
	}

	// Fallback to local save deletion for single-player mode
	if m.saveManager != nil && m.engine != nil {
		seed := m.engine.MasterSeed()
		if seed != 0 {
			_ = m.saveManager.DeleteSave(seed) // Best-effort delete
		}
	}
}

// saveGameAndReturnToMenu saves the game and returns to the main menu.
// Works for both multiplayer (saveCallback) and offline (saveManager) modes.
func (m *Model) saveGameAndReturnToMenu() {
	if m.engine == nil {
		return
	}

	// Try multiplayer save callback first
	if m.saveCallback != nil {
		if savedData, err := m.saveCallback(); err != nil {
			m.statusMsg = "Failed to save game."
		} else {
			m.statusMsg = "Game saved."
			m.hasValidSave = true
			m.pendingSave = savedData
		}
	} else if m.saveManager != nil {
		// Offline mode - save to local files
		saveData := m.GetSaveData()
		if saveData != nil {
			if err := m.saveManager.SaveSync(saveData, save.TriggerManual); err != nil {
				m.statusMsg = "Failed to save game."
			} else {
				m.statusMsg = "Game saved."
				m.hasValidSave = true
			}
		}
	} else {
		m.statusMsg = "Returned to menu."
	}

	m.engine.Shutdown()
	m.engine = nil
	m.player = nil
	m.currentView = ViewMainMenu
	m.gameState = types.StateMainMenu
}

// updateConfirmDialog handles confirmation dialog input.
func (m *Model) updateConfirmDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.confirmAction != nil {
			m.confirmAction()
		}
		return m, nil
	case "n", "N", "esc", "q":
		m.currentView = m.confirmReturnView
		return m, nil
	}
	return m, nil
}
