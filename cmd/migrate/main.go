// Command migrate applies or rolls back database migrations using the
// golang-migrate library (no external CLI required).
//
// Usage:
//
//	go run ./cmd/migrate up        # apply all up migrations
//	go run ./cmd/migrate down 1    # roll back one migration
//	go run ./cmd/migrate version   # print current schema version
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/son-ngo/edu-app/config"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: migrate <up|down|version> [n]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	sourceURL, err := fileSourceURL(cfg.Postgres.MigrationDir)
	if err != nil {
		return err
	}

	m, err := migrate.New(sourceURL, cfg.Postgres.URL)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	switch args[0] {
	case "up":
		err = m.Up()
	case "down":
		steps := 1
		if len(args) > 1 {
			if steps, err = strconv.Atoi(args[1]); err != nil {
				return fmt.Errorf("invalid step count: %w", err)
			}
		}
		err = m.Steps(-steps)
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			return verr
		}
		fmt.Printf("version=%d dirty=%t\n", v, dirty)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	fmt.Println("migrate: ok")
	return nil
}

// fileSourceURL converts a (possibly relative, possibly Windows) migration
// directory path into a file:// URL golang-migrate accepts.
func fileSourceURL(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return "file://" + filepath.ToSlash(abs), nil
}
