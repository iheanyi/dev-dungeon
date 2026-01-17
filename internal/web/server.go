// Package web provides the HTTP server for /dev/dungeon web portal.
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/iheanyi/devdungeon/internal/db"
	"github.com/iheanyi/devdungeon/internal/monitoring"
)

// Session cookie configuration
const (
	sessionCookieName   = "devdungeon_session"
	usernameCookieName  = "devdungeon_user" // Non-HttpOnly, readable by JS
	sessionCookieMaxAge = 7 * 24 * 60 * 60  // 7 days in seconds
)

// Valid run types for leaderboards
var validRunTypes = map[string]bool{
	"standard": true,
	"daily":    true,
	"seeded":   true,
}

// Config holds HTTP server configuration.
type Config struct {
	Host        string
	Port        string
	StaticDir   string // Path to built frontend assets (for development)
	StaticFS    fs.FS  // Embedded static files (for production)
	DatabaseURL string
}

// Server is the HTTP web server.
type Server struct {
	config Config
	db     *db.Client
	mux    *http.ServeMux
	srv    *http.Server
}

// New creates a new HTTP server.
func New(cfg Config, dbClient *db.Client) (*Server, error) {
	s := &Server{
		config: cfg,
		db:     dbClient,
		mux:    http.NewServeMux(),
	}

	s.setupRoutes()

	// Wrap mux with middleware chain: recovery -> security headers
	handler := monitoring.HTTPRecoveryMiddleware(s.securityHeaders(s.mux))

	s.srv = &http.Server{
		Addr:           cfg.Host + ":" + cfg.Port,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 16, // 64KB max headers
	}

	return s, nil
}

// securityHeaders adds security headers to all responses.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Only add HSTS in production (when not localhost)
		if !strings.Contains(r.Host, "localhost") && !strings.Contains(r.Host, "127.0.0.1") {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// Public API routes
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/leaderboard", s.handleLeaderboard)
	s.mux.HandleFunc("GET /api/leaderboard/{runType}", s.handleLeaderboardByType)
	s.mux.HandleFunc("GET /api/leaderboard/daily/{date}", s.handleDailyLeaderboard)
	s.mux.HandleFunc("GET /api/players/{username}", s.handlePlayerProfile)
	s.mux.HandleFunc("GET /api/daily", s.handleDailySeed)

	// Auth routes
	s.mux.HandleFunc("GET /auth/verify", s.handleAuthVerify) // Magic link verification
	s.mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	s.mux.HandleFunc("GET /api/auth/me", s.handleAuthMe) // Get current user

	// Static files (SvelteKit build output) with SPA fallback
	// Prefer embedded FS, fall back to disk if specified
	if s.config.StaticFS != nil {
		s.mux.Handle("/", s.spaHandlerFS(s.config.StaticFS))
	} else if s.config.StaticDir != "" {
		s.mux.Handle("/", s.spaHandlerDir(s.config.StaticDir))
	}
}

// spaHandlerFS serves static files from an fs.FS with SPA fallback.
func (s *Server) spaHandlerFS(staticFS fs.FS) http.Handler {
	fileServer := http.FileServerFS(staticFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "."
		}

		// Try to open the requested file
		f, err := staticFS.Open(path)
		if err == nil {
			defer f.Close()
			stat, err := f.Stat()
			if err == nil {
				if stat.IsDir() {
					// Check for index.html in directory
					indexPath := filepath.Join(path, "index.html")
					if _, err := staticFS.Open(indexPath); err == nil {
						fileServer.ServeHTTP(w, r)
						return
					}
				} else {
					fileServer.ServeHTTP(w, r)
					return
				}
			}
		}

		// Serve index.html for SPA routing
		indexContent, err := fs.ReadFile(staticFS, "index.html")
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexContent)
			return
		}

		http.NotFound(w, r)
	})
}

