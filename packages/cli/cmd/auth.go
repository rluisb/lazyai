package cmd

import (
	"fmt"
	"os"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Manage authentication for providers and services.",
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured authentication",
	Long:  "List configured AI providers and basic authentication environment variables.",
	RunE:  runAuthList,
}

func init() {
	authCmd.AddCommand(authListCmd)
	rootCmd.AddCommand(authCmd)
	authCmd.GroupID = "auth"
}

func runAuthList(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Read store data
	dbPath := db.DefaultDBPath(dir)
	var providers []string
	if _, err := os.Stat(dbPath); err == nil {
		database, err := db.Open(dbPath)
		if err == nil {
			defer database.Close()
			if err := db.RunMigrations(database); err == nil {
				store := db.NewStore(database)
				if data, err := store.ReadStoreData(); err == nil && data != nil {
					providers = data.Selections.OpenCodeProviders
				}
			}
		}
	}

	fmt.Println("=== Configured AI Providers ===")
	if len(providers) > 0 {
		for _, p := range providers {
			fmt.Printf("- %s\n", p)
		}
	} else {
		fmt.Println("No providers configured in the workspace store.")
	}
	fmt.Println()

	// Check common auth environment variables
	fmt.Println("=== Common Authentication Variables ===")
	envVars := []string{
		"OPENAI_API_KEY",
		"ANTHROPIC_API_KEY",
		"GOOGLE_API_KEY",
		"GITHUB_TOKEN",
	}

	foundAny := false
	for _, env := range envVars {
		if val, exists := os.LookupEnv(env); exists {
			foundAny = true
			masked := "*****"
			if len(val) > 4 {
				masked = "*****" + val[len(val)-4:]
			}
			fmt.Printf("- %s: SET (%s)\n", env, masked)
		} else {
			fmt.Printf("- %s: NOT SET\n", env)
		}
	}

	if !foundAny {
		fmt.Println("\nNote: Ensure you set required API keys in your environment for the selected providers.")
	}

	return nil
}
