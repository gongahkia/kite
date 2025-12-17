package commands

import (
	"fmt"

	"github.com/gongahkia/kite/internal/config"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/spf13/cobra"
)

// NewMigrateCmd creates the migrate command
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		Long:  "Manage database schema migrations (up, down, status, version)",
	}

	cmd.AddCommand(newMigrateUpCmd())
	cmd.AddCommand(newMigrateDownCmd())
	cmd.AddCommand(newMigrateStatusCmd())
	cmd.AddCommand(newMigrateVersionCmd())
	cmd.AddCommand(newMigrateCreateCmd())

	return cmd
}

func newMigrateUpCmd() *cobra.Command {
	var steps int

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply pending migrations",
		Long:  "Apply all pending database migrations or a specific number of steps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			db, err := initStorage(cfg)
			if err != nil {
				return err
			}
			defer db.Close()

			migrator, err := storage.NewMigrator(db, "migrations")
			if err != nil {
				return fmt.Errorf("failed to create migrator: %w", err)
			}

			if steps > 0 {
				fmt.Printf("Applying %d migration(s)...\n", steps)
				return migrator.UpSteps(steps)
			}

			fmt.Println("Applying all pending migrations...")
			return migrator.Up()
		},
	}

	cmd.Flags().IntVarP(&steps, "steps", "n", 0, "Number of migrations to apply (0 = all)")

	return cmd
}

func newMigrateDownCmd() *cobra.Command {
	var steps int

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback migrations",
		Long:  "Rollback database migrations by a specific number of steps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			db, err := initStorage(cfg)
			if err != nil {
				return err
			}
			defer db.Close()

			migrator, err := storage.NewMigrator(db, "migrations")
			if err != nil {
				return fmt.Errorf("failed to create migrator: %w", err)
			}

			if steps <= 0 {
				return fmt.Errorf("steps must be greater than 0")
			}

			fmt.Printf("Rolling back %d migration(s)...\n", steps)
			return migrator.DownSteps(steps)
		},
	}

	cmd.Flags().IntVarP(&steps, "steps", "n", 1, "Number of migrations to rollback")
	cmd.MarkFlagRequired("steps")

	return cmd
}

func newMigrateStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  "Display the current status of all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			db, err := initStorage(cfg)
			if err != nil {
				return err
			}
			defer db.Close()

			migrator, err := storage.NewMigrator(db, "migrations")
			if err != nil {
				return fmt.Errorf("failed to create migrator: %w", err)
			}

			status, err := migrator.Status()
			if err != nil {
				return err
			}

			fmt.Println("Migration Status:")
			fmt.Println("================")
			for _, m := range status {
				appliedMark := " "
				if m.Applied {
					appliedMark = "✓"
				}
				fmt.Printf("[%s] %s (version: %d)\n", appliedMark, m.Name, m.Version)
				if m.AppliedAt != nil {
					fmt.Printf("    Applied at: %s\n", m.AppliedAt.Format("2006-01-02 15:04:05"))
				}
			}

			return nil
		},
	}
}

func newMigrateVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current schema version",
		Long:  "Display the current database schema version",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			db, err := initStorage(cfg)
			if err != nil {
				return err
			}
			defer db.Close()

			migrator, err := storage.NewMigrator(db, "migrations")
			if err != nil {
				return fmt.Errorf("failed to create migrator: %w", err)
			}

			version, err := migrator.Version()
			if err != nil {
				return err
			}

			fmt.Printf("Current schema version: %d\n", version)
			return nil
		},
	}
}

func newMigrateCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new migration file",
		Long:  "Generate a new migration file with the given name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			db, err := initStorage(cfg)
			if err != nil {
				return err
			}
			defer db.Close()

			migrator, err := storage.NewMigrator(db, "migrations")
			if err != nil {
				return fmt.Errorf("failed to create migrator: %w", err)
			}

			filename, err := migrator.Create(name)
			if err != nil {
				return err
			}

			fmt.Printf("Created migration file: %s\n", filename)
			return nil
		},
	}
}

// Helper functions

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	configPath, _ := cmd.Flags().GetString("config")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Loading config from: %s\n", configPath)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	db, err := storage.NewStorage(cfg.Database.Driver, cfg.Database.URL, storage.StorageConfig{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return db, nil
}
