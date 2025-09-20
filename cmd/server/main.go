package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	_ "github.com/jackc/pgx/v5/stdlib"

	"backend/internal"
	"backend/internal/config"
)

// @title           Payway API
// @version         0.1.0
// @description     API for managing orders and accounts.
// @schemes         http
// @basePath        /
// @contact.name    API Support
// @contact.email   support@example.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

func main() {
	// Initialize structured logging with source info
	slogHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})
	slog.SetDefault(slog.New(slogHandler))

	cfg := config.FromEnv()

	db, err := setupDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to setup database after retries", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()

	h, err := internal.NewHTTPServer(cfg, db)
	if err != nil {
		slog.Error("failed to start HTTP server", slog.Any("err", err))
		os.Exit(1)
	}

	addr := ":8080"
	slog.Info("listening", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, h); err != nil {
		slog.Error("http server error", slog.Any("err", err))
		os.Exit(1)
	}
}

// setupDB initializes the PostgreSQL pool and retries Ping using exponential backoff.
func setupDB(dsn string) (*sql.DB, error) {
	// Open does not establish connections immediately.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// Reasonable pool settings for dev; tune for prod
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)

	// Configure backoff: start ~2s, cap ~5s, total ~12s (~3 attempts)
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 2 * time.Second
	bo.MaxInterval = 5 * time.Second
	bo.Multiplier = 2
	bo.RandomizationFactor = 0.1
	bo.MaxElapsedTime = 12 * time.Second

	operation := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := db.PingContext(ctx)
		cancel()
		return err
	}

	notify := func(err error, d time.Duration) {
		slog.Warn("database ping failed, will retry", slog.Any("err", err), slog.Duration("next_in", d))
	}

	if err := backoff.RetryNotify(operation, bo, notify); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database not available after retries: %w", err)
	}

	slog.Info("database connection established")
	return db, nil
}
