package app

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/hytonhan/certwatch/internal/audit"
	"github.com/hytonhan/certwatch/internal/config"
	"github.com/hytonhan/certwatch/internal/db"
	"github.com/hytonhan/certwatch/internal/http/handler"
	"github.com/hytonhan/certwatch/internal/monitor"
	"github.com/hytonhan/certwatch/internal/repository"
	"github.com/hytonhan/certwatch/internal/service"
)

type App struct {
	Config config.Config
	DB     *sql.DB
	Server *http.Server
}

func New(cfg config.Config) (*App, error) {

	ctx := context.Background()
	logger := audit.NewLogger()

	sqlDB, err := db.NewSQLite(ctx, cfg.DBPath)
	if err != nil {
		logger.Error("Database initialization failed", "error", err)
		log.Fatal(err)
	}
	//defer sqlDB.Close()

	if err := db.RunMigrations(ctx, sqlDB, "migrations/001_init.sql"); err != nil {
		logger.Error("Migration failed", "error", err)
		log.Fatal(err)
	}

	logger.Info("Database initialized")

	repo := repository.NewCertificateRepository(sqlDB)
	certSrv := service.New(repo)

	handler := handler.NewCertificateHandler(certSrv, logger)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler.Router(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	monitor := monitor.NewMonitor(certSrv, cfg.ExpiryCheckInterval, cfg.ExpiryWindow, logger)
	go monitor.Start(ctx)

	return &App{Config: cfg, DB: sqlDB, Server: srv}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.DB.Close()
	go func() {
		// logger.Info("Server starting", "addr", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// logger.Error("Server failed", "error", err)
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	return a.Server.Shutdown(shutdownCtx)
}
