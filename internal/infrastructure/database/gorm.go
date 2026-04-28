// Package database provides GORM connection and migration helpers.
package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM database connection using the provided config.
func Connect(cfg *config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	log.Info("database connected")
	return db, nil
}

// RunMigrations applies pending SQL migrations in sequential order.
// It maintains a schema_migrations table to track which files have run.
// Only the section above "---- ... (down) ----" is executed.
func RunMigrations(db *gorm.DB, migrationsDir string, log *zap.Logger) error {
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	// Ensure the tracking table exists.
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	// Read and sort migration files.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %s: %w", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Fetch already-applied versions.
	var applied []string
	if err := db.Raw("SELECT version FROM schema_migrations").Scan(&applied).Error; err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}
	appliedSet := make(map[string]bool, len(applied))
	for _, v := range applied {
		appliedSet[v] = true
	}

	// Apply pending migrations.
	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		if appliedSet[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("reading migration file %s: %w", file, err)
		}

		// Execute only the "up" portion (above the first down comment).
		upSQL := extractUpSQL(string(content))
		if strings.TrimSpace(upSQL) == "" {
			continue
		}

		if err := db.Exec(upSQL).Error; err != nil {
			return fmt.Errorf("executing migration %s: %w", file, err)
		}

		if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version).Error; err != nil {
			return fmt.Errorf("recording migration %s: %w", file, err)
		}

		log.Info("migration applied", zap.String("version", version))
	}

	log.Info("all migrations up to date")
	return nil
}

// extractUpSQL returns the up-migration SQL, trimming down-migration content.
//
// Two separator styles are supported:
//   - Old (tern):  "---- create above / drop below ----"  → stop here
//   - New:         "---- Description (down) ----"          → stop here
//
// Descriptive headers like "---- Create roles table ----" (no "down"/"drop below")
// are treated as SQL comments and passed through — PostgreSQL ignores them.
func extractUpSQL(content string) string {
	lines := strings.Split(content, "\n")
	var up []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "----") {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "down") || strings.Contains(lower, "drop below") {
				break
			}
		}
		up = append(up, line)
	}
	return strings.Join(up, "\n")
}
