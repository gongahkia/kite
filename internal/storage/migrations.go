package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: []Migration{},
	}
}

// Register registers a migration
func (mm *MigrationManager) Register(m Migration) {
	mm.migrations = append(mm.migrations, m)
}

// RegisterAll registers multiple migrations
func (mm *MigrationManager) RegisterAll(migrations []Migration) {
	mm.migrations = append(mm.migrations, migrations...)
}

// createMigrationsTable creates the migrations tracking table
func (mm *MigrationManager) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := mm.db.ExecContext(ctx, query)
	return err
}

// getCurrentVersion gets the current migration version
func (mm *MigrationManager) getCurrentVersion(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`

	var version int
	err := mm.db.QueryRowContext(ctx, query).Scan(&version)
	return version, err
}

// recordMigration records a migration as applied
func (mm *MigrationManager) recordMigration(ctx context.Context, version int, description string) error {
	query := `INSERT INTO schema_migrations (version, description) VALUES (?, ?)`
	_, err := mm.db.ExecContext(ctx, query, version, description)
	return err
}

// deleteMigration removes a migration record (for rollback)
func (mm *MigrationManager) deleteMigration(ctx context.Context, version int) error {
	query := `DELETE FROM schema_migrations WHERE version = ?`
	_, err := mm.db.ExecContext(ctx, query, version)
	return err
}

// Migrate runs all pending migrations
func (mm *MigrationManager) Migrate(ctx context.Context) error {
	// Create migrations table
	if err := mm.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := mm.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Sort migrations by version
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version < mm.migrations[j].Version
	})

	// Run pending migrations
	for _, migration := range mm.migrations {
		if migration.Version <= currentVersion {
			continue
		}

		fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Description)

		// Begin transaction
		tx, err := mm.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Run migration
		if err := migration.Up(mm.db); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}

		// Record migration
		if err := mm.recordMigration(ctx, migration.Version, migration.Description); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		fmt.Printf("Migration %d applied successfully\n", migration.Version)
	}

	return nil
}

// Rollback rolls back the last N migrations
func (mm *MigrationManager) Rollback(ctx context.Context, steps int) error {
	if steps <= 0 {
		steps = 1
	}

	// Get current version
	currentVersion, err := mm.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Sort migrations by version (descending)
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version > mm.migrations[j].Version
	})

	// Rollback migrations
	rolledBack := 0
	for _, migration := range mm.migrations {
		if migration.Version > currentVersion {
			continue
		}

		if rolledBack >= steps {
			break
		}

		fmt.Printf("Rolling back migration %d: %s\n", migration.Version, migration.Description)

		// Begin transaction
		tx, err := mm.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Run rollback
		if migration.Down != nil {
			if err := migration.Down(mm.db); err != nil {
				tx.Rollback()
				return fmt.Errorf("rollback %d failed: %w", migration.Version, err)
			}
		}

		// Delete migration record
		if err := mm.deleteMigration(ctx, migration.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete migration record %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback %d: %w", migration.Version, err)
		}

		fmt.Printf("Migration %d rolled back successfully\n", migration.Version)
		rolledBack++
	}

	return nil
}

// Status returns the current migration status
func (mm *MigrationManager) Status(ctx context.Context) ([]MigrationStatus, error) {
	// Create migrations table if it doesn't exist
	if err := mm.createMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	query := `SELECT version, description, applied_at FROM schema_migrations ORDER BY version`
	rows, err := mm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]MigrationStatus)
	for rows.Next() {
		var status MigrationStatus
		if err := rows.Scan(&status.Version, &status.Description, &status.AppliedAt); err != nil {
			return nil, err
		}
		status.Applied = true
		applied[status.Version] = status
	}

	// Build complete status list
	var statuses []MigrationStatus
	for _, migration := range mm.migrations {
		if status, ok := applied[migration.Version]; ok {
			statuses = append(statuses, status)
		} else {
			statuses = append(statuses, MigrationStatus{
				Version:     migration.Version,
				Description: migration.Description,
				Applied:     false,
			})
		}
	}

	// Sort by version
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Version < statuses[j].Version
	})

	return statuses, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version     int       `json:"version"`
	Description string    `json:"description"`
	Applied     bool      `json:"applied"`
	AppliedAt   time.Time `json:"applied_at,omitempty"`
}

// DefaultMigrations returns the default migrations for Kite
func DefaultMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Initial schema",
			Up: func(db *sql.DB) error {
				// This is handled by initSchema in each adapter
				return nil
			},
			Down: func(db *sql.DB) error {
				_, err := db.Exec(`
					DROP TABLE IF EXISTS citations;
					DROP TABLE IF EXISTS judges;
					DROP TABLE IF EXISTS cases;
				`)
				return err
			},
		},
		{
			Version:     2,
			Description: "Add full-text search indexes",
			Up: func(db *sql.DB) error {
				// SQLite-specific FTS implementation
				// This is handled by initSchema in SQLite adapter
				return nil
			},
			Down: func(db *sql.DB) error {
				_, err := db.Exec(`
					DROP TRIGGER IF EXISTS cases_fts_update;
					DROP TRIGGER IF EXISTS cases_fts_delete;
					DROP TRIGGER IF EXISTS cases_fts_insert;
					DROP TABLE IF EXISTS cases_fts;
				`)
				return err
			},
		},
		{
			Version:     3,
			Description: "Add quality score fields",
			Up: func(db *sql.DB) error {
				// Add quality_score and completeness_score columns
				_, err := db.Exec(`
					ALTER TABLE cases ADD COLUMN quality_score REAL DEFAULT 0.0;
					ALTER TABLE cases ADD COLUMN completeness_score REAL DEFAULT 0.0;
				`)
				return err
			},
			Down: func(db *sql.DB) error {
				// SQLite doesn't support DROP COLUMN, so we'd need to recreate table
				// For simplicity, we'll just note this in the down migration
				return fmt.Errorf("rollback not supported for this migration in SQLite")
			},
		},
	}
}
