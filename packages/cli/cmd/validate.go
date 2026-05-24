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
		
		// Check for Dispatch Parameters section
		if !strings.Contains(contentStr, "## Dispatch Parameters") {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "error",
				Message:  "Missing '## Dispatch Parameters' section",
			})
			failCount++
		} else {
			passCount++
		}
		
		// Check for Tool Schema Quick Reference
		if !strings.Contains(contentStr, "## Tool Schema Quick Reference") {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "warning",
				Message:  "Missing '## Tool Schema Quick Reference' section",
			})
		}
		
		// Check for common mistakes
		if strings.Contains(contentStr, "`text`") && strings.Contains(contentStr, "todowrite") {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "warning",
				Message:  "May use 'text' instead of 'content' for todowrite",
			})
		}
		
		if strings.Contains(contentStr, "mode:") && strings.Contains(contentStr, "task(") {
			results = append(results, ValidationResult{
				File:     fileName,
				Severity: "warning",
				Message:  "May use 'mode' as top-level field in task dispatch",
			})
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
