package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// RunMigrations executes SQL migration files in lexical order.
func RunMigrations(ctx context.Context, db *gorm.DB, migrationsDir string) error {
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

func ensureMigrationsTable(ctx context.Context, db *gorm.DB) error {
	err := db.WithContext(ctx).Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`).Error
	if err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}
	return nil
}

func runMigrationFile(ctx context.Context, db *gorm.DB, file, name string) error {
	var count int64
	err := db.WithContext(ctx).
		Model(&SchemaMigrationModel{}).
		Where(&SchemaMigrationModel{Version: name}).
		Count(&count).Error
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

	sqlText := strings.TrimSpace(string(content))
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sqlText).Error; err != nil {
			return fmt.Errorf("exec migration %s: %w", name, err)
		}
		if err := tx.Create(&SchemaMigrationModel{Version: name}).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type SchemaMigrationModel struct {
	Version string `gorm:"column:version;primaryKey"`
}

func (SchemaMigrationModel) TableName() string {
	return "schema_migrations"
}
