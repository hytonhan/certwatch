package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const (
	defaultMaxOpenConns = 1
	defaultMaxIdleConns = 1
	connMaxLifetime     = 0
)

func NewSQLite(ctx context.Context, path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("Database path cannot be empty")
	}

	// Ensure directory exists (if file-based DB)
	if err := ensureDir(path); err != nil {
		return nil, fmt.Errorf("Ensure db dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("Open db: %w", err)
	}

	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Validate connection with timeout
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("Ping db: %w", err)
	}

	return db, nil
}

func ensureDir(path string) error {
	dir := "."
	if idx := len(path) - len("certwatch.db"); idx > 0 {
		dir = path[:idx]
	}

	if dir == "" || dir == "." {
		return nil
	}

	return os.MkdirAll(dir, 0700)
}
