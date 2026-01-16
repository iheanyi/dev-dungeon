package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/db"
	"github.com/iheanyi/devdungeon/internal/monitoring"
	"github.com/iheanyi/devdungeon/internal/save"
	"github.com/iheanyi/devdungeon/internal/ui"
)

// Database operation timeouts
const (
	dbLoadTimeout   = 10 * time.Second
	dbSaveTimeout   = 10 * time.Second
	dbCreateTimeout = 5 * time.Second
)

// GameSession represents an active player session.
type GameSession struct {
	User        *db.User
	Fingerprint string
	Model       *ui.Model
	Session     ssh.Session // SSH session for direct writes
	mu          sync.Mutex
}

// SessionManager tracks active game sessions.
type SessionManager struct {
	sessions map[string]*GameSession // fingerprint -> session
	wg       sync.WaitGroup          // tracks auto-save goroutines
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*GameSession),
	}
}

// Add registers a new session.
func (sm *SessionManager) Add(fingerprint string, session *GameSession) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[fingerprint] = session
}

// Remove unregisters a session.
func (sm *SessionManager) Remove(fingerprint string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, fingerprint)
}

// Get retrieves a session by fingerprint.
func (sm *SessionManager) Get(fingerprint string) *GameSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[fingerprint]
}

// SaveAll saves all active sessions to the database.
func (sm *SessionManager) SaveAll(ctx context.Context, dbClient *db.Client) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for fp, session := range sm.sessions {
		if session.User != nil && session.Model != nil {
			// Get save data from the model's engine
			if err := saveSessionToDatabase(ctx, dbClient, session); err != nil {
				log.Error("Failed to save session", "fingerprint", fp, "error", err)
				monitoring.CaptureException(err, map[string]string{
					"operation":   "save_all_sessions",
					"fingerprint": fp,
				})
			}
		}
	}
}

// WaitForAutoSaves waits for all auto-save goroutines to complete.
// Call this during shutdown before closing the database.
func (sm *SessionManager) WaitForAutoSaves() {
	sm.wg.Wait()
}

// NotifyShutdown sends a shutdown message to all connected sessions.
// Call this before shutting down the SSH server to inform users.
func (sm *SessionManager) NotifyShutdown(message string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionCount := len(sm.sessions)
	log.Info("Notifying connected sessions of shutdown", "session_count", sessionCount, "message", message)

	if sessionCount == 0 {
		log.Info("No active sessions to notify")
		return
	}

	// ANSI escape codes for styling
	const (
		reset  = "\033[0m"
		bold   = "\033[1m"
		yellow = "\033[33m"
		green  = "\033[32m"
		cyan   = "\033[36m"
	)

	// Pad message to fixed width (46 chars for content area)
	paddedMsg := message
	if len(paddedMsg) < 46 {
		paddedMsg += strings.Repeat(" ", 46-len(paddedMsg))
	} else if len(paddedMsg) > 46 {
		paddedMsg = paddedMsg[:46]
	}

	// Format message with terminal styling
	// Box drawing with Unicode - emojis count as 2 display width
	formattedMsg := "\r\n\r\n"
	formattedMsg += fmt.Sprintf("%s%s╔══════════════════════════════════════════════════╗%s\r\n", bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s  %s🔄 SERVER RESTARTING%s                            %s%s║%s\r\n", bold, yellow, reset, bold, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s╠══════════════════════════════════════════════════╣%s\r\n", bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s                                                  %s%s║%s\r\n", bold, yellow, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s  %s%s%s  %s%s║%s\r\n", bold, yellow, reset, green, paddedMsg, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s                                                  %s%s║%s\r\n", bold, yellow, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s  %s💾 Your progress has been saved.%s                %s%s║%s\r\n", bold, yellow, reset, cyan, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s  %s🔗 Reconnect: ssh dev-dungeon.com%s               %s%s║%s\r\n", bold, yellow, reset, cyan, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s║%s                                                  %s%s║%s\r\n", bold, yellow, reset, bold, yellow, reset)
	formattedMsg += fmt.Sprintf("%s%s╚══════════════════════════════════════════════════╝%s\r\n\r\n", bold, yellow, reset)

	notified := 0
	for fingerprint, session := range sm.sessions {
		if session.Session != nil {
			// Write directly to SSH session - ignore errors as session may be closing
			if _, err := session.Session.Write([]byte(formattedMsg)); err != nil {
				log.Warn("Failed to notify session of shutdown", "fingerprint", fingerprint[:16]+"...", "error", err)
			} else {
				notified++
				username := "unknown"
				if session.User != nil {
					username = session.User.Username
				}
				log.Info("Notified session of shutdown", "user", username)
			}
		}
	}
	log.Info("Shutdown notification complete", "notified", notified, "total", sessionCount)
}

