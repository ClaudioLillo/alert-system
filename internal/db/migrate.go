package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RunMigrations executes SQL migration files in lexicographic order and records each one
// so repeated runs remain idempotent.
func RunMigrations(ctx context.Context, database *sql.DB, migrationDir string) error {
	if _, err := database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("ensure migration tracking table: %w", err)
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		var exists bool
		if err := database.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, entry.Name()).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", entry.Name(), err)
		}
		if exists {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if _, err := database.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}

		if _, err := database.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, entry.Name()); err != nil {
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
