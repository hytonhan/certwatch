package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hytonhan/certwatch/migrations"
)

func RunMigrations(ctx context.Context, db *sql.DB, path string) error {
	data, err := migrations.MigrationFiles.ReadFile("001_init.sql")
	if err != nil {
		return fmt.Errorf("Read migration: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, string(data)); err != nil {
		return fmt.Errorf("Execute migration: %w", err)
	}

	return nil
}
