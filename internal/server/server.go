// Package server provides the SSH server for /dev/dungeon multiplayer.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	gossh "golang.org/x/crypto/ssh"

	"github.com/iheanyi/devdungeon/internal/db"
)

// Config holds SSH server configuration.
type Config struct {
	Host        string
	Port        string
	HostKeyPath string
	DatabaseURL string
}

// Server is the SSH game server.
type Server struct {
	config   Config
	db       *db.Client
	sshSrv   *ssh.Server
	sessions *SessionManager
}

// New creates a new SSH game server.
func New(cfg Config) (*Server, error) {
	ctx := context.Background()

	// Connect to database
	dbClient, err := db.NewClient(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	s := &Server{
		config:   cfg,
		db:       dbClient,
		sessions: NewSessionManager(),
	}

	// Create SSH server with Wish
	sshSrv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(cfg.Host, cfg.Port)),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithPublicKeyAuth(s.publicKeyAuth),
		wish.WithMiddleware(
			s.gameMiddleware(),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		dbClient.Close()
		return nil, fmt.Errorf("failed to create SSH server: %w", err)
	}

	s.sshSrv = sshSrv
	return s, nil
}

// Run starts the SSH server and blocks until shutdown.
func (s *Server) Run() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Starting /dev/dungeon SSH server",
		"host", s.config.Host,
		"port", s.config.Port)

	go func() {
		if err := s.sshSrv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Server error", "error", err)
			done <- nil
		}
	}()

	<-done

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Save all active sessions
	s.sessions.SaveAll(ctx, s.db)

	if err := s.sshSrv.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Shutdown error", "error", err)
	}

	s.db.Close()
	log.Info("Server stopped")
	return nil
}

// publicKeyAuth handles SSH public key authentication.
func (s *Server) publicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	fingerprint := gossh.FingerprintSHA256(key)
	username := ctx.User()

	log.Debug("Auth attempt", "user", username, "fingerprint", fingerprint)

	// Look up user by fingerprint
	user, err := s.db.GetUserByFingerprint(context.Background(), fingerprint)
	if err != nil {
		log.Error("Database error during auth", "error", err)
		return false
	}

	if user == nil {
		// Unknown key - allow connection for registration flow
		// Store fingerprint in context for registration middleware
		ctx.SetValue("fingerprint", fingerprint)
		ctx.SetValue("needs_registration", true)
		log.Info("New key, will prompt for registration", "fingerprint", fingerprint)
		return true
	}

	if user.IsBanned {
		log.Warn("Banned user attempted connection", "user", user.Username, "fingerprint", fingerprint)
		return false
	}

	// Known user - store in context
	ctx.SetValue("user", user)
	ctx.SetValue("fingerprint", fingerprint)

	// Update last login
	if err := s.db.UpdateLastLogin(context.Background(), user.ID); err != nil {
		log.Error("Failed to update last login", "error", err)
	}

	log.Info("User authenticated", "user", user.Username)
	return true
}

// gameMiddleware returns the Bubble Tea middleware for the game.
func (s *Server) gameMiddleware() wish.Middleware {
	return bubbletea.Middleware(func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		return s.newGameSession(sess)
	})
}
