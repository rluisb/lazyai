package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	sidecarpkg "github.com/rluisb/lazyai/packages/cli/internal/sidecar"
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
	Use:   "status [--scope workspace|project|global]",
	Short: "Show resolved sidecar paths for the current scope",
	RunE:  runSidecarStatus,
}

var sidecarAttachCmd = &cobra.Command{
	Use:   "attach [<project-path>] [--scope workspace|project]",
	Short: "Attach a sidecar to a workspace or project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSidecarAttach,
}

var sidecarDetachCmd = &cobra.Command{
	Use:   "detach [<project-path>] [--scope workspace|project]",
	Short: "Detach a sidecar from a workspace or project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSidecarDetach,
}

var sidecarDoctorCmd = &cobra.Command{
	Use:   "doctor [--scope workspace|project|global]",
	Short: "Validate sidecar configuration and paths",
	RunE:  runSidecarDoctor,
}

func init() {
	sidecarInitCmd.Flags().String("scope", "", "Scope: workspace|project|global (default: workspace if active workspace exists, else project)")
	sidecarInitCmd.Flags().String("path", "", "Sidecar root path (required)")
	sidecarInitCmd.Flags().String("specs-dir", "", "Override specs directory name")
	sidecarInitCmd.Flags().String("docs-dir", "", "Override docs directory name")
	sidecarInitCmd.Flags().String("plans-dir", "", "Override plans directory name")
	_ = sidecarInitCmd.MarkFlagRequired("path")

	sidecarStatusCmd.Flags().String("scope", "", "Scope: workspace|project|global (default: auto)")

	sidecarAttachCmd.Flags().String("scope", "", "Scope: workspace|project (default: workspace)")
	sidecarAttachCmd.Flags().String("path", "", "Sidecar root path (required)")
	sidecarAttachCmd.Flags().String("specs-dir", "", "Override specs directory name")
	sidecarAttachCmd.Flags().String("docs-dir", "", "Override docs directory name")
	sidecarAttachCmd.Flags().String("plans-dir", "", "Override plans directory name")
	_ = sidecarAttachCmd.MarkFlagRequired("path")

	sidecarDetachCmd.Flags().String("scope", "", "Scope: workspace|project (default: workspace)")
	sidecarDetachCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	sidecarDoctorCmd.Flags().String("scope", "", "Scope: workspace|project|global (default: auto)")

	sidecarCmd.AddCommand(sidecarInitCmd)
	sidecarCmd.AddCommand(sidecarStatusCmd)
	sidecarCmd.AddCommand(sidecarAttachCmd)
	sidecarCmd.AddCommand(sidecarDetachCmd)
	sidecarCmd.AddCommand(sidecarDoctorCmd)
	rootCmd.AddCommand(sidecarCmd)
}

// ---------------------------------------------------------------------------
// Command runners
// ---------------------------------------------------------------------------

func runSidecarInit(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd, true)
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

	switch scope {
	case sidecarpkg.ScopeWorkspace:
		wsCfg, err := loadWorkspaceConfig()
		if err != nil {
			return err
		}
		if wsCfg.Active == "" {
			return fmt.Errorf("no active workspace set; use 'lazyai-cli workspace switch' first")
		}
		w := findWorkspace(wsCfg, wsCfg.Active)
		if w == nil {
			return fmt.Errorf("active workspace '%s' not found", wsCfg.Active)
		}
		w.Sidecar = cfg
		if err := saveWorkspaceConfig(wsCfg); err != nil {
			return err
		}
		fmt.Printf("✅ Sidecar initialized for workspace '%s'\n", w.Name)
	case sidecarpkg.ScopeProject:
		projectRoot, err := getProjectRoot()
		if err != nil {
			return err
		}
		if err := sidecarpkg.WriteProjectSidecar(projectRoot, cfg); err != nil {
			return err
		}
		fmt.Printf("✅ Sidecar initialized for project: %s\n", projectRoot)
	case sidecarpkg.ScopeGlobal:
		if err := sidecarpkg.WriteGlobalSidecar(cfg); err != nil {
			return err
		}
		fmt.Printf("✅ Global sidecar initialized\n")
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
	scope, err := determineScope(cmd, false)
	if err != nil {
		return err
	}

	projectRoot, _ := getProjectRoot()
	resolved, err := sidecarpkg.Resolve(scope, projectRoot)
	if err != nil {
		return err
	}

	// Table format: Scope, Config Level, Docs Dir, Specs Dir, Plans Dir
	fmt.Println("Sidecar Status")
	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("%-12s %-14s %-22s %-22s %-22s\n", "Scope", "Config Level", "Docs Dir", "Specs Dir", "Plans Dir")
	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("%-12s %-14s %-22s %-22s %-22s\n",
		scope.String(),
		resolved.ConfigLevel,
		truncate(resolved.DocsDir, 20),
		truncate(resolved.SpecsDir, 20),
		truncate(resolved.PlansDir, 20),
	)
	fmt.Println(strings.Repeat("-", 90))

	if resolved.ConfigLevel == "default" {
		fmt.Println()
		fmt.Println("  (No sidecar configured — using scope default paths)")
	}
	return nil
}