// saveSessionToDatabase persists the game state.
func saveSessionToDatabase(ctx context.Context, dbClient *db.Client, session *GameSession) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.Model == nil || session.User == nil {
		return nil
	}

	// Get the current save data from the engine
	engine := session.Model.GetEngine()
	if engine == nil {
		return nil
	}

	saveData := engine.GetSaveData()
	if saveData == nil {
		return nil
	}

	return dbClient.UpsertGameSave(ctx, session.User.ID, saveData)
}

// newGameSession creates a new Bubble Tea model for a game session.
func (s *Server) newGameSession(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	ctx := sess.Context()

	// Safe type assertions with validation
	fingerprint, ok := ctx.Value("fingerprint").(string)
	if !ok || fingerprint == "" {
		log.Error("Missing or invalid fingerprint in session context")
		return nil, nil
	}

	needsRegistration, _ := ctx.Value("needs_registration").(bool)

	// Get PTY info
	pty, _, _ := sess.Pty()

	// Create renderer for this session
	renderer := bubbletea.MakeRenderer(sess)

	if needsRegistration {
		// Show registration flow
		return s.newRegistrationModel(sess, fingerprint), []tea.ProgramOption{tea.WithAltScreen()}
	}

	// Get existing user from context - must be present for authenticated sessions
	userVal := ctx.Value("user")
	if userVal == nil {
		log.Error("No user in context after auth")
		return nil, nil
	}
	user, ok := userVal.(*db.User)
	if !ok || user == nil {
		log.Error("Invalid user type in context", "type", fmt.Sprintf("%T", userVal))
		return nil, nil
	}

	// Load game config
	cfg := config.DefaultConfig()
	cfg.Display.MapWidth = pty.Window.Width - 30  // Leave room for stats panel
	cfg.Display.MapHeight = pty.Window.Height - 8 // Leave room for messages

	// Create game session
	gameSession := &GameSession{
		User:        user,
		Fingerprint: fingerprint,
		Session:     sess,
	}

	// Try to load existing save from database
	loadCtx, loadCancel := context.WithTimeout(context.Background(), dbLoadTimeout)
	defer loadCancel()
	gameSave, err := s.db.GetGameSave(loadCtx, user.ID)
	if err != nil {
		log.Error("Failed to load save", "error", err)
	}

	// Create UI model
	model := ui.NewWithRenderer(cfg, renderer)
	model.SetMultiplayerMode(user.Username) // Disable cheats in multiplayer

	// Set up leaderboard fetcher
	model.SetLeaderboardFetcher(func(runType string, limit int) ([]ui.LeaderboardEntry, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Convert "all" to empty string for the database query
		dbRunType := runType
		if runType == "all" {
			dbRunType = ""
		}

		dbEntries, err := s.db.GetLeaderboard(ctx, dbRunType, limit)
		if err != nil {
			return nil, err
		}

		// Convert db entries to ui entries
		entries := make([]ui.LeaderboardEntry, len(dbEntries))
		for i, e := range dbEntries {
			entries[i] = ui.LeaderboardEntry{
				Rank:          i + 1,
				Username:      e.Username,
				Score:         e.Score,
				FloorsCleared: e.FloorsCleared,
				Class:         e.Class,
				Seed:          e.Seed,
				RunType:       e.RunType,
			}
		}
		return entries, nil
	})

	// Set up save callback - this is called when user presses Q to return to menu
	model.SetSaveCallback(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), dbSaveTimeout)
		defer cancel()
		return saveSessionToDatabase(ctx, s.db, gameSession)
	})

	// Set up leaderboard submitter - called on death or victory
	model.SetLeaderboardSubmitter(func(score, floorsCleared int, class string, seed int64, runType string, victory bool) error {
		ctx, cancel := context.WithTimeout(context.Background(), dbSaveTimeout)
		defer cancel()

		entry := &db.LeaderboardEntry{
			UserID:        user.ID,
			Username:      user.Username,
			Score:         score,
			FloorsCleared: floorsCleared,
			Class:         class,
			Seed:          seed,
			RunType:       runType,
		}

		if err := s.db.AddLeaderboardEntry(ctx, entry); err != nil {
			log.Error("Failed to submit leaderboard entry", "error", err, "user", user.Username)
			return err
		}

		log.Info("Leaderboard entry submitted",
			"user", user.Username,
			"score", score,
			"floors", floorsCleared,
			"class", class,
			"runType", runType,
			"victory", victory)
		return nil
	})

	gameSession.Model = model

	// Set up daily leaderboard fetcher (date-navigable)
	model.SetDailyLeaderboardFetcher(func(date time.Time, limit int, _ int) ([]ui.LeaderboardEntry, int, *ui.LeaderboardEntry, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get top entries for this date
		dbEntries, _, err := s.db.GetDailyLeaderboard(ctx, date, limit, nil)
		if err != nil {
			return nil, 0, nil, err
		}

		// Convert db entries to ui entries
		var entries []ui.LeaderboardEntry
		for _, e := range dbEntries {
			entries = append(entries, ui.LeaderboardEntry{
				Rank:          e.Rank,
				Username:      e.Username,
				Score:         e.Score,
				FloorsCleared: e.FloorsCleared,
				Class:         e.Class,
				Seed:          e.Seed,
				RunType:       e.RunType,
			})
		}

		// Get player's rank for this date
		playerRank, dbPlayerEntry, err := s.db.GetPlayerDailyRank(ctx, date, user.ID)
		if err != nil {
			return entries, 0, nil, err
		}

		var playerEntry *ui.LeaderboardEntry
		if dbPlayerEntry != nil {
			playerEntry = &ui.LeaderboardEntry{
				Rank:          playerRank,
				Username:      dbPlayerEntry.Username,
				Score:         dbPlayerEntry.Score,
				FloorsCleared: dbPlayerEntry.FloorsCleared,
				Class:         dbPlayerEntry.Class,
				Seed:          dbPlayerEntry.Seed,
				RunType:       dbPlayerEntry.RunType,
			}
		}

		return entries, playerRank, playerEntry, nil
	})

	// If we have a save, store it for loading when user selects Continue
	if gameSave != nil {
		var saveData save.SaveData
		if err := json.Unmarshal(gameSave.SaveData, &saveData); err != nil {
			log.Error("Failed to unmarshal save data", "error", err)
		} else {
			// Pass save data to model - it will be loaded when Continue is selected
			model.SetHasValidSave(true, &saveData)
			log.Info("Save data available for continue", "user", user.Username)
		}
	}

	// Register session
	s.sessions.Add(fingerprint, gameSession)

	// Create wrapper model that handles saves on quit
	wrapper := &sessionWrapper{
		model:       model,
		server:      s,
		session:     gameSession,
		fingerprint: fingerprint,
		sessCtx:     ctx, // SSH session context for disconnect detection
	}

	return wrapper, []tea.ProgramOption{tea.WithAltScreen()}
}

