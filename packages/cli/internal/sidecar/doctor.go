package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
)

type sidecarCandidate struct {
	level  string
	cfg    *SidecarConfig
	anchor string
}

// Doctor validates sidecar configuration for the given scope.
// It returns a slice of issues; an empty slice means everything is valid.
// Missing sidecar at any level is not an error — it results in an empty slice.
func Doctor(scope Scope, projectRoot string) ([]Issue, error) {
	var issues []Issue

	globalDir, err := getGlobalConfigDir()
	if err != nil {
		return nil, err
	}

	appendCandidate := func(level string, cfg *SidecarConfig, anchor string) {
		issues = append(issues, validateSidecarCandidate(sidecarCandidate{
			level:  level,
			cfg:    cfg,
			anchor: anchor,
		})...)
	}

	switch scope {
	case ScopeGlobal:
		globalCfg, err := LoadGlobalSidecar()
		if err != nil {
			return nil, err
		}
		appendCandidate("global", globalCfg, globalDir)
	case ScopeProject:
		globalCfg, err := LoadGlobalSidecar()
		if err != nil {
			return nil, err
		}
		projectCfg, err := LoadProjectSidecar(projectRoot)
		if err != nil {
			return nil, err
		}
		appendCandidate("global", globalCfg, globalDir)
		appendCandidate("project", projectCfg, projectRoot)
	case ScopeWorkspace:
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
		appendCandidate("global", globalCfg, globalDir)
		appendCandidate("project", projectCfg, projectRoot)
		appendCandidate("workspace", workspaceCfg, globalDir)
	default:
		return nil, fmt.Errorf("unknown scope: %v", scope)
	}

	return issues, nil
}

func validateSidecarCandidate(candidate sidecarCandidate) []Issue {
	var issues []Issue

	if candidate.cfg == nil {
		return issues
	}

	sidecarPath := candidate.cfg.Path
	if sidecarPath == "" {
		return append(issues, Issue{
			Severity: IssueSeverityError,
			Message:  fmt.Sprintf("sidecar %s: path is required but empty", candidate.level),
			Path:     "",
		})
	}

	if !filepath.IsAbs(sidecarPath) {
		sidecarPath = filepath.Join(candidate.anchor, sidecarPath)
	}
	sidecarPath = filepath.Clean(sidecarPath)

	info, err := os.Stat(sidecarPath)
	if err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, Issue{
				Severity: IssueSeverityWarning,
				Message:  fmt.Sprintf("sidecar %s: path does not exist: %s", candidate.level, sidecarPath),
				Path:     sidecarPath,
			})
		} else {
			issues = append(issues, Issue{
				Severity: IssueSeverityError,
				Message:  fmt.Sprintf("sidecar %s: cannot stat path: %s", candidate.level, sidecarPath),
				Path:     sidecarPath,
			})
		}
	} else if !info.IsDir() {
		issues = append(issues, Issue{
			Severity: IssueSeverityError,
			Message:  fmt.Sprintf("sidecar %s: path is a file, not a directory: %s", candidate.level, sidecarPath),
			Path:     sidecarPath,
		})
	} else if !IsWritable(sidecarPath) {
		issues = append(issues, Issue{
			Severity: IssueSeverityError,
			Message:  fmt.Sprintf("sidecar %s: path is not writable: %s", candidate.level, sidecarPath),
			Path:     sidecarPath,
		})
	}

	// Check each *_dir resolves to an existing or creatable path.
	dirs := map[string]string{
		"specs_dir": candidate.cfg.SpecsDir,
		"docs_dir":  candidate.cfg.DocsDir,
		"plans_dir": candidate.cfg.PlansDir,
	}
	for field, dir := range dirs {
		resolved := resolveDir(defaultIfEmpty(dir, field[:len(field)-4]), sidecarPath)
		if _, err := os.Stat(resolved); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, Issue{
					Severity: IssueSeverityWarning,
					Message:  fmt.Sprintf("sidecar %s: %s directory does not exist (will be created on first write): %s", candidate.level, field, resolved),
					Path:     resolved,
				})
			} else {
				issues = append(issues, Issue{
					Severity: IssueSeverityError,
					Message:  fmt.Sprintf("sidecar %s: cannot stat %s: %s", candidate.level, field, resolved),
					Path:     resolved,
				})
			}
		}
	}

	return issues
}