// spaHandlerDir serves static files from a directory with SPA fallback.
func (s *Server) spaHandlerDir(staticDir string) http.Handler {
	dirFS := http.Dir(staticDir)
	fileServer := http.FileServer(dirFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to open the requested file
		f, err := dirFS.Open(path)
		if err == nil {
			defer f.Close()
			stat, err := f.Stat()
			if err == nil {
				if stat.IsDir() {
					indexPath := filepath.Join(path, "index.html")
					if _, err := dirFS.Open(indexPath); err == nil {
						fileServer.ServeHTTP(w, r)
						return
					}
				} else {
					fileServer.ServeHTTP(w, r)
					return
				}
			}
		}

		// Serve index.html for SPA routing
		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}

		http.NotFound(w, r)
	})
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// API response helpers

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: true, Data: data})
}

func (s *Server) jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{Success: false, Error: message})
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "/dev/dungeon",
	})
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit := 100

	entries, err := s.db.GetLeaderboard(ctx, "", limit)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch leaderboard")
		return
	}

	s.jsonResponse(w, http.StatusOK, entries)
}

func (s *Server) handleLeaderboardByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	runType := r.PathValue("runType")

	// Validate runType parameter
	if !validRunTypes[runType] {
		s.jsonError(w, http.StatusBadRequest, "Invalid run type. Must be: standard, daily, or seeded")
		return
	}

	limit := 100

	entries, err := s.db.GetLeaderboard(ctx, runType, limit)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch leaderboard")
		return
	}

	s.jsonResponse(w, http.StatusOK, entries)
}

func (s *Server) handleDailyLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dateStr := r.PathValue("date")

	// Parse date in YYYY-MM-DD format
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Validate date range (last 7 days only)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	minDate := today.AddDate(0, 0, -6)
	if date.After(today) {
		s.jsonError(w, http.StatusBadRequest, "Cannot fetch future dates")
		return
	}
	if date.Before(minDate) {
		s.jsonError(w, http.StatusBadRequest, "Can only fetch last 7 days")
		return
	}

	limit := 100
	entries, _, err := s.db.GetDailyLeaderboard(ctx, date, limit, nil)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch daily leaderboard")
		return
	}

	// Get the seed for this date
	seed, err := s.db.GetDailySeed(ctx, date)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch daily seed")
		return
	}

	response := map[string]interface{}{
		"date":    date.Format("2006-01-02"),
		"seed":    nil,
		"entries": entries,
	}
	if seed != nil {
		response["seed"] = seed.Seed
	}

	s.jsonResponse(w, http.StatusOK, response)
}

func (s *Server) handlePlayerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.PathValue("username")

	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch player")
		return
	}
	if user == nil {
		s.jsonError(w, http.StatusNotFound, "Player not found")
		return
	}

	// Get meta progress
	meta, err := s.db.GetMetaProgress(ctx, user.ID)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch progress")
		return
	}

	profile := map[string]interface{}{
		"username":       user.Username,
		"public_id":      user.NanoID,
		"created_at":     user.CreatedAt,
		"runs_completed": 0,
		"deepest_floor":  0,
		"total_deaths":   0,
	}

	if meta != nil {
		profile["runs_completed"] = meta.RunsCompleted
		profile["deepest_floor"] = meta.DeepestFloor
		profile["total_deaths"] = meta.TotalDeaths
		profile["unlocked_classes"] = meta.UnlockedClasses
	}

	s.jsonResponse(w, http.StatusOK, profile)
}

func (s *Server) handleDailySeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seed, err := s.db.GetOrCreateDailySeed(ctx)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to get daily seed")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"date": time.Now().UTC().Format("2006-01-02"),
		"seed": seed,
	})
}

