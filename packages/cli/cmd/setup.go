package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/setupscan"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Inspect setup inventory and planning output",
	Long:  "Inspect setup inventory and setup planning output without broadening into runtime execution or daemon management.",
	RunE:  runSetup,
}

func init() {
	setupCmd.Flags().Bool("scan", false, "Scan known tool targets and print inventory JSON")
	setupCmd.Flags().Bool("list", false, "List supported setup targets and reusable setup resources")
	setupCmd.Flags().Bool("dry-run", false, "Show the setup plan without writing files")
	setupCmd.Flags().Bool("adopt", false, "Mark adoptable external configs as LazyAI managed")
	setupCmd.Flags().Bool("import", false, "Import external configs into LazyAI reference storage")
	setupCmd.Flags().StringSlice("tool", nil, "Limit setup planning to specific tools (repeatable)")
	setupCmd.Flags().Bool("all", false, "Select all supported setup targets for the requested scope")
	setupCmd.Flags().Bool("global", false, "Use global scope/home layout where supported")
	rootCmd.AddCommand(setupCmd)
	setupCmd.GroupID = "workspace"
}

type setupListResult struct {
	Mode        string                    `json:"mode"`
	ScopeFilter string                    `json:"scopeFilter,omitempty"`
	SharedPaths []setupscan.DesiredPath   `json:"sharedPaths"`
	Targets     []setupscan.DesiredTarget `json:"targets"`
	Agents      []setupscan.ObservedAgent `json:"agents,omitempty"`
}

type setupDryRunResult struct {
	Mode        string                  `json:"mode"`
	Scope       string                  `json:"scope"`
	SharedPaths []setupscan.DesiredPath `json:"sharedPaths"`
	Targets     []setupDryRunTarget     `json:"targets"`
}

type setupDryRunTarget struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Scope          string   `json:"scope"`
	Origin         string   `json:"origin"`
	RootPath       string   `json:"rootPath"`
	ExpectedFiles  []string `json:"expectedFiles"`
	ObservedFiles  []string `json:"observedFiles,omitempty"`
	ExistingStatus string   `json:"existingStatus"`
	ExistingState  string   `json:"existingState,omitempty"`
	Action         string   `json:"action"`
}

func runSetup(cmd *cobra.Command, args []string) error {
	scan, _ := cmd.Flags().GetBool("scan")
	list, _ := cmd.Flags().GetBool("list")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	adopt, _ := cmd.Flags().GetBool("adopt")
	importConfigs, _ := cmd.Flags().GetBool("import")
	selectedTools, _ := cmd.Flags().GetStringSlice("tool")
	selectAll, _ := cmd.Flags().GetBool("all")
	globalScope, _ := cmd.Flags().GetBool("global")

	actionCount := 0
	for _, enabled := range []bool{scan, list, dryRun} {
		if enabled {
			actionCount++
		}
	}
	if actionCount == 0 {
		if adopt || importConfigs {
			return fmt.Errorf("--adopt and --import require --scan")
		}
		return fmt.Errorf("no setup action selected (try: lazyai-cli setup --scan, --list, or --dry-run)")
	}
	if actionCount > 1 {
		return fmt.Errorf("select exactly one of --scan, --list, or --dry-run")
	}
	if (adopt || importConfigs) && !scan {
		return fmt.Errorf("--adopt and --import require --scan")
	}
	if scan && (len(selectedTools) > 0 || selectAll || globalScope) {
		return fmt.Errorf("--tool, --all, and --global are only supported with --list or --dry-run")
	}
	if selectAll && len(selectedTools) > 0 {
		return fmt.Errorf("--all cannot be combined with --tool")
	}

	targetDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("determine working directory: %w", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}
	baseOpts := setupscan.Options{HomeDir: homeDir, TargetDir: targetDir}

	if list || dryRun {
		selection, err := resolveSetupSelection(selectedTools)
		if err != nil {
			return err
		}
		scope := types.SetupScopeProject
		if globalScope {
			scope = types.SetupScopeGlobal
		}
		inventory, err := setupscan.Scan(baseOpts)
		if err != nil {
			return err
		}
		if list {
			result, err := buildSetupListResult(inventory, selection, globalScope, scope)
			if err != nil {
				return err
			}
			return writeSetupJSON(result)
		}
		result, err := buildSetupDryRunResult(inventory, selection, selectAll, scope)
		if err != nil {
			return err
		}
		return writeSetupJSON(result)
	}

	inventory, err := setupscan.Run(setupscan.Options{HomeDir: homeDir, TargetDir: targetDir, Adopt: adopt, Import: importConfigs})
	if err != nil {
		return err
	}
	return writeSetupJSON(inventory)
}

type setupToolSelection struct {
	requested map[string]bool
	explicit  bool
}

