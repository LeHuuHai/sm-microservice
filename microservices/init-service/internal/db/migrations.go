package db

import (
	"embed"
	"fmt"
	"log/slog"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(config *pkgconfig.PostgresConfig, fs embed.FS, path string) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	sourceDriver, err := iofs.New(fs, path)
	if err != nil {
		return fmt.Errorf("failed to create source driver: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %v", err)
	}

	slog.Info("Database migrations completed successfully", "db", config.Database)
	return nil
}
