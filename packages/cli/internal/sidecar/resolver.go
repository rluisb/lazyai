package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve computes the final docs/specs/plans paths for the given scope.
// Resolution priority: workspace > project > global > default.
// When no sidecar is configured at any level, it falls back to the scope root.
func Resolve(scope Scope, projectRoot string) (ResolvedPaths, error) {
	// 1. Determine the default (fallback) root for this scope.
	defaultRoot, err := scopeDefaultRoot(scope, projectRoot)
	if err != nil {
		return ResolvedPaths{}, err
	}

	result := ResolvedPaths{
		DocsDir:     filepath.Join(defaultRoot, "docs"),
		SpecsDir:    filepath.Join(defaultRoot, "specs"),
		PlansDir:    filepath.Join(defaultRoot, "plans"),
		ConfigLevel: "default",
	}

	// 2. Load sidecars in priority order and merge.
	globalCfg, err := LoadGlobalSidecar()
	if err != nil {
		return ResolvedPaths{}, err
	}

	projectCfg, err := LoadProjectSidecar(projectRoot)
	if err != nil {
		return ResolvedPaths{}, err
	}

	workspaceCfg, err := LoadWorkspaceSidecar()
	if err != nil {
		return ResolvedPaths{}, err
	}

	// Apply global if present.
	if globalCfg != nil {
		resolved, err := resolveSidecar(globalCfg, scopeAnchor(scope, projectRoot, ""))
		if err != nil {
			return ResolvedPaths{}, err
		}
		result = resolved
		result.ConfigLevel = "global"
	}

	// Apply project if present.
	if projectCfg != nil {
		resolved, err := resolveSidecar(projectCfg, scopeAnchor(scope, projectRoot, projectRoot))
		if err != nil {
			return ResolvedPaths{}, err
		}
		result = resolved
		result.ConfigLevel = "project"
	}

	// Apply workspace if present.
	if workspaceCfg != nil {
		globalDir, err := getGlobalConfigDir()
		if err != nil {
			return ResolvedPaths{}, err
		}
		resolved, err := resolveSidecar(workspaceCfg, scopeAnchor(scope, projectRoot, globalDir))
		if err != nil {
			return ResolvedPaths{}, err
		}
		result = resolved
		result.ConfigLevel = "workspace"
	}

	return result, nil
}

// scopeDefaultRoot returns the fallback directory when no sidecar is configured.
func scopeDefaultRoot(scope Scope, projectRoot string) (string, error) {
	switch scope {
	case ScopeWorkspace:
		cfg, err := LoadWorkspaceConfig()
		if err != nil {
			return "", err
		}
		if cfg.Active == "" {
			return "", fmt.Errorf("no active workspace set")
		}
		for _, w := range cfg.Workspaces {
			if w.Name == cfg.Active {
				return w.Path, nil
			}
		}
		return "", fmt.Errorf("active workspace %q not found", cfg.Active)
	case ScopeProject:
		return projectRoot, nil
	case ScopeGlobal:
		dir, err := getGlobalConfigDir()
		if err != nil {
			return "", err
		}
		return dir, nil
	default:
		return "", fmt.Errorf("unknown scope: %v", scope)
	}
}

// scopeAnchor returns the anchor directory for resolving relative sidecar.path values.
func scopeAnchor(scope Scope, projectRoot string, fallback string) string {
	switch scope {
	case ScopeWorkspace, ScopeGlobal:
		dir, err := getGlobalConfigDir()
		if err != nil {
			return fallback
		}
		return dir
	case ScopeProject:
		return projectRoot
	default:
		return fallback
	}
}

// resolveSidecar turns a SidecarConfig into ResolvedPaths.
// Relative paths in *_dir are resolved against the sidecar path.
func resolveSidecar(cfg *SidecarConfig, anchor string) (ResolvedPaths, error) {
	if cfg == nil {
		return ResolvedPaths{}, nil
	}

	sidecarPath := cfg.Path
	if sidecarPath == "" {
		// If path is empty, use the anchor as the sidecar root.
		sidecarPath = anchor
	} else if !filepath.IsAbs(sidecarPath) {
		sidecarPath = filepath.Join(anchor, sidecarPath)
	}

	// Clean the path.
	sidecarPath = filepath.Clean(sidecarPath)

	specsDir := defaultIfEmpty(cfg.SpecsDir, "specs")
	docsDir := defaultIfEmpty(cfg.DocsDir, "docs")
	plansDir := defaultIfEmpty(cfg.PlansDir, "plans")

	return ResolvedPaths{
		SpecsDir:    resolveDir(specsDir, sidecarPath),
		DocsDir:     resolveDir(docsDir, sidecarPath),
		PlansDir:    resolveDir(plansDir, sidecarPath),
		ConfigLevel: "", // filled by caller
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
