package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve computes the final docs/specs/plans paths by discovering every
// .lazyai/ layer above cwd and merging them field-by-field: project >
// workspace > global > built-in default (see
// specs/refactors/579-lazyai-config-directory/spec.md §5 for the merge
// algorithm). Unlike the old Resolve(scope, projectRoot), there is no
// caller-supplied Scope — resolution always reflects everything actually
// present on disk relative to cwd.
func Resolve(cwd string) (ResolvedPaths, error) {
	layers, err := DiscoverLayers(cwd)
	if err != nil {
		return ResolvedPaths{}, err
	}
	return mergeLayers(layers, cwd)
}

// ScopeRoot maps a write-target Scope to the directory sidecar
// init/WriteSidecarAt should target: ScopeGlobal -> os.UserHomeDir();
// ScopeProject and ScopeWorkspace -> cwd (identical — "project" vs
// "workspace" is a label applied at write time, not a distinct write path;
// the same file later resolves as "project" or "workspace" depending on the
// caller's cwd at resolve time). Replaces scopeDefaultRoot.
func ScopeRoot(scope Scope, cwd string) (string, error) {
	switch scope {
	case ScopeGlobal:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine home directory: %w", err)
		}
		return home, nil
	case ScopeProject, ScopeWorkspace:
		return cwd, nil
	default:
		return "", fmt.Errorf("unknown scope: %v", scope)
	}
}

// mergeLayers applies firstNonEmpty per field across project > workspace >
// global > default, and records per-field provenance. defaultAnchor is cwd
// (used only when NO layer sets a field at all, i.e. the pre-existing
// "default" fallback behavior).
//
// Deviates from spec.md §3.3's literal no-error ResolvedPaths signature:
// this must call resolveSidecar per present layer, which validates
// SidecarConfig.Path and can legitimately error
// (TestResolve_RejectsEmptySidecarPath); propagating that error is
// required, so mergeLayers returns (ResolvedPaths, error). mergeLayers is
// unexported, so this is a pure implementation detail with no effect on
// Resolve's or any other exported contract.
func mergeLayers(layers *Layers, defaultAnchor string) (ResolvedPaths, error) {
	// Step 1: resolve each present layer's OWN config into absolute
	// directory values, anchored on that layer's own Root (spec.md §3.3
	// Refinement) — reuses resolveSidecar unchanged, once per present layer.
	var resolvedProject, resolvedWorkspace, resolvedGlobal ResolvedPaths

	if layers.Project != nil {
		r, err := resolveSidecar(layers.Project.Config, layers.Project.Root)
		if err != nil {
			return ResolvedPaths{}, err
		}
		resolvedProject = r
	}
	if layers.Workspace != nil {
		r, err := resolveSidecar(layers.Workspace.Config, layers.Workspace.Root)
		if err != nil {
			return ResolvedPaths{}, err
		}
		resolvedWorkspace = r
	}
	if layers.Global.Config != nil {
		r, err := resolveSidecar(layers.Global.Config, layers.Global.Root)
		if err != nil {
			return ResolvedPaths{}, err
		}
		resolvedGlobal = r
	}

	// Step 2: per field, the RAW cfg value (pre-default) decides whether a
	// layer "explicitly set" it — resolveSidecar already ran
	// defaultIfEmpty, so reading the raw struct field here is what makes
	// this a genuine field-level merge instead of a whole-tuple replace.
	anchor := deepestPresentRoot(layers, defaultAnchor)
	result := ResolvedPaths{Provenance: map[string]string{}}

	switch {
	case layers.Project != nil && layers.Project.Config.DocsDir != "":
		result.DocsDir = resolvedProject.DocsDir
		result.Provenance["docs_dir"] = "project"
	case layers.Workspace != nil && layers.Workspace.Config.DocsDir != "":
		result.DocsDir = resolvedWorkspace.DocsDir
		result.Provenance["docs_dir"] = "workspace"
	case layers.Global.Config != nil && layers.Global.Config.DocsDir != "":
		result.DocsDir = resolvedGlobal.DocsDir
		result.Provenance["docs_dir"] = "global"
	default:
		result.DocsDir = resolveDir("docs", anchor)
		result.Provenance["docs_dir"] = "default"
	}

	switch {
	case layers.Project != nil && layers.Project.Config.SpecsDir != "":
		result.SpecsDir = resolvedProject.SpecsDir
		result.Provenance["specs_dir"] = "project"
	case layers.Workspace != nil && layers.Workspace.Config.SpecsDir != "":
		result.SpecsDir = resolvedWorkspace.SpecsDir
		result.Provenance["specs_dir"] = "workspace"
	case layers.Global.Config != nil && layers.Global.Config.SpecsDir != "":
		result.SpecsDir = resolvedGlobal.SpecsDir
		result.Provenance["specs_dir"] = "global"
	default:
		result.SpecsDir = resolveDir("specs", anchor)
		result.Provenance["specs_dir"] = "default"
	}

	switch {
	case layers.Project != nil && layers.Project.Config.PlansDir != "":
		result.PlansDir = resolvedProject.PlansDir
		result.Provenance["plans_dir"] = "project"
	case layers.Workspace != nil && layers.Workspace.Config.PlansDir != "":
		result.PlansDir = resolvedWorkspace.PlansDir
		result.Provenance["plans_dir"] = "workspace"
	case layers.Global.Config != nil && layers.Global.Config.PlansDir != "":
		result.PlansDir = resolvedGlobal.PlansDir
		result.Provenance["plans_dir"] = "global"
	default:
		result.PlansDir = resolveDir("plans", anchor)
		result.Provenance["plans_dir"] = "default"
	}

	return result, nil
}

