// Package migrate provides database migration support using golang-migrate.
package migrate

import (
	"embed"
	"errors"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var Migrations embed.FS

// Run executes all pending migrations against the database.
// It's safe to call on every startup - already applied migrations are skipped.
func Run(pool *pgxpool.Pool) error {
	// Get a *sql.DB from pgxpool for golang-migrate
	db := stdlib.OpenDBFromPool(pool)

	// Create the postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create the source driver from embedded files
	source, err := iofs.New(Migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create the migrator
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Debug("Database schema is up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		log.Warn("Database migration state is dirty", "version", version)
	} else {
		log.Info("Database migrations applied successfully", "version", version)
	}

	return nil
}

// Version returns the current migration version.
func Version(pool *pgxpool.Pool) (uint, bool, error) {
	db := stdlib.OpenDBFromPool(pool)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	source, err := iofs.New(Migrations, "migrations")
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m.Version()
}
