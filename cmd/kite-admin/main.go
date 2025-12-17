package main

import (
	"fmt"
	"os"

	"github.com/gongahkia/kite/internal/admin/commands"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kite-admin",
		Short: "Kite administration CLI tool",
		Long: `kite-admin is the administration tool for Kite Legal Case Law Platform.

It provides commands for managing database migrations, workers, cache,
and other administrative tasks.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
	}

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "configs/default.yaml", "Config file path")
	rootCmd.PersistentFlags().StringP("env", "e", "development", "Environment (development, staging, production)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output in JSON format")

	// Add command groups
	rootCmd.AddCommand(commands.NewMigrateCmd())
	rootCmd.AddCommand(commands.NewWorkerCmd())
	rootCmd.AddCommand(commands.NewCacheCmd())
	rootCmd.AddCommand(commands.NewQueueCmd())
	rootCmd.AddCommand(commands.NewHealthCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewMetricsCmd())
	rootCmd.AddCommand(commands.NewBackupCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
