// Package web provides the HTTP server for /dev/dungeon web portal.
package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iheanyi/devdungeon/internal/db"
)

// Config holds HTTP server configuration.
type Config struct {
	Host        string
	Port        string
	StaticDir   string // Path to built frontend assets
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

	s.srv = &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/leaderboard", s.handleLeaderboard)
	s.mux.HandleFunc("GET /api/leaderboard/{runType}", s.handleLeaderboardByType)
	s.mux.HandleFunc("GET /api/players/{nanoid}", s.handlePlayerProfile)
	s.mux.HandleFunc("GET /api/daily", s.handleDailySeed)
	s.mux.HandleFunc("POST /api/register", s.handleRegister)

	// Static files (SvelteKit build output)
	if s.config.StaticDir != "" {
		s.mux.Handle("/", http.FileServer(http.Dir(s.config.StaticDir)))
	}
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
	json.NewEncoder(w).Encode(apiResponse{Success: true, Data: data})
}

func (s *Server) jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{Success: false, Error: message})
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
	limit := 100

	entries, err := s.db.GetLeaderboard(ctx, runType, limit)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to fetch leaderboard")
		return
	}

	s.jsonResponse(w, http.StatusOK, entries)
}

func (s *Server) handlePlayerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nanoid := r.PathValue("nanoid")

	user, err := s.db.GetUserByNanoID(ctx, nanoid)
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
		"nanoid":         user.NanoID,
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

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Username  string `json:"username"`
		PublicKey string `json:"public_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 20 {
		s.jsonError(w, http.StatusBadRequest, "Username must be 3-20 characters")
		return
	}

	if req.PublicKey == "" {
		s.jsonError(w, http.StatusBadRequest, "Public key is required")
		return
	}

	// Parse and get fingerprint from public key
	fingerprint, err := parsePublicKeyFingerprint(req.PublicKey)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid public key format")
		return
	}

	// Check if fingerprint already registered
	existing, err := s.db.GetUserByFingerprint(ctx, fingerprint)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existing != nil {
		s.jsonError(w, http.StatusConflict, "This SSH key is already registered")
		return
	}

	// Create user
	user, err := s.db.CreateUser(ctx, req.Username, fingerprint)
	if err != nil {
		s.jsonError(w, http.StatusConflict, "Username already taken")
		return
	}

	s.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"username": user.Username,
		"nanoid":   user.NanoID,
		"message":  "Account created! Connect with: ssh -p 2222 " + user.Username + "@devdungeon.io",
	})
}

// parsePublicKeyFingerprint extracts SHA256 fingerprint from an SSH public key.
func parsePublicKeyFingerprint(pubKey string) (string, error) {
	// Parse the public key
	key, _, _, _, err := parseAuthorizedKey([]byte(pubKey))
	if err != nil {
		return "", err
	}

	return fingerprintSHA256(key), nil
}
