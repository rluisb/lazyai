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
	// 1. Load global config dir before applying configured sidecars.
	globalDir, err := getGlobalConfigDir()
	if err != nil {
		return ResolvedPaths{}, err
	}

	// 2. Determine the default (fallback) root for this scope.
	defaultRoot, err := scopeDefaultRoot(scope, projectRoot, globalDir)
	if err != nil {
		return ResolvedPaths{}, err
	}

	result := ResolvedPaths{
		DocsDir:     filepath.Join(defaultRoot, "docs"),
		SpecsDir:    filepath.Join(defaultRoot, "specs"),
		PlansDir:    filepath.Join(defaultRoot, "plans"),
		ConfigLevel: "default",
	}

	apply := func(cfg *SidecarConfig, level, anchor string) error {
		if cfg == nil {
			return nil
		}

		resolved, err := resolveSidecar(cfg, anchor)
		if err != nil {
			return err
		}
		result = resolved
		result.ConfigLevel = level
		return nil
	}

	// 3. Load and apply sidecars in scope-specific chain order.
	switch scope {
	case ScopeGlobal:
		globalCfg, err := LoadGlobalSidecar()
		if err != nil {
			return ResolvedPaths{}, err
		}
		if err := apply(globalCfg, "global", globalDir); err != nil {
			return ResolvedPaths{}, err
		}
	case ScopeProject:
		globalCfg, err := LoadGlobalSidecar()
		if err != nil {
			return ResolvedPaths{}, err
		}
		projectCfg, err := LoadProjectSidecar(projectRoot)
		if err != nil {
			return ResolvedPaths{}, err
		}
		if err := apply(globalCfg, "global", globalDir); err != nil {
			return ResolvedPaths{}, err
		}
		if err := apply(projectCfg, "project", projectRoot); err != nil {
			return ResolvedPaths{}, err
		}
	case ScopeWorkspace:
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
		if err := apply(globalCfg, "global", globalDir); err != nil {
			return ResolvedPaths{}, err
		}
		if err := apply(projectCfg, "project", projectRoot); err != nil {
			return ResolvedPaths{}, err
		}
		if err := apply(workspaceCfg, "workspace", globalDir); err != nil {
			return ResolvedPaths{}, err
		}
	default:
		return ResolvedPaths{}, fmt.Errorf("unknown scope: %v", scope)
	}

	return result, nil
}

// scopeDefaultRoot returns the fallback directory when no sidecar is configured.
func scopeDefaultRoot(scope Scope, projectRoot, globalDir string) (string, error) {
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
		return globalDir, nil
	default:
		return "", fmt.Errorf("unknown scope: %v", scope)
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
