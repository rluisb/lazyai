package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate agent and skill files",
	Long:  "Check that agent files and skill files follow the correct structure and tool schemas.",
}

var validateAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Validate agent files",
	RunE:  runValidateAgents,
}

var validateSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Validate skill files",
	RunE:  runValidateSkills,
}

func init() {
	validateCmd.AddCommand(validateAgentsCmd)
	validateCmd.AddCommand(validateSkillsCmd)
	rootCmd.AddCommand(validateCmd)
	validateCmd.GroupID = "audit"
}

// ValidationResult represents a single validation issue
type ValidationResult struct {
	File     string `json:"file"`
	Severity string `json:"severity"` // error, warning
	Message  string `json:"message"`
}

func runValidateAgents(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	agentsDir := filepath.Join(dir, ".opencode", "agents")

	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return fmt.Errorf("agents directory not found: %s", agentsDir)
	}

	var results []ValidationResult
	var passCount, failCount int

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		contentStr := string(content)
		fileName := entry.Name()
		hasError := false

		// Check for frontmatter
		if !strings.HasPrefix(contentStr, "---\n") {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "error",
				Message:  "Missing YAML frontmatter",
			})
			failCount++
			hasError = true
		}

		// Check for System Prompt
		systemPromptIndex := strings.Index(contentStr, "# System Prompt")
		if systemPromptIndex == -1 {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "error",
				Message:  "Missing '# System Prompt' heading",
			})
			failCount++
			hasError = true
		}

		// Check for managed marker
		if !strings.Contains(contentStr, "vibe-lab:managed kind=agent") {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "warning",
				Message:  "Missing 'vibe-lab:managed kind=agent' marker",
			})
		}

		// Check for a section heading after System Prompt
		if systemPromptIndex != -1 {
			afterSystemPrompt := contentStr[systemPromptIndex:]
			if !strings.Contains(afterSystemPrompt, "## ") {
				results = append(results, ValidationResult{
					File:     fileName,
					Severity: "warning",
					Message:  "Missing section heading after 'System Prompt'",
				})
			}
		}

		if !hasError {
			passCount++
		}
	}

	// Output results
	fmt.Println("🔍 Agent Validation Results")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("✅ All agents pass validation")
		fmt.Printf("   Checked: %d files\n", passCount)
		return nil
	}

	for _, result := range results {
		var emoji string
		if result.Severity == "error" {
			emoji = "❌"
		} else {
			emoji = "⚠️"
		}
		fmt.Printf("  %s %s: %s\n", emoji, result.File, result.Message)
	}

	fmt.Println()
	fmt.Printf("  Summary: %d passed, %d issues found\n", passCount, len(results))

	// Append to ledger
	status := "pass"
	if failCount > 0 {
		status = "fail"
	}
	_ = appendToLedger("validate_agents", map[string]string{
		"status":         status,
		"agents_checked": fmt.Sprintf("%d", passCount+failCount),
		"issues_found":   fmt.Sprintf("%d", len(results)),
	})

	if failCount > 0 {
		return fmt.Errorf("validation failed: %d errors", failCount)
	}

	return nil
}

func runValidateSkills(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	skillsDir := filepath.Join(dir, ".opencode", "skills")

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return fmt.Errorf("skills directory not found: %s", skillsDir)
	}

	fmt.Println("🔍 Skill Validation Results")
	fmt.Println()
	fmt.Println("  ℹ️  Skill validation not yet implemented")
	fmt.Println("  Future checks will include:")
	fmt.Println("    - Quick Reference section presence")
	fmt.Println("    - Proper frontmatter format")
	fmt.Println("    - Script references validity")

	return nil
}
