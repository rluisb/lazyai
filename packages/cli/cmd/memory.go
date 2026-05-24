package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Long-term memory vault",
	Long:  `Save and search durable knowledge across sessions.`,
}

var memorySaveCmd = &cobra.Command{
	Use:   "save [content]",
	Short: "Save a memory",
	Long:  `Save a lesson, context, or decision for future sessions.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content := strings.Join(args, " ")

		memoryType, _ := cmd.Flags().GetString("type")
		if memoryType == "" {
			memoryType = "lesson"
		}

		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Create memory file
		memoryDir := filepath.Join(".specify", "memory")
		if err := os.MkdirAll(memoryDir, 0755); err != nil {
			return fmt.Errorf("error creating memory directory: %w", err)
		}

		// Generate filename
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_%s.md", timestamp, memoryType)
		filepath := filepath.Join(memoryDir, filename)

		// Build memory content
		var memoryContent strings.Builder
		memoryContent.WriteString(fmt.Sprintf("# %s\n\n", strings.Title(memoryType)))
		memoryContent.WriteString(fmt.Sprintf("**Created:** %s\n\n", time.Now().Format(time.RFC3339)))

		if len(tags) > 0 {
			memoryContent.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(tags, ", ")))
		}

		memoryContent.WriteString("## Content\n\n")
		memoryContent.WriteString(content)
		memoryContent.WriteString("\n")

		// Write to file
		if err := os.WriteFile(filepath, []byte(memoryContent.String()), 0644); err != nil {
			return fmt.Errorf("error writing memory file: %w", err)
		}

		fmt.Printf("✅ Memory saved: %s\n", filename)
		fmt.Printf("   Type: %s\n", memoryType)
		if len(tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(tags, ", "))
		}

		// Record to ledger
		appendToLedger("memory_saved", map[string]string{
			"file": filename,
			"type": memoryType,
		})

		return nil
	},
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all memories",
	Long:  `List all saved memories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		memoryDir := filepath.Join(".specify", "memory")

		entries, err := os.ReadDir(memoryDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No memories found.")
				return nil
			}
			return fmt.Errorf("error reading memory directory: %w", err)
		}

		fmt.Println("Memories:")
		fmt.Println("───────────────────────────────────────────────────────────────")

		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				count++
				fmt.Printf("  • %s\n", entry.Name())
			}
		}

		if count == 0 {
			fmt.Println("  No memories found.")
		} else {
			fmt.Printf("\n  Total: %d memories\n", count)
		}

		return nil
	},
}

var memorySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search memories",
	Long:  `Search through saved memories by content.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.ToLower(strings.Join(args, " "))

		memoryDir := filepath.Join(".specify", "memory")

		entries, err := os.ReadDir(memoryDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No memories found.")
				return nil
			}
			return fmt.Errorf("error reading memory directory: %w", err)
		}

		fmt.Printf("Searching for: %s\n", query)
		fmt.Println("───────────────────────────────────────────────────────────────")

		matches := 0
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			content, err := os.ReadFile(filepath.Join(memoryDir, entry.Name()))
			if err != nil {
				continue
			}

			if strings.Contains(strings.ToLower(string(content)), query) {
				matches++
				fmt.Printf("  ✅ %s\n", entry.Name())

				// Show first matching line
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.Contains(strings.ToLower(line), query) {
						fmt.Printf("     %s\n", strings.TrimSpace(line))
						break
					}
				}
			}
		}

		if matches == 0 {
			fmt.Println("  No matches found.")
		} else {
			fmt.Printf("\n  Found %d matches\n", matches)
		}

		return nil
	},
}

func init() {
	memorySaveCmd.Flags().StringP("type", "t", "lesson", "Memory type (lesson, context, decision, idea)")
	memorySaveCmd.Flags().StringSlice("tags", []string{}, "Tags for categorization")

	memoryCmd.AddCommand(memorySaveCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memorySearchCmd)
	rootCmd.AddCommand(memoryCmd)
}
