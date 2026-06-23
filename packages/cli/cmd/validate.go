package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/evals"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/internal/validate"
	"github.com/spf13/cobra"
)

var (
	validateAllFlag     bool
	validateProfileFlag string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate agent, skill, and eval assets",
	Long:  "Check that agent, skill, and eval files follow the correct structure and tool schemas.",
	RunE:  runValidate,
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

var validateEvalsCmd = &cobra.Command{
	Use:   "evals",
	Short: "Validate eval cases, holdouts, and rubrics",
	RunE:  runValidateEvals,
}

func init() {
	validateCmd.AddCommand(validateAgentsCmd)
	validateCmd.AddCommand(validateSkillsCmd)
	validateCmd.AddCommand(validateEvalsCmd)
	rootCmd.AddCommand(validateCmd)
	validateCmd.GroupID = "audit"
	validateCmd.Flags().BoolVar(&validateAllFlag, "all", false, "Run all validators (skills, agents, hooks, MCP, secrets, path safety) over the canonical .ai/ tree")
	validateCmd.Flags().StringVar(&validateProfileFlag, "profile", "", "Safety profile: team (inline secrets fail) or personal (warn). Defaults to the manifest profile.")
}

// runValidate dispatches the bare `validate` command. With --all it runs the
// consolidated engine over .ai/; otherwise it prints help (subcommands handle
// the per-surface checks).
func runValidate(cmd *cobra.Command, args []string) error {
	if !validateAllFlag {
		return cmd.Help()
	}
	return runValidateAll(cmd)
}

func runValidateAll(cmd *cobra.Command) error {
	dir, _ := os.Getwd()
	root := dir

	// Honor workspace scope when a store is available; tolerate its absence so
	// `validate --all` works in CI on a bare checkout.
	if database, err := openStore(dir); err == nil {
		store := db.NewStore(database)
		if storeData, readErr := store.ReadStoreData(); readErr == nil {
			if storeData.Config.SetupScope == types.SetupScopeWorkspace && storeData.Config.WorkspaceRoot != "" {
				root = storeData.Config.WorkspaceRoot
			}
		}
		database.Close()
	}

	profile := resolveValidateProfile(root)
	report := validate.All(validate.Options{Root: root, Profile: profile})

	fmt.Println("🔍 Validation Results (validate --all)")
	fmt.Printf("   Profile: %s\n\n", profile)
	if len(report.Issues) == 0 {
		fmt.Println("✅ All checks passed")
	} else {
		for _, issue := range report.Issues {
			emoji := "⚠️"
			if issue.Severity == validate.SeverityError {
				emoji = "❌"
			}
			fmt.Printf("  %s [%s] %s: %s\n", emoji, issue.Rule, issue.File, issue.Message)
		}
		fmt.Println()
		fmt.Printf("  Summary: %d error(s), %d warning(s)\n", report.Errors(), report.Warnings())
	}

	status := "pass"
	if report.HasErrors() {
		status = "fail"
	}
	_ = appendToLedger("validate_all", map[string]string{
		"status":   status,
		"profile":  string(profile),
		"errors":   fmt.Sprintf("%d", report.Errors()),
		"warnings": fmt.Sprintf("%d", report.Warnings()),
	})

	if report.HasErrors() {
		return fmt.Errorf("validation failed: %d error(s)", report.Errors())
	}
	return nil
}

// resolveValidateProfile picks the safety profile: the --profile flag wins,
// then the manifest's profile field, defaulting to personal.
func resolveValidateProfile(root string) validate.Profile {
	if validateProfileFlag != "" {
		return validate.NormalizeProfile(validateProfileFlag)
	}
	if mf, err := aimanifest.Load(filepath.Join(root, ".ai")); err == nil && mf.Profile != "" {
		return validate.NormalizeProfile(mf.Profile)
	}
	return validate.ProfilePersonal
}

// ValidationResult represents a single validation issue
type ValidationResult struct {
	File     string `json:"file"`
	Severity string `json:"severity"` // error, warning
	Message  string `json:"message"`
}

// runValidateAgents validates agent files from the canonical .ai/agents
// directory using the consolidated validation engine. This replaces the
// legacy .opencode/agents check (FR-009, issue #408).
func runValidateAgents(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	aiDir := filepath.Join(dir, ".ai", "agents")

	if _, err := os.Stat(aiDir); os.IsNotExist(err) {
		return fmt.Errorf("agents directory not found: %s (run 'lazyai-cli init' first)", aiDir)
	}

	// Use the consolidated validation engine over the canonical .ai/ tree.
	profile := resolveValidateProfile(dir)
	report := validate.All(validate.Options{Root: dir, Profile: profile})

	// Filter to agent-related issues only.
	var results []ValidationResult
	var passCount, failCount int

	// Count agent files in .ai/agents.
	entries, err := os.ReadDir(aiDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			passCount++
		}
	}

	for _, issue := range report.Issues {
		if issue.Rule != "agent" {
			continue
		}
		sev := "warning"
		if issue.Severity == validate.SeverityError {
			sev = "error"
		}
		results = append(results, ValidationResult{
			File:     issue.File,
			Severity: sev,
			Message:  issue.Message,
		})
		if issue.Severity == validate.SeverityError {
			failCount++
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

// runValidateSkills validates skill files from the canonical .ai/skills
// directory using the consolidated validation engine. This replaces the
// legacy .opencode/skills check (FR-009, issue #408).
func runValidateSkills(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	aiDir := filepath.Join(dir, ".ai", "skills")

	if _, err := os.Stat(aiDir); os.IsNotExist(err) {
		return fmt.Errorf("skills directory not found: %s (run 'lazyai-cli init' first)", aiDir)
	}

	// Use the consolidated validation engine over the canonical .ai/ tree.
	profile := resolveValidateProfile(dir)
	report := validate.All(validate.Options{Root: dir, Profile: profile})

	// Filter to skill-related issues only.
	var results []ValidationResult
	var passCount, failCount int

	// Count skill files in .ai/skills (flat .md files and dir/SKILL.md).
	entries, err := os.ReadDir(aiDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				candidate := filepath.Join(aiDir, entry.Name(), "SKILL.md")
				if _, statErr := os.Stat(candidate); statErr == nil {
					passCount++
				}
				continue
			}
			if strings.HasSuffix(entry.Name(), ".md") {
				passCount++
			}
		}
	}

	for _, issue := range report.Issues {
		if issue.Rule != "skill" {
			continue
		}
		sev := "warning"
		if issue.Severity == validate.SeverityError {
			sev = "error"
		}
		results = append(results, ValidationResult{
			File:     issue.File,
			Severity: sev,
			Message:  issue.Message,
		})
		if issue.Severity == validate.SeverityError {
			failCount++
		}
	}

	// Output results
	fmt.Println("🔍 Skill Validation Results")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("✅ All skills pass validation")
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

	status := "pass"
	if failCount > 0 {
		status = "fail"
	}
	_ = appendToLedger("validate_skills", map[string]string{
		"status":         status,
		"skills_checked": fmt.Sprintf("%d", passCount+failCount),
		"issues_found":   fmt.Sprintf("%d", len(results)),
	})

	if failCount > 0 {
		return fmt.Errorf("validation failed: %d errors", failCount)
	}

	return nil
}

func runValidateEvals(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	evalsDir := filepath.Join(dir, ".ai", "evals")

	if _, err := os.Stat(evalsDir); os.IsNotExist(err) {
		return fmt.Errorf("evals directory not found: %s", evalsDir)
	}

	var results []ValidationResult
	var passCount, failCount int

	validateYAMLDir(filepath.Join(evalsDir, "cases"), []string{".yaml", ".yml"}, "case", evals.ValidateCase, &results, &passCount, &failCount)
	validateYAMLDir(filepath.Join(evalsDir, "holdouts"), []string{".yaml", ".yml"}, "holdout", evals.ValidateHoldout, &results, &passCount, &failCount)
	validateRubricDir(filepath.Join(evalsDir, "rubrics"), &results, &passCount, &failCount)

	fmt.Println("🔍 Evals Validation Results")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("✅ All eval assets pass validation")
		fmt.Printf("   Checked: %d files\n", passCount)
		return nil
	}

	for _, result := range results {
		emoji := "❌"
		fmt.Printf("  %s %s: %s\n", emoji, result.File, result.Message)
	}

	fmt.Println()
	fmt.Printf("  Summary: %d passed, %d issues found\n", passCount, len(results))

	status := "pass"
	if failCount > 0 {
		status = "fail"
	}
	_ = appendToLedger("validate_evals", map[string]string{
		"status":        status,
		"evals_checked": fmt.Sprintf("%d", passCount+failCount),
		"issues_found":  fmt.Sprintf("%d", len(results)),
	})
	if failCount > 0 {
		return fmt.Errorf("validation failed: %d errors", failCount)
	}
	return nil
}

