package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DatabaseURL             string // e.g. postgres connection string for pgx: postgres://user:pass@host:5432/dbname?sslmode=disable
	FirebaseCredentialsFile string // path to Firebase service account JSON
	WebsiteURL              string
}

func FromEnv() Config {
	c := Config{
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		FirebaseCredentialsFile: os.Getenv("FIREBASE_CREDENTIALS_FILE"),
		WebsiteURL:              os.Getenv("WEBSITE_URL"),
	}

	if c.DatabaseURL == "" {
		slog.Error("missing required environment variable", slog.String("var", "DATABASE_URL"))
		os.Exit(1)
	}

	if c.FirebaseCredentialsFile == "" {
		slog.Info("FIREBASE_CREDENTIALS_FILE is empty – defaulting to applicationDefault()")
	}

	if c.WebsiteURL == "" {
		slog.Error("missing required environment variable", slog.String("var", "WEBSITE_URL"))
		os.Exit(1)
	}

	return c
}