// Auto-save interval - aggressive to minimize progress loss on disconnect/deploy
const autoSaveInterval = 5 * time.Second

// sessionWrapper wraps the game UI to handle session lifecycle.
type sessionWrapper struct {
	model        *ui.Model
	server       *Server
	session      *GameSession
	fingerprint  string
	sessCtx      context.Context // SSH session context - cancelled on disconnect
	showingLink  bool            // Whether the link modal is currently displayed
	linkURL      string          // The generated magic link URL
	linkError    string          // Error message if link generation failed
	stopAutoSave chan struct{}
	autoSaveDone chan struct{}
}

func (sw *sessionWrapper) Init() tea.Cmd {
	// Start background auto-save goroutine
	sw.stopAutoSave = make(chan struct{})
	sw.autoSaveDone = make(chan struct{})

	// Track goroutine in WaitGroup for clean shutdown
	sw.server.sessions.wg.Add(1)
	go sw.autoSaveLoop()

	return sw.model.Init()
}

// autoSaveLoop runs in a background goroutine and saves periodically.
func (sw *sessionWrapper) autoSaveLoop() {
	ticker := time.NewTicker(autoSaveInterval)
	defer ticker.Stop()
	defer close(sw.autoSaveDone)
	defer sw.server.sessions.wg.Done() // Signal goroutine completion

	for {
		select {
		case <-ticker.C:
			sw.performAutoSave()
		case <-sw.stopAutoSave:
			// Final save before clean shutdown
			sw.performAutoSave()
			return
		case <-sw.sessCtx.Done():
			// SSH session ended (disconnect, network drop, etc.)
			// Perform final save and clean up
			log.Info("SSH session ended, performing final save", "user", sw.session.User.Username)
			sw.performAutoSave()
			sw.server.sessions.Remove(sw.fingerprint)
			return
		}
	}
}