func validateYAMLDir(dir string, extensions []string, category string, validateFn func(string, []byte) []evals.ValidationIssue, results *[]ValidationResult, passCount *int, failCount *int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		*results = append(*results, ValidationResult{
			File:     category + ": " + dir,
			Severity: "error",
			Message:  "directory unreadable: " + err.Error(),
		})
		*failCount++
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		allowed := false
		for _, a := range extensions {
			if ext == a {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			*results = append(*results, ValidationResult{
				File:     filepath.ToSlash(filepath.Join("evals", filepath.Base(dir), entry.Name())),
				Severity: "error",
				Message:  "unreadable: " + err.Error(),
			})
			*failCount++
			continue
		}

		issues := validateFn(filepath.ToSlash(filepath.Join("evals", filepath.Base(dir), entry.Name())), contents)
		if len(issues) == 0 {
			*passCount++
			continue
		}
		for _, issue := range issues {
			*results = append(*results, ValidationResult{
				File:     issue.File,
				Severity: "error",
				Message:  issue.Message,
			})
		}
		*failCount++
	}
}

func validateRubricDir(dir string, results *[]ValidationResult, passCount *int, failCount *int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Rubrics are optional; do not fail when missing in the evals tree.
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".md" && ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			*results = append(*results, ValidationResult{
				File:     filepath.ToSlash(filepath.Join("evals", "rubrics", entry.Name())),
				Severity: "error",
				Message:  "unreadable: " + err.Error(),
			})
			*failCount++
			continue
		}

		issues := evals.ValidateRubric(filepath.ToSlash(filepath.Join("evals", "rubrics", entry.Name())), contents)
		if len(issues) == 0 {
			*passCount++
			continue
		}
		for _, issue := range issues {
			*results = append(*results, ValidationResult{
				File:     issue.File,
				Severity: "error",
				Message:  issue.Message,
			})
		}
		*failCount++
	}
}

// frontmatterBlock returns the YAML frontmatter block (between the leading
// "---" fences) of a markdown document, or "" if none is present.
func frontmatterBlock(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return rest
	}
	return rest[:end]
}
