package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewConfigCmd creates the config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long:  "View and validate configuration",
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigValidateCmd())
	cmd.AddCommand(newConfigEnvCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current configuration with sensitive values redacted",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			// Redact sensitive values
			cfgCopy := *cfg
			cfgCopy.Database.URL = redactURL(cfgCopy.Database.URL)
			cfgCopy.Cache.URL = redactURL(cfgCopy.Cache.URL)
			cfgCopy.Queue.URL = redactURL(cfgCopy.Queue.URL)

			switch format {
			case "json":
				data, err := json.MarshalIndent(cfgCopy, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))

			case "yaml":
				data, err := yaml.Marshal(cfgCopy)
				if err != nil {
					return err
				}
				fmt.Print(string(data))

			default:
				return fmt.Errorf("unsupported format: %s (use json or yaml)", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format (json, yaml)")

	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  "Check configuration for errors and warnings",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			fmt.Println("Validating configuration...")

			errors := 0
			warnings := 0

			// Check database configuration
			if cfg.Database.MaxOpenConns > 100 {
				fmt.Println("⚠ Warning: Database max_open_conns is very high (>100)")
				warnings++
			}

			if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
				fmt.Println("✗ Error: max_idle_conns cannot exceed max_open_conns")
				errors++
			}

			// Check worker configuration
			if cfg.Workers.Count < 1 {
				fmt.Println("✗ Error: Worker count must be at least 1")
				errors++
			}

			if cfg.Workers.QueueSize < 100 {
				fmt.Println("⚠ Warning: Worker queue size is quite small (<100)")
				warnings++
			}

			// Check cache configuration
			if cfg.Cache.MaxKeys < 1000 {
				fmt.Println("⚠ Warning: Cache max keys is low (<1000)")
				warnings++
			}

			// Summary
			fmt.Println()
			if errors == 0 && warnings == 0 {
				fmt.Println("✓ Configuration is valid")
			} else {
				fmt.Printf("Found %d error(s) and %d warning(s)\n", errors, warnings)
			}

			if errors > 0 {
				return fmt.Errorf("configuration validation failed")
			}

			return nil
		},
	}
}

func newConfigEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Show environment variables",
		Long:  "Display environment variables used by Kite",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Environment Variables:")
			fmt.Println("======================")
			fmt.Println()

			envVars := []struct {
				name        string
				description string
				example     string
			}{
				{"DATABASE_URL", "Database connection string", "postgres://user:pass@localhost:5432/kite"},
				{"CACHE_URL", "Cache connection string", "redis://localhost:6379"},
				{"QUEUE_URL", "Queue connection string", "nats://localhost:4222"},
				{"LOG_LEVEL", "Logging level", "info"},
				{"METRICS_ENABLED", "Enable metrics", "true"},
				{"JWT_SECRET", "JWT signing secret", "your-secret-key"},
			}

			for _, env := range envVars {
				fmt.Printf("%s\n", env.name)
				fmt.Printf("  Description: %s\n", env.description)
				fmt.Printf("  Example:     %s\n", env.example)
				fmt.Println()
			}

			return nil
		},
	}
}

func redactURL(url string) string {
	// Simple redaction - in real implementation, would parse and redact password
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-5:]
	}
	return "***"
}
