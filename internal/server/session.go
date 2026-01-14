package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	mu          sync.Mutex
}

// SessionManager tracks active game sessions.
type SessionManager struct {
	sessions map[string]*GameSession // fingerprint -> session
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
	gameSession.Model = model

	// If we have a save, restore it
	if gameSave != nil {
		var saveData save.SaveData
		if err := json.Unmarshal(gameSave.SaveData, &saveData); err != nil {
			log.Error("Failed to unmarshal save data", "error", err)
		} else {
			if engine := model.GetEngine(); engine != nil {
				if err := engine.LoadGame(&saveData); err != nil {
					log.Error("Failed to load game state", "error", err)
				} else {
					log.Info("Restored game from save", "user", user.Username)
				}
			}
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
	}

	return wrapper, []tea.ProgramOption{tea.WithAltScreen()}
}

// sessionWrapper wraps the game UI to handle session lifecycle.
type sessionWrapper struct {
	model       *ui.Model
	server      *Server
	session     *GameSession
	fingerprint string
}

func (sw *sessionWrapper) Init() tea.Cmd {
	return sw.model.Init()
}

func (sw *sessionWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle quit to save game
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			// Save before quitting
			ctx := context.Background()
			if err := saveSessionToDatabase(ctx, sw.server.db, sw.session); err != nil {
				log.Error("Failed to save on quit", "error", err)
			}
			sw.server.sessions.Remove(sw.fingerprint)
			return sw, tea.Quit
		}
	}

	// Delegate to wrapped model
	newModel, cmd := sw.model.Update(msg)
	if m, ok := newModel.(*ui.Model); ok {
		sw.model = m
		sw.session.Model = m
	}

	// Check if the model returned quit
	if cmd != nil {
		// If quitting, save first
		// Note: We could inspect the command here if needed
	}

	return sw, cmd
}

func (sw *sessionWrapper) View() string {
	return sw.model.View()
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
	}

	model := ui.NewWithRenderer(cfg, renderer)
	model.SetMultiplayerMode(user.Username) // Disable cheats in multiplayer
	gameSession.Model = model

	s.sessions.Add(fingerprint, gameSession)

	wrapper := &sessionWrapper{
		model:       model,
		server:      s,
		session:     gameSession,
		fingerprint: fingerprint,
	}

	return wrapper, nil
}
