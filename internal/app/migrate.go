package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/migrations"
)

func RunMigrations(dbURL string, direction string) error {
	dir, err := os.MkdirTemp("", "wtbot-bots-migrations-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()

	entries, err := migrations.Embedded.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		data, err := migrations.Embedded.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, entry.Name()), data, 0o644); err != nil {
			return err
		}
	}

	m, err := migrate.New("file://"+dir, dbURL)
	if err != nil {
		return err
	}
	switch direction {
	case "up", "":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unknown migration direction: %s", direction)
	}
}

func Ready(ctx context.Context) context.Context { return ctx }
