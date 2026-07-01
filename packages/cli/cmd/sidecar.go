package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sidecarpkg "github.com/rluisb/lazyai/packages/cli/internal/sidecar"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Cobra commands
// ---------------------------------------------------------------------------

var sidecarCmd = &cobra.Command{
	Use:   "sidecar",
	Short: "Manage LazyAI sidecar configuration",
	Long:  `Configure optional sidecar directories for docs, specs, and plans.`,
}

var sidecarInitCmd = &cobra.Command{
	Use:   "init --path <path> [--scope workspace|project|global]",
	Short: "Initialize a sidecar config at the given scope",
	RunE:  runSidecarInit,
}

var sidecarStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show resolved sidecar paths for the current scope",
	RunE:  runSidecarStatus,
}

var sidecarDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate sidecar configuration and paths",
	RunE:  runSidecarDoctor,
}

func init() {
	sidecarInitCmd.Flags().String("scope", "", "Scope label: workspace|project|global (default: project) — workspace and project write identically, see docs")
	sidecarInitCmd.Flags().String("path", "", "Sidecar root path (required)")
	sidecarInitCmd.Flags().String("specs-dir", "", "Override specs directory name")
	sidecarInitCmd.Flags().String("docs-dir", "", "Override docs directory name")
	sidecarInitCmd.Flags().String("plans-dir", "", "Override plans directory name")
	_ = sidecarInitCmd.MarkFlagRequired("path")

	sidecarCmd.AddCommand(sidecarInitCmd)
	sidecarCmd.AddCommand(sidecarStatusCmd)
	sidecarCmd.AddCommand(sidecarDoctorCmd)
	rootCmd.AddCommand(sidecarCmd)
	sidecarCmd.GroupID = "workspace"
}

// ---------------------------------------------------------------------------
// Command runners
// ---------------------------------------------------------------------------

func runSidecarInit(cmd *cobra.Command, args []string) error {
	scope := sidecarpkg.ScopeProject
	if scopeFlag, _ := cmd.Flags().GetString("scope"); scopeFlag != "" {
		parsed, err := parseScope(scopeFlag)
		if err != nil {
			return err
		}
		scope = parsed
	}

	cwd, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("getting project root: %w", err)
	}
	dir, err := sidecarpkg.ScopeRoot(scope, cwd)
	if err != nil {
		return err
	}

	path, _ := cmd.Flags().GetString("path")
	if err := ValidateNotEmpty(path, "path"); err != nil {
		return err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("path is a file, not a directory: %s", absPath)
	}

	cfg := &sidecarpkg.SidecarConfig{
		Path:     absPath,
		SpecsDir: defaultDir(cmd, "specs-dir", "specs"),
		DocsDir:  defaultDir(cmd, "docs-dir", "docs"),
		PlansDir: defaultDir(cmd, "plans-dir", "plans"),
	}

	if err := sidecarpkg.WriteSidecarAt(dir, cfg); err != nil {
		return err
	}

	if scope == sidecarpkg.ScopeGlobal {
		fmt.Printf("✅ Global sidecar initialized: %s\n", dir)
	} else {
		fmt.Printf("✅ Sidecar initialized (%s scope): %s\n", scope.String(), dir)
	}

	fmt.Printf("   Path:      %s\n", cfg.Path)
	fmt.Printf("   Docs:      %s\n", resolveDir(cfg.Path, cfg.DocsDir))
	fmt.Printf("   Specs:     %s\n", resolveDir(cfg.Path, cfg.SpecsDir))
	fmt.Printf("   Plans:     %s\n", resolveDir(cfg.Path, cfg.PlansDir))

	_ = appendToLedger("sidecar_init", map[string]string{
		"scope": scope.String(),
		"path":  cfg.Path,
	})
	return nil
}

func runSidecarStatus(cmd *cobra.Command, args []string) error {
	cwd, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("getting project root: %w", err)
	}
	layers, err := sidecarpkg.DiscoverLayers(cwd)
	if err != nil {
		return err
	}
	resolved, err := sidecarpkg.Resolve(cwd)
	if err != nil {
		return err
	}

	fmt.Printf("Sidecar Status (cwd: %s)\n\n", cwd)
	fmt.Println("Layers discovered:")
	printLayerLine("global", layers.Global)
	if layers.Workspace != nil {
		printLayerLine("workspace", *layers.Workspace)
	} else {
		fmt.Printf("  %-10s %-45s (not found)\n", "workspace", "")
	}
	if layers.Project != nil {
		printLayerLine("project", *layers.Project)
	} else {
		fmt.Printf("  %-10s %-45s (not found)\n", "project", "")
	}

	fmt.Println()
	fmt.Println("Resolved paths:")
	fmt.Printf("  %-10s %-25s (from: %s)\n", "docs_dir", resolved.DocsDir, resolved.Provenance["docs_dir"])
	fmt.Printf("  %-10s %-25s (from: %s)\n", "specs_dir", resolved.SpecsDir, resolved.Provenance["specs_dir"])
	fmt.Printf("  %-10s %-25s (from: %s)\n", "plans_dir", resolved.PlansDir, resolved.Provenance["plans_dir"])

	if resolved.IsAllDefault() {
		fmt.Println()
		fmt.Println("(No .lazyai/ configuration found — using built-in defaults. Run 'sidecar init' to configure.)")
	}
	return nil
}

// printLayerLine prints one discovered-layer row for `sidecar status`,
// showing the layer's sidecar.yaml path and whether it was actually found.
func printLayerLine(label string, layer sidecarpkg.Layer) {
	sidecarPath := filepath.Join(layer.Root, ".lazyai", "sidecar.yaml")
	status := "not found"
	if layer.Config != nil {
		status = "found"
	}
	if label == "global" && layer.Config == nil {
		status = "not found — built-in defaults"
	}
	fmt.Printf("  %-10s %-45s (%s)\n", label, sidecarPath, status)
}

func runSidecarDoctor(cmd *cobra.Command, args []string) error {
	cwd, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("getting project root: %w", err)
	}
	issues, err := sidecarpkg.Doctor(cwd)
	if err != nil {
		return err
	}

	if len(issues) == 0 {
		fmt.Println("✅ Sidecar doctor: all discovered layers valid.")
		return nil
	}

	hasErrors := false
	for _, issue := range issues {
		if issue.Level == "" {
			fmt.Printf("[general] %s: %s\n", issue.Severity, issue.Message)
		} else {
			fmt.Printf("[%s] %s: %s\n", issue.Level, issue.Severity, issue.Message)
		}
		if issue.Severity == sidecarpkg.IssueSeverityError {
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("sidecar doctor found errors")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

func parseScope(s string) (sidecarpkg.Scope, error) {
	switch strings.ToLower(s) {
	case "workspace":
		return sidecarpkg.ScopeWorkspace, nil
	case "project":
		return sidecarpkg.ScopeProject, nil
	case "global":
		return sidecarpkg.ScopeGlobal, nil
	default:
		return sidecarpkg.ScopeWorkspace, fmt.Errorf("invalid scope %q: use workspace, project, or global", s)
	}
}

var getProjectRoot = func() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return cwd, nil
}

func resolveDir(base, dir string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Clean(filepath.Join(base, dir))
}

func defaultDir(cmd *cobra.Command, flagName, def string) string {
	v, _ := cmd.Flags().GetString(flagName)
	if v == "" {
		return def
	}
	return v
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
