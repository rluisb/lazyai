package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git integration",
	Long:  `Integrate LazyAI with git for traceability.`,
}

var gitSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync changes with git",
	Long:  `Auto-commit all changes with descriptive messages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			message = fmt.Sprintf("lazyai: Auto-commit at %s", time.Now().Format(time.RFC3339))
		}
		
		// Check if we're in a git repo
		if err := runGitCommand("rev-parse", "--git-dir"); err != nil {
			return fmt.Errorf("not a git repository. Run 'git init' first")
		}
		
		// Check for changes
		status, err := runGitOutput("status", "--porcelain")
		if err != nil {
			return fmt.Errorf("error checking git status: %w", err)
		}
		
		if strings.TrimSpace(status) == "" {
			fmt.Println("✅ No changes to commit")
			return nil
		}
		
		// Add all changes
		fmt.Println("📝 Adding changes...")
		if err := runGitCommand("add", "-A"); err != nil {
			return fmt.Errorf("error adding changes: %w", err)
		}
		
		// Commit
		fmt.Printf("📝 Committing: %s\n", message)
		if err := runGitCommand("commit", "-m", message); err != nil {
			return fmt.Errorf("error committing: %w", err)
		}
		
		// Get commit hash
		commitHash, err := runGitOutput("rev-parse", "--short", "HEAD")
		if err != nil {
			commitHash = "unknown"
		}
		
		fmt.Printf("✅ Committed: %s\n", strings.TrimSpace(commitHash))
		
		// Record to ledger
		appendToLedger("git_commit", map[string]string{
			"hash":    strings.TrimSpace(commitHash),
			"message": message,
		})
		
		return nil
	},
}

var gitLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Show agent-attributed commits",
	Long:  `Show recent commits with LazyAI attribution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		if limit <= 0 {
			limit = 10
		}
		
		// Get git log with custom format
		output, err := runGitOutput("log", fmt.Sprintf("-%d", limit), "--format=%h|%s|%an|%ad", "--date=short")
		if err != nil {
			return fmt.Errorf("error getting git log: %w", err)
		}
		
		fmt.Println("Recent Commits:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			
			parts := strings.SplitN(line, "|", 4)
			if len(parts) >= 3 {
				hash := parts[0]
				message := parts[1]
				author := parts[2]
				date := ""
				if len(parts) >= 4 {
					date = parts[3]
				}
				
				// Highlight LazyAI commits
				if strings.Contains(message, "lazyai") || strings.Contains(author, "LazyAI") {
					fmt.Printf("  🤖 %s | %s | %s | %s\n", hash, date, author, message)
				} else {
					fmt.Printf("     %s | %s | %s | %s\n", hash, date, author, message)
				}
			}
		}
		
		return nil
	},
}

var gitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show git status",
	Long:  `Show current git status with LazyAI context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := runGitOutput("status")
		if err != nil {
			return fmt.Errorf("error getting git status: %w", err)
		}
		
		fmt.Println(output)
		
		// Show current session if any
		database, err := EnsureDB()
		if err == nil {
			defer SafeCloseDB(database)
			
			var activeSession string
			database.QueryRow("SELECT id FROM sessions WHERE status = 'active' ORDER BY started_at DESC LIMIT 1").Scan(&activeSession)
			
			if activeSession != "" {
				fmt.Printf("\n🟢 Active Session: %s\n", activeSession)
			}
		}
		
		return nil
	},
}

// runGitCommand runs a git command and returns error if it fails
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runGitOutput runs a git command and returns output
func runGitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func init() {
	gitSyncCmd.Flags().StringP("message", "m", "", "Commit message")
	gitLogCmd.Flags().IntP("limit", "n", 10, "Number of commits to show")
	
	gitCmd.AddCommand(gitSyncCmd)
	gitCmd.AddCommand(gitLogCmd)
	gitCmd.AddCommand(gitStatusCmd)
	rootCmd.AddCommand(gitCmd)
}
