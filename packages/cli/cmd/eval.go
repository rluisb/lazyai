package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Run evaluation suites",
	Long:  `Run evaluation suites to measure agent quality and performance.`,
}

var evalRunCmd = &cobra.Command{
	Use:   "run [suite-name]",
	Short: "Run an evaluation suite",
	Long:  `Run a specific evaluation suite and generate results.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		suiteName := args[0]

		// Check if suite exists
		suitePath := filepath.Join(".specify", "evals", "suites", suiteName+".yaml")
		if _, err := os.Stat(suitePath); os.IsNotExist(err) {
			fmt.Printf("❌ Evaluation suite '%s' not found at %s\n", suiteName, suitePath)
			fmt.Println("\nAvailable suites:")
			listEvalSuites()
			return
		}

		fmt.Printf("🧪 Running evaluation suite: %s\n", suiteName)
		fmt.Println("═══════════════════════════════════════════════════════════════")

		// TODO: Implement actual evaluation logic
		// For now, show what would be evaluated
		fmt.Printf("\n📊 Suite: %s\n", suiteName)
		fmt.Println("   Status: Not yet implemented")
		fmt.Println("   This will run the evaluation suite and generate results")
		fmt.Println("   in .specify/evals/results/")

		// Record to ledger
		appendToLedger("eval_run", map[string]string{
			"suite":  suiteName,
			"status": "started",
		})
	},
}

var evalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available evaluation suites",
	Long:  `List all available evaluation suites.`,
	Run: func(cmd *cobra.Command, args []string) {
		listEvalSuites()
	},
}

func listEvalSuites() {
	suitesDir := filepath.Join(".specify", "evals", "suites")

	entries, err := os.ReadDir(suitesDir)
	if err != nil {
		fmt.Printf("❌ Error reading suites directory: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No evaluation suites found.")
		fmt.Println("Create suites in .specify/evals/suites/")
		return
	}

	fmt.Println("Available evaluation suites:")
	fmt.Println("───────────────────────────────────────────────────────────────")

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .yaml
			fmt.Printf("  • %s\n", name)
		}
	}
}

func init() {
	evalCmd.AddCommand(evalRunCmd)
	evalCmd.AddCommand(evalListCmd)
	rootCmd.AddCommand(evalCmd)
	evalCmd.GroupID = "audit"
}
