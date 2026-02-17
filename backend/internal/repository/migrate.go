package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL, migrationsPath string) error {
	// golang-migrate's pgx/v5 driver registers as "pgx5", not "postgres"
	migrateURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		migrateURL,
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
