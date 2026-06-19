package db

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// toPgx5URL converts postgres:// to pgx5:// for the migrate driver.
func toPgx5URL(url string) string {
	return strings.Replace(url, "postgres://", "pgx5://", 1)
}

// RunMigrations runs all up migrations from the given path.
func RunMigrations(ctx context.Context, dbURL, migrationsPath string, log *zap.Logger) error {
	if migrationsPath == "" {
		migrationsPath = os.Getenv("CNOW_MIGRATIONS_PATH")
	}
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}
	if !strings.HasPrefix(migrationsPath, "file://") {
		migrationsPath = "file://" + migrationsPath
	}

	m, err := migrate.New(migrationsPath, toPgx5URL(dbURL))
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Info("migrations applied", zap.Uint("version", version), zap.Bool("dirty", dirty))
	return nil
}

// RollbackN rolls back N migration steps.
func RollbackN(dbURL, migrationsPath string, steps int) error {
	if !strings.HasPrefix(migrationsPath, "file://") {
		migrationsPath = "file://" + migrationsPath
	}
	m, err := migrate.New(migrationsPath, toPgx5URL(dbURL))
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()
	if steps <= 0 {
		steps = 1
	}
	return m.Steps(-steps)
}

// MigrationStatus returns the current migration version and dirty flag.
func MigrationStatus(dbURL, migrationsPath string) (uint, bool, error) {
	if !strings.HasPrefix(migrationsPath, "file://") {
		migrationsPath = "file://" + migrationsPath
	}
	m, err := migrate.New(migrationsPath, toPgx5URL(dbURL))
	if err != nil {
		return 0, false, fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()
	return m.Version()
}
