// cmd/migrate provides a CLI for database migrations.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	dbmigrate "github.com/iheanyi/devdungeon/internal/db/migrate"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: migrate [options] <command>\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  up        Apply all pending migrations\n")
		fmt.Fprintf(os.Stderr, "  down      Roll back the last migration\n")
		fmt.Fprintf(os.Stderr, "  drop      Drop all tables (DANGEROUS)\n")
		fmt.Fprintf(os.Stderr, "  version   Show current migration version\n")
		fmt.Fprintf(os.Stderr, "  force N   Force set version to N (for fixing dirty state)\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	databaseURL := flag.String("database-url", "", "PostgreSQL connection URL (or set DATABASE_URL env)")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Get database URL
	dbURL := *databaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "Error: DATABASE_URL is required")
		fmt.Fprintln(os.Stderr, "Set via --database-url flag or DATABASE_URL environment variable")
		os.Exit(1)
	}

	command := flag.Arg(0)

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create migrator
	m, err := createMigrator(pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating migrator: %v\n", err)
		os.Exit(1)
	}

	// Execute command
	switch command {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("Database is already up to date")
				return
			}
			fmt.Fprintf(os.Stderr, "Error running migrations: %v\n", err)
			os.Exit(1)
		}
		version, _, _ := m.Version()
		fmt.Printf("Migrations applied successfully (version %d)\n", version)

	case "down":
		if err := m.Steps(-1); err != nil {
			fmt.Fprintf(os.Stderr, "Error rolling back migration: %v\n", err)
			os.Exit(1)
		}
		version, _, _ := m.Version()
		fmt.Printf("Rolled back to version %d\n", version)

	case "drop":
		fmt.Print("WARNING: This will drop all tables. Type 'yes' to confirm: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("Aborted")
			return
		}
		if err := m.Drop(); err != nil {
			fmt.Fprintf(os.Stderr, "Error dropping tables: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All tables dropped")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("No migrations applied yet")
				return
			}
			fmt.Fprintf(os.Stderr, "Error getting version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current version: %d\n", version)
		if dirty {
			fmt.Println("WARNING: Database is in dirty state")
		}

	case "force":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "Error: force requires a version number")
			os.Exit(1)
		}
		var version int
		if _, err := fmt.Sscanf(flag.Arg(1), "%d", &version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid version number: %s\n", flag.Arg(1))
			os.Exit(1)
		}
		if err := m.Force(version); err != nil {
			fmt.Fprintf(os.Stderr, "Error forcing version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Forced version to %d\n", version)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func createMigrator(pool *pgxpool.Pool) (*migrate.Migrate, error) {
	db := stdlib.OpenDBFromPool(pool)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	source, err := iofs.New(dbmigrate.Migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	return migrate.NewWithInstance("iofs", source, "postgres", driver)
}
