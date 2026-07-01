package sidecar

import (
	"os"
	"path/filepath"
)

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
		cfg, err := LoadSidecarAt(globalRoot)
		if err != nil {
			return nil, err
		}
		globalConfig = cfg
	}

	// --- Project layer: cwd itself, checked first, independent of the walk ---
	projectConfig, err := LoadSidecarAt(cwd)
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

		cfg, err := LoadSidecarAt(current)
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