// performAutoSave saves the game state in the background.
func (sw *sessionWrapper) performAutoSave() {
	if sw.session.User == nil || sw.session.Model == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbSaveTimeout)
	defer cancel()

	if err := saveSessionToDatabase(ctx, sw.server.db, sw.session); err != nil {
		log.Error("Auto-save failed", "error", err, "user", sw.session.User.Username)
	} else {
		log.Debug("Auto-saved game", "user", sw.session.User.Username)
	}
}

// stopAutoSaveGoroutine signals the auto-save goroutine to stop and waits for it.
func (sw *sessionWrapper) stopAutoSaveGoroutine() {
	if sw.stopAutoSave != nil {
		close(sw.stopAutoSave)
		<-sw.autoSaveDone // Wait for final save to complete
	}
}

func (sw *sessionWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle link modal dismissal first
	if sw.showingLink {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Any key dismisses the link modal
			if msg.String() == "enter" || msg.String() == "esc" || msg.String() == " " {
				sw.showingLink = false
				sw.linkURL = ""
				sw.linkError = ""
				return sw, nil
			}
		}
		return sw, nil
	}

	// Handle quit and link commands
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Stop auto-save goroutine (triggers final save)
			sw.stopAutoSaveGoroutine()
			sw.server.sessions.Remove(sw.fingerprint)
			return sw, tea.Quit

		case "ctrl+l":
			// Generate magic link for browser authentication
			sw.generateMagicLink()
			return sw, nil
		}
	}

	// Delegate to wrapped model
	newModel, cmd := sw.model.Update(msg)
	if m, ok := newModel.(*ui.Model); ok {
		sw.model = m
		sw.session.Model = m

		// Check if the model wants to quit - save before exiting
		if m.WantsToQuit() {
			// Stop auto-save goroutine (triggers final save)
			sw.stopAutoSaveGoroutine()
			sw.server.sessions.Remove(sw.fingerprint)
			m.ClearQuitRequest()
		}
	}

	return sw, cmd
}

// generateMagicLink creates a magic link token and displays it to the user.
func (sw *sessionWrapper) generateMagicLink() {
	if sw.session.User == nil {
		sw.showingLink = true
		sw.linkError = "You must be logged in to generate a link"
		return
	}

	if sw.server.config.WebBaseURL == "" {
		sw.showingLink = true
		sw.linkError = "Web portal not configured"
		return
	}

	// Generate token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := sw.server.db.CreateAuthToken(ctx, sw.session.User.ID)
	if err != nil {
		log.Error("Failed to generate magic link", "error", err, "user", sw.session.User.Username)
		sw.showingLink = true
		sw.linkError = "Failed to generate link. Try again."
		return
	}

	// Build URL
	sw.linkURL = fmt.Sprintf("%s/auth/verify?token=%s", sw.server.config.WebBaseURL, token)
	sw.showingLink = true
	sw.linkError = ""

	log.Info("Magic link generated", "user", sw.session.User.Username)
}

func (sw *sessionWrapper) View() string {
	// Show link modal overlay if active
	if sw.showingLink {
		return sw.viewLinkModal()
	}
	return sw.model.View()
}

