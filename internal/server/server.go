// Package server provides the SSH server for /dev/dungeon multiplayer.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
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
	"github.com/iheanyi/devdungeon/internal/monitoring"
)

// rateLimiter tracks authentication attempts per IP.
type rateLimiter struct {
	attempts map[string]*attemptInfo
	mu       sync.RWMutex
	stop     chan struct{}
}

type attemptInfo struct {
	count     int
	firstTime time.Time
	blocked   bool
}

// Rate limit settings
const (
	maxAuthAttempts = 5           // Max attempts per window
	rateLimitWindow = time.Minute // Window duration
	blockDuration   = 5 * time.Minute
)

// SSH banner shown during connection
const sshBanner = `
╔═══════════════════════════════════════════════════════╗
║                    /dev/dungeon                       ║
║           A Unix-themed terminal roguelike            ║
╚═══════════════════════════════════════════════════════╝

`

// SSH key help shown when auth fails due to unsupported key type
const sshKeyHelp = `
╔═══════════════════════════════════════════════════════╗
║              Unsupported SSH Key Type                 ║
╠═══════════════════════════════════════════════════════╣
║  Supported keys:                                      ║
║    • Ed25519 (recommended)                            ║
║    • ECDSA (P-256, P-384, P-521)                      ║
║    • RSA (2048+ bits)                                 ║
║    • Security Keys (YubiKey, etc.)                    ║
║                                                       ║
║  Generate a new key:                                  ║
║    ssh-keygen -t ed25519                              ║
╚═══════════════════════════════════════════════════════╝

`

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		attempts: make(map[string]*attemptInfo),
		stop:     make(chan struct{}),
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, info := range rl.attempts {
				// Remove entries older than block duration
				if now.Sub(info.firstTime) > blockDuration {
					delete(rl.attempts, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Close stops the cleanup goroutine.
func (rl *rateLimiter) Close() {
	close(rl.stop)
}

func (rl *rateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	info, exists := rl.attempts[ip]
	now := time.Now()

	if !exists {
		rl.attempts[ip] = &attemptInfo{count: 1, firstTime: now}
		return true
	}

	// Check if blocked
	if info.blocked {
		if now.Sub(info.firstTime) > blockDuration {
			// Unblock after duration
			info.blocked = false
			info.count = 1
			info.firstTime = now
			return true
		}
		return false
	}

	// Check if window expired
	if now.Sub(info.firstTime) > rateLimitWindow {
		info.count = 1
		info.firstTime = now
		return true
	}

	// Increment counter
	info.count++
	if info.count > maxAuthAttempts {
		info.blocked = true
		log.Warn("Rate limiting IP", "ip", ip, "attempts", info.count)
		return false
	}

	return true
}

// Config holds SSH server configuration.
type Config struct {
	Host        string
	Port        string
	HostKeyPath string
	DatabaseURL string
	WebBaseURL  string // Base URL for magic links (e.g., "https://dev-dungeon.com")
}

// Server is the SSH game server.
type Server struct {
	config      Config
	db          *db.Client
	sshSrv      *ssh.Server
	sessions    *SessionManager
	rateLimiter *rateLimiter
}

// New creates a new SSH game server.
func New(cfg Config) (*Server, error) {
	ctx := context.Background()

	// Connect to database
	dbClient, err := db.NewClient(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensure host key directory exists
	if err := ensureHostKeyDir(cfg.HostKeyPath); err != nil {
		dbClient.Close()
		return nil, fmt.Errorf("failed to prepare host key directory: %w", err)
	}

	// Log host key information
	logHostKeyInfo(cfg.HostKeyPath)

	s := &Server{
		config:      cfg,
		db:          dbClient,
		sessions:    NewSessionManager(),
		rateLimiter: newRateLimiter(),
	}

	// Create SSH server with Wish
	// Note: Wish will generate the host key if it doesn't exist
	sshSrv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(cfg.Host, cfg.Port)),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithBanner(sshBanner),
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

// ensureHostKeyDir ensures the directory for the host key exists.
func ensureHostKeyDir(hostKeyPath string) error {
	dir := filepath.Dir(hostKeyPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0700)
}

// logHostKeyInfo logs information about the SSH host key.
func logHostKeyInfo(hostKeyPath string) {
	if _, err := os.Stat(hostKeyPath); os.IsNotExist(err) {
		log.Info("SSH host key will be generated", "path", hostKeyPath)
	} else if err != nil {
		log.Warn("Could not check host key status", "path", hostKeyPath, "error", err)
	} else {
		log.Info("Using existing SSH host key", "path", hostKeyPath)
	}
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

	// Notify all connected users before disconnecting
	s.sessions.NotifyShutdown("New deploy detected, please reconnect!")

	// Give users a moment to see the message
	time.Sleep(500 * time.Millisecond)

	// Save all active sessions
	s.sessions.SaveAll(ctx, s.db)

	// Shutdown SSH server (cancels session contexts)
	if err := s.sshSrv.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Shutdown error", "error", err)
	}

	// Wait for all auto-save goroutines to complete before closing DB
	log.Info("Waiting for auto-save goroutines to complete...")
	s.sessions.WaitForAutoSaves()

	s.rateLimiter.Close()
	s.db.Close()
	log.Info("Server stopped")
	return nil
}

// publicKeyAuth handles SSH public key authentication.
func (s *Server) publicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	// Get client IP for rate limiting
	remoteAddr := ctx.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(remoteAddr)
	if ip == "" {
		ip = remoteAddr
	}

	// Check rate limit
	if !s.rateLimiter.isAllowed(ip) {
		log.Warn("Rate limited", "ip", ip)
		return false
	}

	fingerprint := gossh.FingerprintSHA256(key)
	username := ctx.User()

	log.Debug("Auth attempt", "user", username, "fingerprint", fingerprint, "ip", ip, "key_type", key.Type())

	// Validate key type - only allow secure algorithms
	// If unsupported, we still accept to show a helpful error message in middleware
	if !isKeyTypeSupported(key) {
		log.Warn("Unsupported key type", "type", key.Type(), "fingerprint", fingerprint)
		ctx.SetValue("unsupported_key_type", key.Type())
		return true
	}

	// Look up user by fingerprint with timeout
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := s.db.GetUserByFingerprint(dbCtx, fingerprint)
	if err != nil {
		log.Error("Database error during auth", "error", err)
		monitoring.CaptureException(err, map[string]string{
			"operation":   "auth_lookup",
			"fingerprint": fingerprint,
			"ip":          ip,
		})
		return false
	}

	if user == nil {
		// Unknown key - allow connection for registration flow
		// Store fingerprint in context for registration middleware
		ctx.SetValue("fingerprint", fingerprint)
		ctx.SetValue("needs_registration", true)
		log.Info("New key, will prompt for registration", "fingerprint", fingerprint, "ip", ip)
		return true
	}

	if user.IsBanned {
		log.Warn("Banned user attempted connection", "user", user.Username, "fingerprint", fingerprint, "ip", ip)
		return false
	}

	// Known user - store in context
	ctx.SetValue("user", user)
	ctx.SetValue("fingerprint", fingerprint)

	// Update last login with timeout
	updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer updateCancel()
	if err := s.db.UpdateLastLogin(updateCtx, user.ID); err != nil {
		log.Error("Failed to update last login", "error", err)
	}

	// Set Sentry user context for error tracking
	monitoring.SetUser(user.NanoID, user.Username, fingerprint)

	log.Info("User authenticated", "user", user.Username, "ip", ip)
	return true
}

// Supported SSH key types (secure algorithms only).
var supportedKeyTypes = map[string]bool{
	gossh.KeyAlgoED25519:    true, // Ed25519 - recommended
	gossh.KeyAlgoECDSA256:   true, // ECDSA P-256
	gossh.KeyAlgoECDSA384:   true, // ECDSA P-384
	gossh.KeyAlgoECDSA521:   true, // ECDSA P-521
	gossh.KeyAlgoRSA:        true, // RSA (>= 2048 bits enforced by library)
	gossh.KeyAlgoSKED25519:  true, // Security Key Ed25519
	gossh.KeyAlgoSKECDSA256: true, // Security Key ECDSA
}

// isKeyTypeSupported checks if the key type is one of the supported algorithms.
func isKeyTypeSupported(key ssh.PublicKey) bool {
	return supportedKeyTypes[key.Type()]
}

// gameMiddleware returns the Bubble Tea middleware for the game.
func (s *Server) gameMiddleware() wish.Middleware {
	// Wrap the bubbletea middleware to add a goodbye message
	btMiddleware := bubbletea.Middleware(func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		return s.newGameSession(sess)
	})

	return func(next ssh.Handler) ssh.Handler {
		// Get the bubbletea handler
		btHandler := btMiddleware(next)

		return func(sess ssh.Session) {
			// Run the game
			btHandler(sess)

			// After the game ends, show goodbye message
			goodbye := "\r\n\033[1;32m" + // Bold green
				"╔════════════════════════════════════════╗\r\n" +
				"║   Thanks for playing /dev/dungeon!     ║\r\n" +
				"║   Your progress has been saved.        ║\r\n" +
				"╚════════════════════════════════════════╝\r\n" +
				"\033[0m" // Reset
			_, _ = sess.Write([]byte(goodbye))
		}
	}
}