// handleAuthVerify handles magic link verification.
// This endpoint is called when a user clicks the magic link from the terminal.
func (s *Server) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from query string
	token := r.URL.Query().Get("token")
	if token == "" {
		s.renderAuthError(w, "Missing authentication token")
		return
	}

	// Validate token format (64 hex chars)
	if len(token) != 64 {
		s.renderAuthError(w, "Invalid token format")
		return
	}

	// Verify the magic link token
	user, err := s.db.VerifyAuthToken(ctx, token)
	if err != nil {
		log.Error("Failed to verify auth token", "error", err)
		s.renderAuthError(w, "Authentication failed. Please try again.")
		return
	}
	if user == nil {
		s.renderAuthError(w, "Invalid or expired link. Please generate a new one with Ctrl+L in the game.")
		return
	}

	// Create a web session
	sessionToken, err := s.db.CreateWebSession(ctx, user.ID)
	if err != nil {
		log.Error("Failed to create web session", "error", err, "user", user.Username)
		s.renderAuthError(w, "Failed to create session. Please try again.")
		return
	}

	// Set secure session cookie and username cookie
	s.setSessionCookie(w, r, sessionToken, user.Username)

	log.Info("User authenticated via magic link", "user", user.Username)

	// Redirect to profile or home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleLogout handles user logout.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		s.jsonResponse(w, http.StatusOK, map[string]string{"message": "Already logged out"})
		return
	}

	// Delete session from database
	if err := s.db.DeleteWebSession(ctx, cookie.Value); err != nil {
		log.Error("Failed to delete session", "error", err)
	}

	// Clear the cookie
	s.clearSessionCookie(w, r)

	s.jsonResponse(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// handleAuthMe returns the current authenticated user's info.
func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		s.jsonError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Validate session
	user, err := s.db.GetWebSession(ctx, cookie.Value)
	if err != nil {
		log.Error("Failed to validate session", "error", err)
		s.jsonError(w, http.StatusInternalServerError, "Session validation failed")
		return
	}
	if user == nil {
		s.clearSessionCookie(w, r)
		s.jsonError(w, http.StatusUnauthorized, "Session expired")
		return
	}

	// Get meta progress for additional info
	meta, _ := s.db.GetMetaProgress(ctx, user.ID)

	response := map[string]interface{}{
		"username":   user.Username,
		"public_id":  user.NanoID,
		"created_at": user.CreatedAt,
	}

	if meta != nil {
		response["runs_completed"] = meta.RunsCompleted
		response["deepest_floor"] = meta.DeepestFloor
		response["total_deaths"] = meta.TotalDeaths
		response["unlocked_classes"] = meta.UnlockedClasses
		response["total_exit_codes"] = meta.TotalExitCodes
	}

	s.jsonResponse(w, http.StatusOK, response)
}

// setSessionCookie sets a secure session cookie and a username cookie.
// The session cookie is HttpOnly (not accessible to JS) for security.
// The username cookie is readable by JS for UI display without an API call.
func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, token string, username string) {
	secure := !strings.Contains(r.Host, "localhost") && !strings.Contains(r.Host, "127.0.0.1")

	// Session token - HttpOnly for security
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionCookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})

	// Username - readable by JS for UI display (not sensitive)
	http.SetCookie(w, &http.Cookie{
		Name:     usernameCookieName,
		Value:    username,
		Path:     "/",
		MaxAge:   sessionCookieMaxAge,
		HttpOnly: false, // JS can read this
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

// clearSessionCookie removes both the session and username cookies.
func (s *Server) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	secure := !strings.Contains(r.Host, "localhost") && !strings.Contains(r.Host, "127.0.0.1")

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete immediately
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     usernameCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

// renderAuthError renders an HTML error page for auth failures.
func (s *Server) renderAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>Authentication Failed - /dev/dungeon</title>
	<style>
		body {
			font-family: monospace;
			background: #1a1a2e;
			color: #eee;
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
			margin: 0;
		}
		.error-box {
			background: #16213e;
			border: 2px solid #e94560;
			padding: 2rem;
			border-radius: 8px;
			max-width: 500px;
			text-align: center;
		}
		h1 { color: #e94560; }
		p { line-height: 1.6; }
		code { background: #0f3460; padding: 0.2rem 0.4rem; border-radius: 4px; }
	</style>
</head>
<body>
	<div class="error-box">
		<h1>Authentication Failed</h1>
		<p>%s</p>
		<p>To authenticate, press <code>Ctrl+L</code> in the game to generate a new link.</p>
	</div>
</body>
</html>`, message)
}