func runSidecarAttach(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd, true)
	if err != nil {
		return err
	}

	// Determine project path: positional arg, or active workspace path, or current directory
	var absProject string
	if len(args) > 0 {
		absProject, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving project path: %w", err)
		}
	} else if scope == sidecarpkg.ScopeWorkspace {
		wsCfg, err := loadWorkspaceConfig()
		if err != nil {
			return err
		}
		if wsCfg.Active == "" {
			return fmt.Errorf("no active workspace set; use 'lazyai-cli workspace switch' first")
		}
		w := findWorkspace(wsCfg, wsCfg.Active)
		if w == nil {
			return fmt.Errorf("active workspace '%s' not found", wsCfg.Active)
		}
		absProject, _ = filepath.Abs(w.Path)
	} else {
		absProject, err = getProjectRoot()
		if err != nil {
			return err
		}
	}

	path, _ := cmd.Flags().GetString("path")
	if err := ValidateNotEmpty(path, "path"); err != nil {
		return err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving sidecar path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("sidecar path is a file, not a directory: %s", absPath)
	}

	cfg := &sidecarpkg.SidecarConfig{
		Path:     absPath,
		SpecsDir: defaultDir(cmd, "specs-dir", "specs"),
		DocsDir:  defaultDir(cmd, "docs-dir", "docs"),
		PlansDir: defaultDir(cmd, "plans-dir", "plans"),
	}

	switch scope {
	case sidecarpkg.ScopeWorkspace:
		wsCfg, err := loadWorkspaceConfig()
		if err != nil {
			return err
		}
		if wsCfg.Active == "" {
			return fmt.Errorf("no active workspace set; use 'lazyai-cli workspace switch' first")
		}
		w := findWorkspace(wsCfg, wsCfg.Active)
		if w == nil {
			return fmt.Errorf("active workspace '%s' not found", wsCfg.Active)
		}
		wAbs, _ := filepath.Abs(w.Path)
		if wAbs != absProject {
			return fmt.Errorf("active workspace path (%s) does not match requested project path (%s)", wAbs, absProject)
		}
		w.Sidecar = cfg
		if err := saveWorkspaceConfig(wsCfg); err != nil {
			return err
		}
		fmt.Printf("✅ Sidecar attached to workspace '%s'\n", w.Name)
	case sidecarpkg.ScopeProject:
		if err := sidecarpkg.WriteProjectSidecar(absProject, cfg); err != nil {
			return err
		}
		fmt.Printf("✅ Sidecar attached to project: %s\n", absProject)
	case sidecarpkg.ScopeGlobal:
		return fmt.Errorf("attach does not support global scope; use 'sidecar init --scope global'")
	}

	fmt.Printf("   Path:  %s\n", cfg.Path)
	fmt.Printf("   Docs:  %s\n", resolveDir(cfg.Path, cfg.DocsDir))
	fmt.Printf("   Specs: %s\n", resolveDir(cfg.Path, cfg.SpecsDir))
	fmt.Printf("   Plans: %s\n", resolveDir(cfg.Path, cfg.PlansDir))

	_ = appendToLedger("sidecar_attach", map[string]string{
		"scope":        scope.String(),
		"path":         cfg.Path,
		"project_path": absProject,
	})
	return nil
}

