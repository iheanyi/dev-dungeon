// /dev/dungeon - A Unix-themed terminal roguelike
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/db"
	"github.com/iheanyi/devdungeon/internal/db/migrate"
	"github.com/iheanyi/devdungeon/internal/monitoring"
	"github.com/iheanyi/devdungeon/internal/server"
	"github.com/iheanyi/devdungeon/internal/ui"
	"github.com/iheanyi/devdungeon/internal/web"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Parse command line flags
	serverMode := flag.Bool("server", false, "Run as multiplayer server (SSH + HTTP)")
	sshOnly := flag.Bool("ssh-only", false, "Run SSH server only (no HTTP)")
	httpOnly := flag.Bool("http-only", false, "Run HTTP server only (no SSH)")

	// SSH server flags
	sshHost := flag.String("ssh-host", "0.0.0.0", "SSH server host")
	sshPort := flag.String("ssh-port", "2222", "SSH server port")
	hostKeyPath := flag.String("host-key", "", "Path to SSH host key (or set SSH_HOST_KEY_PATH env)")

	// HTTP server flags
	httpHost := flag.String("http-host", "0.0.0.0", "HTTP server host")
	httpPort := flag.String("http-port", "8080", "HTTP server port")
	staticDir := flag.String("static-dir", "web/build", "Path to static frontend assets")

	// Database
	databaseURL := flag.String("database-url", "", "PostgreSQL connection URL (or set DATABASE_URL env)")
	migrateForce := flag.Int("migrate-force", -1, "Force migration version (for fixing dirty state, use -1 to skip)")

	// Web base URL for magic links
	webBaseURL := flag.String("web-base-url", "", "Base URL for web portal (or set WEB_BASE_URL env)")

	flag.Parse()

	if *serverMode || *sshOnly || *httpOnly {
		// Resolve host key path: flag > env > default
		resolvedHostKeyPath := *hostKeyPath
		if resolvedHostKeyPath == "" {
			resolvedHostKeyPath = os.Getenv("SSH_HOST_KEY_PATH")
		}
		if resolvedHostKeyPath == "" {
			resolvedHostKeyPath = ".ssh/devdungeon_host_key"
		}

		// Resolve web base URL: flag > env > auto-detect
		resolvedWebBaseURL := *webBaseURL
		if resolvedWebBaseURL == "" {
			resolvedWebBaseURL = os.Getenv("WEB_BASE_URL")
		}
		// Auto-detect from HTTP config if not set
		if resolvedWebBaseURL == "" && !*sshOnly {
			if *httpHost == "0.0.0.0" || *httpHost == "" {
				resolvedWebBaseURL = fmt.Sprintf("http://localhost:%s", *httpPort)
			} else {
				resolvedWebBaseURL = fmt.Sprintf("http://%s:%s", *httpHost, *httpPort)
			}
		}

		cfg := serverConfig{
			sshEnabled:   !*httpOnly,
			httpEnabled:  !*sshOnly,
			sshHost:      *sshHost,
			sshPort:      *sshPort,
			hostKeyPath:  resolvedHostKeyPath,
			httpHost:     *httpHost,
			httpPort:     *httpPort,
			staticDir:    *staticDir,
			databaseURL:  *databaseURL,
			webBaseURL:   resolvedWebBaseURL,
			migrateForce: *migrateForce,
		}
		runServer(cfg)
	} else {
		runLocal()
	}
}

type serverConfig struct {
	sshEnabled   bool
	httpEnabled  bool
	sshHost      string
	sshPort      string
	hostKeyPath  string
	httpHost     string
	httpPort     string
	staticDir    string
	databaseURL  string
	webBaseURL   string
	migrateForce int
}