func resolveSetupSelection(names []string) (setupToolSelection, error) {
	selection := setupToolSelection{requested: map[string]bool{}, explicit: len(names) > 0}
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		toolID := types.ToolId(trimmed)
		if !types.IsValidToolId(toolID) {
			return setupToolSelection{}, fmt.Errorf("unknown tool %q", trimmed)
		}
		selection.requested[string(toolID)] = true
	}
	return selection, nil
}

func buildSetupListResult(inventory *setupscan.Inventory, selection setupToolSelection, filterScope bool, scope types.SetupScope) (*setupListResult, error) {
	targets, err := filterDesiredTargets(inventory.DesiredState.Targets, selection, filterScope, scope)
	if err != nil {
		return nil, err
	}
	result := &setupListResult{
		Mode:        "list",
		SharedPaths: filterDesiredSharedPaths(inventory.DesiredState.SharedPaths, filterScope, scope),
		Targets:     targets,
		Agents:      append([]setupscan.ObservedAgent(nil), inventory.CurrentState.Agents...),
	}
	if filterScope {
		result.ScopeFilter = string(scope)
	}
	return result, nil
}

func buildSetupDryRunResult(inventory *setupscan.Inventory, selection setupToolSelection, _ bool, scope types.SetupScope) (*setupDryRunResult, error) {
	targets, err := filterDesiredTargets(inventory.DesiredState.Targets, selection, true, scope)
	if err != nil {
		return nil, err
	}
	planned := make([]setupDryRunTarget, 0, len(targets))
	for _, target := range targets {
		if len(target.CandidateRoots) == 0 {
			continue
		}
		root := target.CandidateRoots[0]
		detection, ok := findDetection(inventory.CurrentState.Targets, target.ID, root.Scope)
		item := setupDryRunTarget{
			ID:             target.ID,
			Name:           target.Name,
			Scope:          root.Scope,
			Origin:         root.Origin,
			RootPath:       root.RootPath,
			ExpectedFiles:  append([]string(nil), root.ExpectedFiles...),
			Action:         "initialize",
			ExistingStatus: "missing",
		}
		if ok {
			item.ObservedFiles = append([]string(nil), detection.ObservedFiles...)
			item.ExistingStatus = detection.Status
			item.ExistingState = detection.State
			if detection.Status == "detected" {
				item.Action = "preserve-existing"
			}
		}
		planned = append(planned, item)
	}
	return &setupDryRunResult{
		Mode:        "dry-run",
		Scope:       string(scope),
		SharedPaths: filterDesiredSharedPaths(inventory.DesiredState.SharedPaths, true, scope),
		Targets:     planned,
	}, nil
}

func filterDesiredTargets(targets []setupscan.DesiredTarget, selection setupToolSelection, filterScope bool, scope types.SetupScope) ([]setupscan.DesiredTarget, error) {
	filtered := make([]setupscan.DesiredTarget, 0, len(targets))
	seen := map[string]bool{}
	for _, target := range targets {
		if selection.explicit && !selection.requested[target.ID] {
			continue
		}
		if filterScope && !adapter.IsScopeSupported(types.ToolId(target.ID), scope) {
			continue
		}
		next := target
		if filterScope {
			next.SupportedScopes = []string{string(scope)}
			candidateRoots := make([]setupscan.DesiredRoot, 0, len(target.CandidateRoots))
			for _, root := range target.CandidateRoots {
				if root.Scope == string(scope) {
					candidateRoots = append(candidateRoots, root)
				}
			}
			next.CandidateRoots = candidateRoots
		}
		filtered = append(filtered, next)
		seen[target.ID] = true
	}
	if filterScope && selection.explicit {
		for targetID := range selection.requested {
			if !seen[targetID] {
				return nil, fmt.Errorf("tool %q does not support scope %q", targetID, scope)
			}
		}
	}
	return filtered, nil
}

func filterDesiredSharedPaths(paths []setupscan.DesiredPath, filterScope bool, scope types.SetupScope) []setupscan.DesiredPath {
	if !filterScope {
		return append([]setupscan.DesiredPath(nil), paths...)
	}
	prefix := "project-"
	if scope == types.SetupScopeGlobal {
		prefix = "global-"
	}
	filtered := make([]setupscan.DesiredPath, 0, 1)
	for _, path := range paths {
		if strings.HasPrefix(path.ID, prefix) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func findDetection(targets []setupscan.ObservedTarget, targetID, scope string) (setupscan.TargetDetection, bool) {
	for _, target := range targets {
		if target.ID != targetID {
			continue
		}
		for _, detection := range target.Detections {
			if detection.Scope == scope {
				return detection, true
			}
		}
	}
	return setupscan.TargetDetection{}, false
}

func writeSetupJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
