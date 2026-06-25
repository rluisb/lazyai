package adapter

import (
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/configmerge"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// writePiSettings emits the canonical Pi settings.json at the scope-appropriate
// path and records it as a tracked file.
//
// Settings contract (issue #532):
//
//   - Project scope: .pi/settings.json  (paths resolve relative to .pi)
//   - Global  scope: ~/.pi/agent/settings.json (paths resolve relative to ~/.pi/agent)
//   - Workspace scope: .pi/settings.json under the workspace root
//
// Pi merges nested objects across scopes (global → project), and project values
// override global on a per-key basis. LazyAI uses configmerge.MergeJSONFile so
// user-authored keys survive re-runs and idempotent re-runs produce identical
// bytes. The patch below declares only the resource references LazyAI manages;
// sibling issues (#537 packages, #535 themes, #533 extension layout) extend
// this patch map rather than introducing a second write path.
//
// Resource arrays point at the on-disk subdirectories the Install step already
// populated so Pi discovers the LazyAI-installed extensions, skills, and
// prompts from settings.json.
func writePiSettings(ctx *AdapterContext, piDir string) error {
	settingsPath := piSettingsPath(piDir, ctx.SetupScope)
	if err := files.EnsureDir(filepath.Dir(settingsPath)); err != nil {
		return err
	}

	patch := defaultPiSettingsPatch()

	if _, err := configmerge.MergeJSONFile(settingsPath, patch); err != nil {
		return err
	}
	return trackFile(ctx, settingsPath, "pi/settings.json")
}

// piSettingsPath returns the on-disk settings.json path for the given Pi root
// and scope. Global scope writes under the `agent` subdirectory per the Pi
// settings documentation; project/workspace scopes write alongside the other
// .pi resources.
func piSettingsPath(piDir string, scope types.SetupScope) string {
	if scope == types.SetupScopeGlobal {
		return filepath.Join(piDir, "agent", "settings.json")
	}
	return filepath.Join(piDir, "settings.json")
}

// defaultPiSettingsPatch returns the LazyAI-managed settings keys for Pi.
// Only resource references are declared by default; model/theme/compaction
// settings stay user-owned. Sibling issues extend this map:
//   - #537 adds package configuration entries.
//   - #535 adds theme references if theme support lands.
//   - #533 may add/adjust extension references.
func defaultPiSettingsPatch() map[string]any {
	return map[string]any{
		// Point Pi at the LazyAI-installed resource directories. Paths in
		// .pi/settings.json resolve relative to .pi; paths in the global
		// ~/.pi/agent/settings.json resolve relative to ~/.pi/agent, so
		// the relative references stay identical across scopes.
		"extensions": []any{"extensions"},
		"skills":     []any{"skills"},
		"prompts":    []any{"prompts"},
	}
}
