package sidecar

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// loadSidecarAtProbe is a Phase 1 stand-in for the Phase 2 LoadSidecarAt
// primitive (spec.md §3.2), which does not exist yet. It implements the
// identical contract LoadSidecarAt will carry: returns (nil, nil) when
// <scopeRoot>/.lazyai/sidecar.yaml does not exist as a file, and only
// returns a non-nil error on a real I/O/parse failure. DiscoverLayers is
// written to call this helper exclusively so that swapping it for the real
// LoadSidecarAt in Phase 2 is a zero-behavior-change, single-call-site edit.
func loadSidecarAtProbe(scopeRoot string) (*SidecarConfig, error) {
	path := filepath.Join(scopeRoot, ".lazyai", "sidecar.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sidecar at %s: %w", scopeRoot, err)
	}
	var file ProjectSidecarConfig
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing sidecar at %s: %w", scopeRoot, err)
	}
	return file.Sidecar, nil
}

// evalSymlinksOrRaw resolves path via filepath.EvalSymlinks, falling back to
// the raw (unresolved) path on any error (e.g. a dangling symlink or a
// permission error). Symlink resolution failures must never abort discovery
// (spec.md §1.2/§1.3).
func evalSymlinksOrRaw(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

// DiscoverLayers walks up from cwd to find the project and workspace
// .lazyai/ layers, and always includes the global layer. See
// specs/refactors/579-lazyai-config-directory/spec.md §1 for the exact
// walk-up algorithm, stop conditions, and edge cases.
//
// A "layer hit" requires <dir>/.lazyai/sidecar.yaml to exist as a file — a
// bare .lazyai/ directory with no sidecar.yaml inside is not a hit.
//
// Errors only on a real, non-IsNotExist I/O failure while probing a
// candidate directory; a missing $HOME (os.UserHomeDir() error) is not
// fatal to discovery itself.
func DiscoverLayers(cwd string) (*Layers, error) {
	home, homeErr := os.UserHomeDir()
	resolvedHome := ""
	if homeErr == nil {
		resolvedHome = evalSymlinksOrRaw(home)
	}

	// --- Global layer (always present) ---
	globalRoot := home // "" if homeErr != nil
	var globalConfig *SidecarConfig
	if homeErr == nil {
		cfg, err := loadSidecarAtProbe(globalRoot)
		if err != nil {
			return nil, err
		}
		globalConfig = cfg
	}

	// --- Project layer: cwd itself, checked first, independent of the walk ---
	projectConfig, err := loadSidecarAtProbe(cwd)
	if err != nil {
		return nil, err
	}
	var project *Layer
	if projectConfig != nil {
		project = &Layer{Level: "project", Root: cwd, Config: projectConfig}
	}

	// --- Workspace layer: walk starts strictly ABOVE cwd, never re-checks cwd ---
	var workspace *Layer
	current := filepath.Dir(cwd)

	for {
		resolvedCurrent := evalSymlinksOrRaw(current)

		// Stop BEFORE entering $HOME: if this candidate IS $HOME, do not
		// check it, do not treat it as "workspace" — it already IS the
		// global layer.
		if homeErr == nil && resolvedCurrent == resolvedHome {
			break
		}

		cfg, err := loadSidecarAtProbe(current)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			workspace = &Layer{Level: "workspace", Root: current, Config: cfg}
			break // first hit wins; do not keep climbing
		}

		// Stop AT filesystem root, but only AFTER checking it once above
		// (root is a valid, if unusual, workspace location).
		parent := filepath.Dir(current)
		if parent == current { // portable root test: "/" on POSIX, "C:\" on Windows
			break
		}
		current = parent
	}

	return &Layers{
		Global:    Layer{Level: "global", Root: globalRoot, Config: globalConfig},
		Workspace: workspace,
		Project:   project,
	}, nil
}
