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
	DocsDir     string
	SpecsDir    string
	PlansDir    string
	ConfigLevel string // "workspace", "project", "global", "default"
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
}

// GlobalSidecarConfig is the top-level structure of ~/.lazyai/sidecar.yaml.
type GlobalSidecarConfig struct {
	Sidecar *SidecarConfig `yaml:"sidecar"`
}

// ProjectSidecarConfig is the top-level structure of .lazyai-sidecar.yaml.
type ProjectSidecarConfig struct {
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
