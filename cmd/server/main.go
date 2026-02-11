package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hytonhan/certwatch/internal/audit"
	"github.com/hytonhan/certwatch/internal/db"
)

func main() {
	ctx := context.Background()
	logger := audit.NewLogger()

	dbPath := getEnv("DB_PATH", "./data/certwatch.db")

	sqlDB, err := db.NewSQLite(ctx, dbPath)
	if err != nil {
		logger.Error("Database initialization failed", "error", err)
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := db.RunMigrations(ctx, sqlDB, "migrations/001_init.sql"); err != nil {
		logger.Error("Migration failed", "error", err)
		log.Fatal(err)
	}

	logger.Info("Database initialized")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		logger.Info("Server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			log.Fatal(err)
		}
	}()

	waitForShutdown(logger, srv)
}

func waitForShutdown(logger *slog.Logger, srv *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
	}
}

func getEnv(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
