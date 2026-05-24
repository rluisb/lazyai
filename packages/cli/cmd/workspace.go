package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	sidecarpkg "github.com/rluisb/lazyai/packages/cli/internal/sidecar"
)

// WorkspaceConfig holds the global workspace registry.
type WorkspaceConfig = sidecarpkg.WorkspaceConfig

// LinkedProject is a cross-project reference within a sidecar.
type LinkedProject = sidecarpkg.LinkedProject

// SidecarConfig holds the sidecar configuration for a single scope level.
type SidecarConfig = sidecarpkg.SidecarConfig

// WorkspaceEntry represents a registered project/workspace.
type WorkspaceEntry = sidecarpkg.WorkspaceEntry

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage multi-project workspaces",
	Long:  `Register, list, switch, and inspect LazyAI project workspaces.`,
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered workspaces",
	RunE:  runWorkspaceList,
}

var workspaceAddCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Register a project path as a workspace",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkspaceAdd,
}

var workspaceSwitchCmd = &cobra.Command{
	Use:   "switch [name]",
	Short: "Set the active workspace by name",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkspaceSwitch,
}

var workspaceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active workspace details",
	RunE:  runWorkspaceStatus,
}

func init() {
	workspaceAddCmd.Flags().String("name", "", "Override workspace name (default: directory basename)")

	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceAddCmd)
	workspaceCmd.AddCommand(workspaceSwitchCmd)
	workspaceCmd.AddCommand(workspaceStatusCmd)
	rootCmd.AddCommand(workspaceCmd)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func getGlobalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(home, ".lazyai"), nil
}

func getWorkspacesConfigPath() (string, error) {
	dir, err := getGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "workspaces.yaml"), nil
}

func loadWorkspaceConfig() (*WorkspaceConfig, error) {
	path, err := getWorkspacesConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &WorkspaceConfig{}, nil
		}
		return nil, fmt.Errorf("reading workspace config: %w", err)
	}

	var cfg WorkspaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing workspace config: %w", err)
	}
	return &cfg, nil
}

func saveWorkspaceConfig(cfg *WorkspaceConfig) error {
	path, err := getWorkspacesConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling workspace config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing workspace config: %w", err)
	}
	return nil
}

func workspaceExists(cfg *WorkspaceConfig, name string) bool {
	for _, w := range cfg.Workspaces {
		if w.Name == name {
			return true
		}
	}
	return false
}

func findWorkspace(cfg *WorkspaceConfig, name string) *WorkspaceEntry {
	for i := range cfg.Workspaces {
		if cfg.Workspaces[i].Name == name {
			return &cfg.Workspaces[i]
		}
	}
	return nil
}

func defaultNameFromPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	base := filepath.Base(abs)
	if base == "" || base == "." || base == "/" {
		return "workspace"
	}
	return strings.ReplaceAll(strings.ReplaceAll(base, " ", "-"), "_", "-")
}

func hasLazyAIArtifacts(dir string) bool {
	markers := []string{
		".ai-setup.db",
		".ai-setup.json",
		".specify",
		".opencode",
		"AGENTS.md",
	}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

func runWorkspaceList(cmd *cobra.Command, args []string) error {
	cfg, err := loadWorkspaceConfig()
	if err != nil {
		return err
	}

	if len(cfg.Workspaces) == 0 {
		fmt.Println("No workspaces registered.")
		fmt.Println("  Use 'lazyai-cli workspace add <path>' to register one.")
		return nil
	}

	fmt.Println("Workspaces:")
	fmt.Println(strings.Repeat("-", 60))
	for _, w := range cfg.Workspaces {
		exists := "✓"
		if _, err := os.Stat(w.Path); os.IsNotExist(err) {
			exists = "✗"
		}
		active := ""
		if w.Name == cfg.Active {
			active = " ← active"
		}
		fmt.Printf("  %s %-20s %s%s\n", exists, w.Name, w.Path, active)
	}
	return nil
}

func runWorkspaceAdd(cmd *cobra.Command, args []string) error {
	path := args[0]

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		name = defaultNameFromPath(absPath)
	}
	if err := ValidateNotEmpty(name, "name"); err != nil {
		return err
	}

	cfg, err := loadWorkspaceConfig()
	if err != nil {
		return err
	}

	if workspaceExists(cfg, name) {
		return fmt.Errorf("workspace '%s' already exists", name)
	}

	cfg.Workspaces = append(cfg.Workspaces, WorkspaceEntry{
		Name: name,
		Path: absPath,
	})

	// If this is the first workspace, auto-activate it.
	if cfg.Active == "" {
		cfg.Active = name
	}

	if err := saveWorkspaceConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✅ Workspace added: %s → %s\n", name, absPath)
	if cfg.Active == name {
		fmt.Printf("   (set as active)\n")
	}

	// Best-effort ledger append — do not fail command if ledger is absent.
	_ = appendToLedger("workspace_add", map[string]string{
		"name": name,
		"path": absPath,
	})

	return nil
}

func runWorkspaceSwitch(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := loadWorkspaceConfig()
	if err != nil {
		return err
	}

	w := findWorkspace(cfg, name)
	if w == nil {
		return fmt.Errorf("workspace '%s' not found. Run 'workspace list' to see registered workspaces", name)
	}

	cfg.Active = name
	if err := saveWorkspaceConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✅ Active workspace set to: %s\n", name)
	fmt.Printf("   Path: %s\n", w.Path)

	// Best-effort ledger append.
	_ = appendToLedger("workspace_switch", map[string]string{
		"name": name,
		"path": w.Path,
	})

	return nil
}

func runWorkspaceStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadWorkspaceConfig()
	if err != nil {
		return err
	}

	if cfg.Active == "" {
		fmt.Println("No active workspace.")
		fmt.Printf("  Registered workspaces: %d\n", len(cfg.Workspaces))
		if len(cfg.Workspaces) > 0 {
			fmt.Println("  Use 'lazyai-cli workspace switch <name>' to activate one.")
		} else {
			fmt.Println("  Use 'lazyai-cli workspace add <path>' to register one.")
		}
		return nil
	}

	w := findWorkspace(cfg, cfg.Active)
	if w == nil {
		fmt.Printf("Active workspace '%s' not found in registry.\n", cfg.Active)
		return nil
	}

	exists := false
	if _, err := os.Stat(w.Path); err == nil {
		exists = true
	}

	hasArtifacts := false
	if exists {
		hasArtifacts = hasLazyAIArtifacts(w.Path)
	}

	fmt.Println("Workspace Status:")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  Name:   %s\n", w.Name)
	fmt.Printf("  Path:   %s\n", w.Path)
	if exists {
		fmt.Printf("  Exists: ✓ yes\n")
	} else {
		fmt.Printf("  Exists: ✗ no (path missing)\n")
	}
	if hasArtifacts {
		fmt.Printf("  LazyAI: ✓ artifacts present\n")
	} else {
		fmt.Printf("  LazyAI: ✗ no artifacts detected\n")
	}
	fmt.Printf("  Total registered: %d\n", len(cfg.Workspaces))

	return nil
}