// viewLinkModal renders the magic link modal.
func (sw *sessionWrapper) viewLinkModal() string {
	var b strings.Builder

	// Box styling with lipgloss
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")). // Cyan
		Padding(0, 1)

	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")) // Red

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Gray

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2)

	b.WriteString("\n\n")

	var content strings.Builder
	content.WriteString(titleStyle.Render("🔐 BROWSER AUTHENTICATION"))
	content.WriteString("\n\n")

	if sw.linkError != "" {
		content.WriteString(errorStyle.Render("ERROR: " + sw.linkError))
	} else {
		content.WriteString("Click or copy this link to authenticate in your browser:\n\n")

		// OSC-8 hyperlink format: \x1b]8;;URL\x07DISPLAY_TEXT\x1b]8;;\x07
		// This makes the URL clickable in supported terminals (iTerm2, Windows Terminal, etc.)
		// IMPORTANT: Don't wrap with lipgloss - it breaks the OSC-8 escape sequence detection
		// Apply color INSIDE the hyperlink using raw ANSI codes
		colorStart := "\x1b[1;38;5;117m" // Bold + color 117 (light blue)
		colorEnd := "\x1b[0m"
		hyperlink := fmt.Sprintf("\x1b]8;;%s\x07%s%s%s\x1b]8;;\x07", sw.linkURL, colorStart, sw.linkURL, colorEnd)
		content.WriteString(hyperlink)

		content.WriteString("\n\n")
		content.WriteString(hintStyle.Render("⏱  Expires in 5 minutes • Single use only"))
	}

	content.WriteString("\n\n")
	content.WriteString(hintStyle.Render("[Enter] or [Esc] to close"))

	b.WriteString(boxStyle.Render(content.String()))
	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("  Tip: Press Ctrl+L anytime to generate a new link"))
	b.WriteString("\n")

	return b.String()
}

// registrationModel handles new player registration.
type registrationModel struct {
	server      *Server
	sess        ssh.Session
	fingerprint string
	username    string
	cursor      int
	err         string
	done        bool
}

func (s *Server) newRegistrationModel(sess ssh.Session, fingerprint string) *registrationModel {
	return &registrationModel{
		server:      s,
		sess:        sess,
		fingerprint: fingerprint,
	}
}

func (m *registrationModel) Init() tea.Cmd {
	return nil
}

func (m *registrationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if len(m.username) >= 3 {
				// Try to create user with timeout
				createCtx, cancel := context.WithTimeout(context.Background(), dbCreateTimeout)
				user, err := m.server.db.CreateUser(createCtx, m.username, m.fingerprint)
				cancel()
				if err != nil {
					m.err = "Registration failed - please try again"
				} else {
					// Success! Store user and transition to game
					m.sess.Context().SetValue("user", user)
					m.done = true
					log.Info("User registered", "user", user.Username, "fingerprint", m.fingerprint)

					// Create new game session for this user
					return m.server.createGameModel(m.sess, user, m.fingerprint)
				}
			} else {
				m.err = "Username must be at least 3 characters"
			}
		case "backspace":
			if len(m.username) > 0 {
				m.username = m.username[:len(m.username)-1]
			}
			m.err = ""
		default:
			// Only allow lowercase alphanumeric and dashes (SSH-friendly)
			if len(msg.String()) == 1 && len(m.username) < 20 {
				c := msg.String()[0]
				// Convert uppercase to lowercase
				if c >= 'A' && c <= 'Z' {
					c = c + 32 // to lowercase
				}
				// Allow: a-z, 0-9, dash (but dash can't be first char)
				if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (c == '-' && len(m.username) > 0) {
					m.username += string(c)
					m.err = ""
				}
			}
		}
	}
	return m, nil
}

func (m *registrationModel) View() string {
	s := "\n"
	s += "  ╔════════════════════════════════════════╗\n"
	s += "  ║         WELCOME TO /dev/dungeon        ║\n"
	s += "  ╠════════════════════════════════════════╣\n"
	s += "  ║                                        ║\n"
	s += "  ║  New SSH key detected!                 ║\n"
	s += "  ║  Let's create your account.            ║\n"
	s += "  ║                                        ║\n"
	s += fmt.Sprintf("  ║  Username: %-27s ║\n", m.username+"█")
	s += "  ║                                        ║\n"
	if m.err != "" {
		s += fmt.Sprintf("  ║  ERROR: %-30s ║\n", m.err)
	} else {
		s += "  ║  (3-20 chars, lowercase + numbers + -) ║\n"
	}
	s += "  ║                                        ║\n"
	s += "  ║  [Enter] Create   [Esc] Cancel         ║\n"
	s += "  ║                                        ║\n"
	s += "  ╚════════════════════════════════════════╝\n"
	s += "\n"
	s += fmt.Sprintf("  Your key: %s\n", m.fingerprint[:20]+"...")
	return s
}

