package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := setupDB(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to setup database", slog.Any("err", err))
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

// setupDB initializes the PostgreSQL connection pool using database/sql and pgx.
func setupDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// Reasonable pool settings for dev; tune for prod
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Hour)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