func runSidecarDetach(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd, true)
	if err != nil {
		return err
	}

	// Determine project path: positional arg, or active workspace path, or current directory
	var absProject string
	if len(args) > 0 {
		absProject, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving project path: %w", err)
		}
	} else if scope == sidecarpkg.ScopeWorkspace {
		wsCfg, err := loadWorkspaceConfig()
		if err != nil {
			return err
		}
		if wsCfg.Active == "" {
			fmt.Println("No active workspace set.")
			return nil
		}
		w := findWorkspace(wsCfg, wsCfg.Active)
		if w == nil {
			fmt.Printf("Active workspace '%s' not found.\n", wsCfg.Active)
			return nil
		}
		absProject, _ = filepath.Abs(w.Path)
	} else {
		absProject, err = getProjectRoot()
		if err != nil {
			return err
		}
	}

	force, _ := cmd.Flags().GetBool("force")

	switch scope {
	case sidecarpkg.ScopeWorkspace:
		wsCfg, err := loadWorkspaceConfig()
		if err != nil {
			return err
		}
		if wsCfg.Active == "" {
			fmt.Println("No active workspace set.")
			return nil
		}
		w := findWorkspace(wsCfg, wsCfg.Active)
		if w == nil {
			fmt.Printf("Active workspace '%s' not found.\n", wsCfg.Active)
			return nil
		}
		if w.Sidecar == nil {
			fmt.Println("No sidecar configured for active workspace.")
			return nil
		}
		wAbs, _ := filepath.Abs(w.Path)
		if wAbs != absProject {
			return fmt.Errorf("active workspace path (%s) does not match requested project path (%s)", wAbs, absProject)
		}
		if !force {
			fmt.Printf("Detach sidecar from workspace '%s'?\n", w.Name)
			fmt.Printf("  Sidecar path: %s\n", w.Sidecar.Path)
			fmt.Printf("  This will remove the sidecar block from the workspace config. [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		w.Sidecar = nil
		if err := saveWorkspaceConfig(wsCfg); err != nil {
			return err
		}
		fmt.Printf("✅ Sidecar detached from workspace '%s'\n", w.Name)
	case sidecarpkg.ScopeProject:
		cfg, err := sidecarpkg.LoadProjectSidecar(absProject)
		if err != nil {
			return err
		}
		if cfg == nil {
			fmt.Println("No sidecar configured for project.")
			return nil
		}
		if !force {
			fmt.Printf("Detach sidecar from project '%s'?\n", absProject)
			fmt.Printf("  Sidecar path: %s\n", cfg.Path)
			fmt.Printf("  This will remove the project sidecar file (.lazyai-sidecar.yaml). [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		if err := sidecarpkg.RemoveProjectSidecar(absProject); err != nil {
			return err
		}
		fmt.Printf("✅ Sidecar detached from project: %s\n", absProject)
	case sidecarpkg.ScopeGlobal:
		return fmt.Errorf("detach does not support global scope; use 'sidecar init --scope global' to update or remove ~/.lazyai/sidecar.yaml manually")
	}

	_ = appendToLedger("sidecar_detach", map[string]string{
		"scope":        scope.String(),
		"project_path": absProject,
	})
	return nil
}

func runSidecarDoctor(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd, false)
	if err != nil {
		return err
	}

	projectRoot, _ := getProjectRoot()
	issues, err := sidecarpkg.Doctor(scope, projectRoot)
	if err != nil {
		return err
	}

	if len(issues) == 0 {
		fmt.Printf("✅ Sidecar doctor: all paths valid for %s scope.\n", scope.String())
		return nil
	}

	hasErrors := false
	for _, issue := range issues {
		fmt.Printf("%s: %s\n", issue.Severity, issue.Message)
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

func determineScope(cmd *cobra.Command, requireWorkspaceFallback bool) (sidecarpkg.Scope, error) {
	scopeFlag, _ := cmd.Flags().GetString("scope")
	if scopeFlag != "" {
		return parseScope(scopeFlag)
	}

	wsCfg, err := loadWorkspaceConfig()
	if err != nil {
		return sidecarpkg.ScopeProject, nil
	}
	if wsCfg.Active != "" && findWorkspace(wsCfg, wsCfg.Active) != nil {
		return sidecarpkg.ScopeWorkspace, nil
	}
	if requireWorkspaceFallback {
		return sidecarpkg.ScopeProject, nil
	}
	return sidecarpkg.ScopeProject, nil
}

func getProjectRoot() (string, error) {
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
