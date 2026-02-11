package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

func RunMigrations(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
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