// deepestPresentRoot returns the Root of the most-specific present layer
// (project > workspace > global), or defaultAnchor if no layer is present
// anywhere — used only as the anchor for fields no layer explicitly sets.
func deepestPresentRoot(layers *Layers, defaultAnchor string) string {
	if layers.Project != nil {
		return layers.Project.Root
	}
	if layers.Workspace != nil {
		return layers.Workspace.Root
	}
	if layers.Global.Config != nil {
		return layers.Global.Root
	}
	return defaultAnchor
}

// resolveSidecar turns a SidecarConfig into ResolvedPaths.
// Relative paths in *_dir are resolved against the sidecar path.
func resolveSidecar(cfg *SidecarConfig, anchor string) (ResolvedPaths, error) {
	if cfg == nil {
		return ResolvedPaths{}, nil
	}

	sidecarPath := cfg.Path
	if sidecarPath == "" {
		return ResolvedPaths{}, fmt.Errorf("sidecar path is required but empty")
	}
	if !filepath.IsAbs(sidecarPath) {
		sidecarPath = filepath.Join(anchor, sidecarPath)
	}

	// Clean the path.
	sidecarPath = filepath.Clean(sidecarPath)

	specsDir := defaultIfEmpty(cfg.SpecsDir, "specs")
	docsDir := defaultIfEmpty(cfg.DocsDir, "docs")
	plansDir := defaultIfEmpty(cfg.PlansDir, "plans")

	return ResolvedPaths{
		SpecsDir: resolveDir(specsDir, sidecarPath),
		DocsDir:  resolveDir(docsDir, sidecarPath),
		PlansDir: resolveDir(plansDir, sidecarPath),
	}, nil
}

// resolveDir resolves a directory path against the sidecar root.
func resolveDir(dir, sidecarPath string) string {
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Clean(filepath.Join(sidecarPath, dir))
}

// defaultIfEmpty returns value if non-empty, otherwise fallback.
func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// IsWritable reports whether the directory at path is writable.
func IsWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}
	// Try creating a temporary file.
	f, err := os.CreateTemp(path, ".lazyai_write_test_")
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return true
}
