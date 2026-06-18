package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
)

// Doctor validates sidecar configuration for the given scope.
// It returns a slice of issues; an empty slice means everything is valid.
// Missing sidecar at any level is not an error — it results in an empty slice.
func Doctor(scope Scope, projectRoot string) ([]Issue, error) {
	var issues []Issue

	// Load sidecars in priority order so we can report which level has issues.
	globalCfg, err := LoadGlobalSidecar()
	if err != nil {
		return nil, err
	}

	projectCfg, err := LoadProjectSidecar(projectRoot)
	if err != nil {
		return nil, err
	}

	workspaceCfg, err := LoadWorkspaceSidecar()
	if err != nil {
		return nil, err
	}

	// Determine which config is active (highest priority present).
	activeCfg := globalCfg
	activeLevel := "global"
	if projectCfg != nil {
		activeCfg = projectCfg
		activeLevel = "project"
	}
	if workspaceCfg != nil {
		activeCfg = workspaceCfg
		activeLevel = "workspace"
	}

	if activeCfg == nil {
		return issues, nil
	}

	anchor := scopeAnchor(scope, projectRoot, "")
	if scope == ScopeWorkspace || scope == ScopeGlobal {
		globalDir, err := getGlobalConfigDir()
		if err == nil {
			anchor = globalDir
		}
	}

	sidecarPath := activeCfg.Path
	if sidecarPath == "" {
		issues = append(issues, Issue{
			Severity: IssueSeverityError,
			Message:  fmt.Sprintf("sidecar %s: path is required but empty", activeLevel),
			Path:     "",
		})
		return issues, nil
	}

	if !filepath.IsAbs(sidecarPath) {
		sidecarPath = filepath.Join(anchor, sidecarPath)
	}
	sidecarPath = filepath.Clean(sidecarPath)

	info, err := os.Stat(sidecarPath)
	if err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, Issue{
				Severity: IssueSeverityWarning,
				Message:  fmt.Sprintf("sidecar %s: path does not exist: %s", activeLevel, sidecarPath),
				Path:     sidecarPath,
			})
		} else {
			issues = append(issues, Issue{
				Severity: IssueSeverityError,
				Message:  fmt.Sprintf("sidecar %s: cannot stat path: %s", activeLevel, sidecarPath),
				Path:     sidecarPath,
			})
		}
	} else if !info.IsDir() {
		issues = append(issues, Issue{
			Severity: IssueSeverityError,
			Message:  fmt.Sprintf("sidecar %s: path is a file, not a directory: %s", activeLevel, sidecarPath),
			Path:     sidecarPath,
		})
	} else if !IsWritable(sidecarPath) {
		issues = append(issues, Issue{
			Severity: IssueSeverityError,
			Message:  fmt.Sprintf("sidecar %s: path is not writable: %s", activeLevel, sidecarPath),
			Path:     sidecarPath,
		})
	}

	// Check each *_dir resolves to an existing or creatable path.
	dirs := map[string]string{
		"specs_dir": activeCfg.SpecsDir,
		"docs_dir":  activeCfg.DocsDir,
		"plans_dir": activeCfg.PlansDir,
	}
	for field, dir := range dirs {
		resolved := resolveDir(defaultIfEmpty(dir, field[:len(field)-4]), sidecarPath)
		if _, err := os.Stat(resolved); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, Issue{
					Severity: IssueSeverityWarning,
					Message:  fmt.Sprintf("sidecar %s: %s directory does not exist (will be created on first write): %s", activeLevel, field, resolved),
					Path:     resolved,
				})
			} else {
				issues = append(issues, Issue{
					Severity: IssueSeverityError,
					Message:  fmt.Sprintf("sidecar %s: cannot stat %s: %s", activeLevel, field, resolved),
					Path:     resolved,
				})
			}
		}
	}

	// Check linked_projects.
	for _, lp := range activeCfg.LinkedProjects {
		if _, err := os.Stat(lp.Path); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, Issue{
					Severity: IssueSeverityWarning,
					Message:  fmt.Sprintf("sidecar %s: linked project %q path does not exist: %s", activeLevel, lp.Name, lp.Path),
					Path:     lp.Path,
				})
			} else {
				issues = append(issues, Issue{
					Severity: IssueSeverityError,
					Message:  fmt.Sprintf("sidecar %s: linked project %q cannot be accessed: %s", activeLevel, lp.Name, lp.Path),
					Path:     lp.Path,
				})
			}
		}
	}

	return issues, nil
}
