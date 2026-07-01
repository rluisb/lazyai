// Package sidecar provides the internal LazyAI sidecar configuration domain.
// It resolves docs/specs/plans directories with an optional sidecar override.
package sidecar

// SidecarConfig holds the sidecar configuration for a single scope level.
type SidecarConfig struct {
	Path     string `yaml:"path"`
	SpecsDir string `yaml:"specs_dir,omitempty"`
	DocsDir  string `yaml:"docs_dir,omitempty"`
	PlansDir string `yaml:"plans_dir,omitempty"`
}

// ResolvedPaths holds the final resolved directory paths.
type ResolvedPaths struct {
	DocsDir    string
	SpecsDir   string
	PlansDir   string
	Provenance map[string]string // keys: "docs_dir","specs_dir","plans_dir" -> "project"|"workspace"|"global"|"default"
}

// IsAllDefault reports whether every resolved field fell through to the
// built-in default (i.e. no .lazyai/ layer anywhere set anything).
func (r ResolvedPaths) IsAllDefault() bool {
	for _, v := range r.Provenance {
		if v != "default" {
			return false
		}
	}
	return true
}

// Scope represents the resolution scope.
type Scope int

const (
	ScopeWorkspace Scope = iota
	ScopeProject
	ScopeGlobal
)

// String returns the human-readable scope name.
func (s Scope) String() string {
	switch s {
	case ScopeWorkspace:
		return "workspace"
	case ScopeProject:
		return "project"
	case ScopeGlobal:
		return "global"
	default:
		return "unknown"
	}
}

// IssueSeverity indicates the severity of a doctor issue.
type IssueSeverity string

const (
	IssueSeverityError   IssueSeverity = "ERROR"
	IssueSeverityWarning IssueSeverity = "WARN"
)

// Issue represents a single problem found by the doctor.
type Issue struct {
	Severity IssueSeverity
	Message  string
	Path     string
	Level    string // "global" | "workspace" | "project" | "" (e.g. the workspaces.yaml migration hint)
}

// SidecarFile is the on-disk shape of every <scope-root>/.lazyai/sidecar.yaml,
// at every scope. Replaces GlobalSidecarConfig + ProjectSidecarConfig (identical
// shape, previously duplicated once per scope for no reason — no scope has ever
// had a differently-shaped file).
type SidecarFile struct {
	Sidecar *SidecarConfig `yaml:"sidecar"`
}

// WorkspaceConfig holds the global workspace registry.
// Defined here to avoid a circular dependency with cmd package.
type WorkspaceConfig struct {
	Workspaces []WorkspaceEntry `yaml:"workspaces"`
	Active     string           `yaml:"active"`
}

// WorkspaceEntry represents a registered project/workspace.
type WorkspaceEntry struct {
	Name    string         `yaml:"name"`
	Path    string         `yaml:"path"`
	Sidecar *SidecarConfig `yaml:"sidecar,omitempty"`
}

// Layer is one discovered (or always-present, for Global) config source.
type Layer struct {
	Level  string         // "global" | "workspace" | "project"
	Root   string         // scope root: home dir (global), discovered ancestor dir (workspace), cwd (project)
	Config *SidecarConfig // never nil when the Layer itself is non-nil/present (Global.Config may be nil if absent)
}

// Layers is the result of one discovery pass from a given cwd.
// Global is always populated (Config may be nil if ~/.lazyai/sidecar.yaml is absent).
// Workspace and Project are nil when no layer was found at that level.
type Layers struct {
	Global    Layer
	Workspace *Layer
	Project   *Layer
}