// createGameModel creates the actual game model after registration.
func (s *Server) createGameModel(sess ssh.Session, user *db.User, fingerprint string) (tea.Model, tea.Cmd) {
	pty, _, _ := sess.Pty()
	renderer := bubbletea.MakeRenderer(sess)

	cfg := config.DefaultConfig()
	cfg.Display.MapWidth = pty.Window.Width - 30
	cfg.Display.MapHeight = pty.Window.Height - 8

	gameSession := &GameSession{
		User:        user,
		Fingerprint: fingerprint,
		Session:     sess,
	}

	model := ui.NewWithRenderer(cfg, renderer)
	model.SetMultiplayerMode(user.Username) // Disable cheats in multiplayer

	// Set up leaderboard fetcher
	model.SetLeaderboardFetcher(func(runType string, limit int) ([]ui.LeaderboardEntry, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Convert "all" to empty string for the database query
		dbRunType := runType
		if runType == "all" {
			dbRunType = ""
		}

		dbEntries, err := s.db.GetLeaderboard(ctx, dbRunType, limit)
		if err != nil {
			return nil, err
		}

		// Convert db entries to ui entries
		entries := make([]ui.LeaderboardEntry, len(dbEntries))
		for i, e := range dbEntries {
			entries[i] = ui.LeaderboardEntry{
				Rank:          i + 1,
				Username:      e.Username,
				Score:         e.Score,
				FloorsCleared: e.FloorsCleared,
				Class:         e.Class,
				Seed:          e.Seed,
				RunType:       e.RunType,
			}
		}
		return entries, nil
	})

	// Set up save callback - this is called when user presses Q to return to menu
	model.SetSaveCallback(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), dbSaveTimeout)
		defer cancel()
		return saveSessionToDatabase(ctx, s.db, gameSession)
	})

	// Set up leaderboard submitter - called on death or victory
	model.SetLeaderboardSubmitter(func(score, floorsCleared int, class string, seed int64, runType string, victory bool) error {
		ctx, cancel := context.WithTimeout(context.Background(), dbSaveTimeout)
		defer cancel()

		entry := &db.LeaderboardEntry{
			UserID:        user.ID,
			Username:      user.Username,
			Score:         score,
			FloorsCleared: floorsCleared,
			Class:         class,
			Seed:          seed,
			RunType:       runType,
		}

		if err := s.db.AddLeaderboardEntry(ctx, entry); err != nil {
			log.Error("Failed to submit leaderboard entry", "error", err, "user", user.Username)
			return err
		}

		log.Info("Leaderboard entry submitted",
			"user", user.Username,
			"score", score,
			"floors", floorsCleared,
			"class", class,
			"runType", runType,
			"victory", victory)
		return nil
	})

	gameSession.Model = model

	// Set up daily leaderboard fetcher (date-navigable)
	model.SetDailyLeaderboardFetcher(func(date time.Time, limit int, _ int) ([]ui.LeaderboardEntry, int, *ui.LeaderboardEntry, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get top entries for this date
		dbEntries, _, err := s.db.GetDailyLeaderboard(ctx, date, limit, nil)
		if err != nil {
			return nil, 0, nil, err
		}

		// Convert db entries to ui entries
		var entries []ui.LeaderboardEntry
		for _, e := range dbEntries {
			entries = append(entries, ui.LeaderboardEntry{
				Rank:          e.Rank,
				Username:      e.Username,
				Score:         e.Score,
				FloorsCleared: e.FloorsCleared,
				Class:         e.Class,
				Seed:          e.Seed,
				RunType:       e.RunType,
			})
		}

		// Get player's rank for this date
		playerRank, dbPlayerEntry, err := s.db.GetPlayerDailyRank(ctx, date, user.ID)
		if err != nil {
			return entries, 0, nil, err
		}

		var playerEntry *ui.LeaderboardEntry
		if dbPlayerEntry != nil {
			playerEntry = &ui.LeaderboardEntry{
				Rank:          playerRank,
				Username:      dbPlayerEntry.Username,
				Score:         dbPlayerEntry.Score,
				FloorsCleared: dbPlayerEntry.FloorsCleared,
				Class:         dbPlayerEntry.Class,
				Seed:          dbPlayerEntry.Seed,
				RunType:       dbPlayerEntry.RunType,
			}
		}

		return entries, playerRank, playerEntry, nil
	})

	s.sessions.Add(fingerprint, gameSession)

	wrapper := &sessionWrapper{
		model:       model,
		server:      s,
		session:     gameSession,
		fingerprint: fingerprint,
		sessCtx:     sess.Context(), // SSH session context for disconnect detection
	}

	// IMPORTANT: Must call Init() to start the auto-save goroutine.
	// When returning a new model from Update(), Bubble Tea does NOT
	// automatically call Init() on it - we must do it explicitly.
	return wrapper, wrapper.Init()
}
