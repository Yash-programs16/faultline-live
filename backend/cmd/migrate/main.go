package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL not set")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create migrations table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		slog.Error("create migrations table", "error", err)
		os.Exit(1)
	}

	// Get applied migrations
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		slog.Error("query migrations", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		rows.Scan(&v)
		applied[v] = true
	}

	// Find migration files
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		slog.Error("read migrations dir", "error", err)
		os.Exit(1)
	}

	var migrations []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".sql") {
			version := strings.TrimSuffix(f.Name(), ".sql")
			if !applied[version] {
				migrations = append(migrations, version)
			}
		}
	}
	sort.Strings(migrations)

	if len(migrations) == 0 {
		slog.Info("no pending migrations")
		return
	}

	for _, version := range migrations {
		filePath := filepath.Join(migrationsDir, version+".sql")
		content, err := os.ReadFile(filePath)
		if err != nil {
			slog.Error("read migration", "version", version, "error", err)
			os.Exit(1)
		}

		slog.Info("applying migration", "version", version)
		tx, err := pool.Begin(ctx)
		if err != nil {
			slog.Error("begin transaction", "error", err)
			os.Exit(1)
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			slog.Error("execute migration", "version", version, "error", err)
			os.Exit(1)
		}

		_, err = tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version)
		if err != nil {
			tx.Rollback(ctx)
			slog.Error("record migration", "version", version, "error", err)
			os.Exit(1)
		}

		if err := tx.Commit(ctx); err != nil {
			slog.Error("commit migration", "version", version, "error", err)
			os.Exit(1)
		}

		slog.Info("migration applied", "version", version)
	}

	slog.Info("all migrations applied", "count", len(migrations))
}
