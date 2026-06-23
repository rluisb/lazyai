// Package adapter provides tool-specific adapters that know how to install
// agents, skills, and configuration files for each supported AI coding tool.
// Ported from the TypeScript adapters in src/adapters/.
package adapter

import (
	"io/fs"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// AdapterContext carries all the information an adapter needs to perform
// installation. It is the Go equivalent of the TypeScript AdapterContext.
type AdapterContext struct {
	// TargetDir is the project root directory where files will be installed.
	TargetDir string
	// SetupScope is the installation scope (global, workspace, project).
	SetupScope types.SetupScope
	// WorkspaceRoot, when set and scope is workspace, overrides TargetDir
	// for tool config paths. AI tool configs go to workspace root, specs
	// go to TargetDir (the planning repo). Empty for project/global scopes.
	WorkspaceRoot string
	// HomeDir is the user's home directory, used for global scope path resolution.
	HomeDir string
	// LibraryDir is the path to the library containing source templates.
	LibraryDir string
	// LibraryFS is the filesystem for reading library data (embedded or disk).
	LibraryFS fs.FS
	// FileRecords accumulates records of all files written during installation.
	FileRecords []types.TrackedFile
	// EnableServers lists MCP servers to enable (e.g. "memory").
	EnableServers []string
	// Force overwrites existing files without prompting.
	Force bool
	// DryRun reports what would be done without writing files.
	DryRun bool
	// DriveCLI, when true, asks adapters that support it to delegate
	// scaffolding to the tool's own CLI instead of direct-write.
	// Falls back silently to direct-write when the binary is
	// absent or the CLI call fails.
	DriveCLI bool
	// LocalSecrets, when true, routes Claude Code MCP/settings writes to the
	// gitignored `.claude/settings.local.json` instead of committed surfaces
	// (`.mcp.json` / `.claude/settings.json`). Opt-in; default false.
	LocalSecrets bool
	// Strategy controls how file conflicts are handled.
	Strategy types.ConflictStrategy
	// PerFileOverrides allows per-file conflict strategy overrides.
	PerFileOverrides map[string]types.ConflictStrategy
	// Selections controls which agents, skills, and prompts to install.
	Selections AdapterSelections
	// ConfiguredProviders is the set of provider IDs (e.g., "openai",
	// "ollama-cloud") the user has authenticated for OpenCode-style
	// multi-provider tools. Populated by the wizard from auth.DetectAll
	// results. Adapters that don't honour ConfiguredProviders leave it
	// untouched.
	ConfiguredProviders []string
}

// AdapterSelections holds the user's selections for which items to install.
type AdapterSelections struct {
	Agents           []types.AgentId
	Skills           []types.SkillId
	Prompts          []types.PromptId
	ChatModes        []types.ChatModeId
	OpenCodeCommands []types.OpenCodeCommandId
	OpenCodeModes    []types.OpenCodeModeId
	OpenCodePlugins  []string
}

// CompileContext carries the information adapters need to compile per-tool
// MCP configuration. Unlike AdapterContext (used by Install), this is
// scope-aware so compile writes to the correct per-scope path.
type CompileContext struct {
	// TargetDir is the project root (always). Adapters use it to read the
	// canonical .ai/mcp.json and to locate project-scoped config paths.
	TargetDir string
	// HomeDir is the user's home dir, required when SetupScope is global.
	HomeDir string
	// SetupScope tells adapters which on-disk layout to emit for.
	SetupScope types.SetupScope
	// FileRecords accumulates records of all files written during compile.
	FileRecords []types.TrackedFile
	// LocalSecrets, when true, routes Claude Code MCP to the gitignored
	// `.claude/settings.local.json` instead of the committed `.mcp.json`.
	LocalSecrets bool
	// WorkspaceRoot is the canonical workspace root directory. At workspace
	// scope this may differ from TargetDir (the planning repo); at project scope
	// it is empty. Spec 022 / E2.3: propagation logic uses WorkspaceRoot + Repos
	// to write per-repo tool configs without colliding with the planning-repo write.
	WorkspaceRoot string
	// Repos lists the per-repo subdirectories the workspace owns. Empty
	// outside workspace scope. Each repo's Path is relative to
	// WorkspaceRoot; PropagateMcpToRepos walks this list to write tool
	// configs into each repo.
	Repos []types.RepoInfo
	// Tools is the set of selected tool targets for this compile. When
	// non-empty, propagation (and root compile) MUST only write configs
	// for these tools. Empty means "all registered tools" (legacy
	// behavior). Spec 022 / E2.3: workspace propagation uses this to
	// honor the same target selection as root compile.
	Tools []types.ToolId
}

// toAdapterContext builds a minimal AdapterContext suitable for calling
// ResolveToolRoot from a CompileContext.
func (c CompileContext) toAdapterContext() *AdapterContext {
	return &AdapterContext{
		TargetDir:     c.TargetDir,
		HomeDir:       c.HomeDir,
		SetupScope:    c.SetupScope,
		WorkspaceRoot: c.WorkspaceRoot,
	}
}

// ToolAdapter is the interface each tool adapter must implement.
type ToolAdapter interface {
	// ID returns the tool identifier (e.g. "opencode", "claude-code").
	ID() types.ToolId
	// Name returns the human-readable tool name.
	Name() string
	// ConfigDir returns the tool's config directory name (e.g., ".opencode", ".claude").
	// Returns empty string if the tool has no single config directory.
	ConfigDir() string
	// Capabilities reports which canonical surfaces this adapter can emit and
	// its support level (stable/beta/...). Declarative metadata; see TECHSPEC §5.2.
	Capabilities() Capability
	// Install copies agents, skills, and config files for this tool.
	// Returns the list of tracked files that were created or modified.
	Install(ctx *AdapterContext) ([]types.TrackedFile, error)
	// CompileMCP generates the per-tool MCP config from the canonical .ai/mcp.json.
	// The ctx carries scope information so the adapter writes to the correct
	// on-disk path (project/workspace vs global).
	CompileMCP(ctx CompileContext) ([]types.TrackedFile, error)
	// CanRunHeadless returns true if this adapter supports headless CLI mode
	// for validation or generation.
	CanRunHeadless() bool
	// RunHeadlessValidation runs a headless validation command if the tool is
	// installed. Non-fatal: logs a warning on error and returns nil.
	RunHeadlessValidation(ctx *AdapterContext) error
	// RunHeadlessInit attempts to fill AGENTS.md placeholders via this tool's
	// headless CLI mode. prompt is the populate instruction text. Returns nil
	// on success or if the tool binary is not available. Errors are non-fatal
	// — the caller logs and continues.
	RunHeadlessInit(ctx *AdapterContext, prompt string) error
}
