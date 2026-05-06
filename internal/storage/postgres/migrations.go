package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations 执行指定目录下的 SQL migration 文件。
// migrationsDir: migration 文件所在目录路径（如 "./migrations"）
// 按文件名字典序执行（001_init.sql, 002_xxx.sql...）。
func RunMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		name := filepath.Base(file)
		if err := runMigrationFile(ctx, db, file, name); err != nil {
			return err
		}
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func runMigrationFile(ctx context.Context, db *sql.DB, file, name string) error {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, name,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check migration %s: %w", name, err)
	}
	if count > 0 {
		return nil
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for %s: %w", name, err)
	}

	if _, err := tx.ExecContext(ctx, strings.TrimSpace(string(content))); err != nil {
		tx.Rollback()
		return fmt.Errorf("exec migration %s: %w", name, err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1)`, name,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	return tx.Commit()
}
