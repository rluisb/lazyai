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

// Doctor validates every discovered .lazyai/ layer (global, workspace if
// found, project if found) relative to cwd, plus checks for a stale
// ~/.lazyai/workspaces.yaml (migration hint, see
// specs/refactors/579-lazyai-config-directory/spec.md §7). There is no
// caller-supplied Scope — Doctor mirrors Resolve's full-discovery contract.
// It returns a slice of issues; an empty slice means everything is valid.
// Missing sidecar at any level is not an error — it results in no issues
// for that level.
func Doctor(cwd string) ([]Issue, error) {
	layers, err := DiscoverLayers(cwd)
	if err != nil {
		return nil, err
	}

	var issues []Issue

	appendCandidate := func(layer Layer) {
		for _, issue := range validateSidecarCandidate(sidecarCandidate{
			level:  layer.Level,
			cfg:    layer.Config,
			anchor: layer.Root,
		}) {
			issue.Level = layer.Level
			issues = append(issues, issue)
		}
	}

	appendCandidate(layers.Global)
	if layers.Workspace != nil {
		appendCandidate(*layers.Workspace)
	}
	if layers.Project != nil {
		appendCandidate(*layers.Project)
	}

	if home, err := os.UserHomeDir(); err == nil {
		if _, err := os.Stat(filepath.Join(home, ".lazyai", "workspaces.yaml")); err == nil {
			issues = append(issues, Issue{
				Severity: IssueSeverityWarning,
				Level:    "",
				Message:  "found legacy ~/.lazyai/workspaces.yaml (workspace registry removed in #579) — re-create any workspace/project sidecars you still need with 'sidecar init' from the intended directory",
			})
		}
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