// runLocal runs the game in local single-player mode.
func runLocal() {
	// Load configuration
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create the UI model
	model := ui.New(cfg)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// runServer runs the game as a multiplayer server with SSH and/or HTTP.
func runServer(cfg serverConfig) {
	// Initialize Sentry error tracking
	sentryDSN := os.Getenv("SENTRY_DSN")
	sentryEnv := os.Getenv("SENTRY_ENVIRONMENT")
	if sentryEnv == "" {
		sentryEnv = "development"
	}

	sentryCleanup, err := monitoring.Init(monitoring.Config{
		DSN:         sentryDSN,
		Environment: sentryEnv,
		Debug:       sentryEnv == "development",
	})
	if err != nil {
		log.Error("Failed to initialize Sentry", "error", err)
	}
	defer sentryCleanup()

	// Get database URL from env if not provided
	databaseURL := cfg.databaseURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "Error: DATABASE_URL is required for server mode")
		fmt.Fprintln(os.Stderr, "Set via --database-url flag or DATABASE_URL environment variable")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Determine if SSL should be required (production environments)
	requireSSL := sentryEnv == "production" || os.Getenv("REQUIRE_DB_SSL") == "true"

	// Connect to database (shared between SSH and HTTP)
	dbClient, err := db.NewClientWithOptions(ctx, databaseURL, requireSSL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer dbClient.Close()

	// Force migration version if specified (for fixing dirty state)
	if cfg.migrateForce >= 0 {
		log.Warn("Forcing migration version", "version", cfg.migrateForce)
		if err := migrate.Force(dbClient.Pool(), cfg.migrateForce); err != nil {
			fmt.Fprintf(os.Stderr, "Error forcing migration version: %v\n", err)
			os.Exit(1)
		}
	}

	// Run database migrations
	log.Info("Running database migrations...")
	if err := migrate.Run(dbClient.Pool()); err != nil {
		fmt.Fprintf(os.Stderr, "Error running migrations: %v\n", err)
		os.Exit(1)
	}

	// Use errgroup for coordinated goroutine management
	g, gctx := errgroup.WithContext(ctx)

	// Periodically clean up expired DB rows to keep tables bounded.
	g.Go(func() error {
		runCleanup := func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(gctx, 10*time.Second)
			defer cleanupCancel()

			dropsDeleted, dropsErr := dbClient.CleanupExpiredDrops(cleanupCtx)
			tokensDeleted, tokensErr := dbClient.CleanupExpiredTokens(cleanupCtx)
			sessionsDeleted, sessionsErr := dbClient.CleanupExpiredSessions(cleanupCtx)

			if dropsErr != nil {
				log.Warn("Failed to clean up expired world drops", "error", dropsErr)
			}
			if tokensErr != nil {
				log.Warn("Failed to clean up expired auth tokens", "error", tokensErr)
			}
			if sessionsErr != nil {
				log.Warn("Failed to clean up expired web sessions", "error", sessionsErr)
			}

			totalDeleted := dropsDeleted + tokensDeleted + sessionsDeleted
			if totalDeleted > 0 {
				log.Info("Expired data cleanup complete",
					"worldDropsDeleted", dropsDeleted,
					"authTokensDeleted", tokensDeleted,
					"webSessionsDeleted", sessionsDeleted)
			}
		}

		// Run one cleanup pass at startup, then on an interval.
		runCleanup()
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-gctx.Done():
				return nil
			case <-ticker.C:
				runCleanup()
			}
		}
	})

	// Start SSH server if enabled
	var sshSrv *server.Server
	if cfg.sshEnabled {
		sshCfg := server.Config{
			Host:        cfg.sshHost,
			Port:        cfg.sshPort,
			HostKeyPath: cfg.hostKeyPath,
			DatabaseURL: databaseURL,
			WebBaseURL:  cfg.webBaseURL,
		}

		sshSrv, err = server.New(sshCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating SSH server: %v\n", err)
			os.Exit(1)
		}

		g.Go(func() error {
			log.Info("SSH server starting", "host", cfg.sshHost, "port", cfg.sshPort)
			return sshSrv.Run()
		})
	}

	// Start HTTP server if enabled
	var httpSrv *web.Server
	if cfg.httpEnabled {
		httpCfg := web.Config{
			Host:        cfg.httpHost,
			Port:        cfg.httpPort,
			DatabaseURL: databaseURL,
		}

		// Use embedded static files if available, otherwise fall back to disk
		if embeddedFS := getEmbeddedStaticFS(); embeddedFS != nil {
			httpCfg.StaticFS = embeddedFS
			log.Info("Using embedded static files")
		} else if cfg.staticDir != "" {
			httpCfg.StaticDir = cfg.staticDir
			log.Info("Using disk static files", "dir", cfg.staticDir)
		}

		httpSrv, err = web.New(httpCfg, dbClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating HTTP server: %v\n", err)
			os.Exit(1)
		}

		g.Go(func() error {
			log.Info("HTTP server starting", "host", cfg.httpHost, "port", cfg.httpPort)
			return httpSrv.Run()
		})
	}

	// Print startup info
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║           /dev/dungeon server starting                ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	if cfg.sshEnabled {
		fmt.Printf("  SSH:  ssh -p %s <username>@<host>\n", cfg.sshPort)
	}
	if cfg.httpEnabled {
		fmt.Printf("  HTTP: http://%s:%s\n", cfg.httpHost, cfg.httpPort)
	}
	fmt.Println()

	// Handle shutdown signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or error
	select {
	case <-done:
		log.Info("Shutdown signal received")
	case <-gctx.Done():
		log.Info("Server error, shutting down")
	}

	// Graceful shutdown
	log.Info("Shutting down servers...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if httpSrv != nil {
		_ = httpSrv.Shutdown(shutdownCtx)
	}

	cancel() // Cancel context to stop SSH server

	// Wait for servers to finish
	if err := g.Wait(); err != nil {
		log.Error("Server error during shutdown", "error", err)
	}

	log.Info("Servers stopped")
}
