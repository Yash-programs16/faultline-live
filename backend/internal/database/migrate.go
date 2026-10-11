package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FindMigrationsDir locates the SQL migrations directory. The server can run
// from the backend directory (local `go run .`, Render start command) or
// from the repository root, so several candidates are probed in order.
func FindMigrationsDir() (string, error) {
	candidates := []string{"migrations", "backend/migrations", "faultline-live-demo/backend/migrations"}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c, nil
		}
	}
	// Fall back to the executable's directory (single-binary deploys).
	if exe, err := os.Executable(); err == nil {
		for _, c := range []string{
			filepath.Join(filepath.Dir(exe), "migrations"),
			filepath.Join(filepath.Dir(exe), "backend", "migrations"),
		} {
			if st, err := os.Stat(c); err == nil && st.IsDir() {
				return c, nil
			}
		}
	}
	return "", fmt.Errorf("migrations directory not found (tried migrations, backend/migrations)")
}

// RunMigrations applies every pending *.sql migration. It is idempotent via
// the schema_migrations table and safe to run on every boot when
// AUTO_MIGRATE=true (the Render start path).
func RunMigrations(ctx context.Context, dbURL string) error {
	dir, err := FindMigrationsDir()
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("connect database for migrations: %w", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("query migrations: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err == nil {
			applied[v] = true
		}
	}
	rows.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var pending []string
	for _, f := range entries {
		if strings.HasSuffix(f.Name(), ".sql") {
			if v := strings.TrimSuffix(f.Name(), ".sql"); !applied[v] {
				pending = append(pending, v)
			}
		}
	}
	sort.Strings(pending)
	if len(pending) == 0 {
		return nil
	}

	for _, version := range pending {
		content, err := os.ReadFile(filepath.Join(dir, version+".sql"))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}
	return nil
}
