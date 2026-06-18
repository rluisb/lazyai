package wizard

import (
	"os"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/globalpaths"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// InstallPlan describes what files to install, directories to create,
// and conflicts to resolve.
type InstallPlan struct {
	FilesToInstall []PlannedFile
	DirsToCreate   []string
	Conflicts      []ConflictInfo
}

// PlannedFile describes a single file to be installed.
type PlannedFile struct {
	Source    string // library path (relative to library dir)
	Target    string // destination path (absolute or relative to target dir)
	Type      string // category: agent, skill, prompt, template, rule, infra, root, constitution
	Content   []byte // new content from the library
	Existing  bool   // true if file already exists at target
	HashMatch bool   // true if existing file hash matches library hash (no conflict)
}

// ConflictInfo describes a file conflict between current and new content.
type ConflictInfo struct {
	Target          string // path to the conflicting file
	ExistingContent []byte // content of the existing file
	Content         []byte // new content from the library
	Type            string // file category
}

// ComputePlan computes what files to install and what conflicts exist
// based on the wizard configuration.
//
// Currently this produces a placeholder plan. Full implementation requires
// the scaffold/registry packages to list available templates. The planner
// will be extended once those packages are ported.
func ComputePlan(config *WizardConfig) (*InstallPlan, error) {
	plan := &InstallPlan{}

	// Derive target dir from scope.
	targetDir := config.TargetDir
	if config.CLIScope == types.SetupScopeGlobal && config.HomeDir != "" {
		globalPath, err := globalpaths.ResolveGlobalToolTargetDir(types.ToolIdOpenCode, config.HomeDir)
		if err == nil {
			targetDir = globalPath
		} else {
			targetDir = filepath.Join(config.HomeDir, ".config", "opencode")
		}
	}

	// Build file list based on tools and preset.
	// Determine preset and features.
	preset := types.PresetLevelStandard
	// This will be populated from Phase 2 results.

	_ = targetDir
	_ = preset

	// TODO: Implement full plan computation once scaffold packages are ported.
	// For now, return an empty plan with no conflicts.
	return plan, nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
