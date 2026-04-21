// Package adapter provides tool-specific adapters that know how to install
// agents, skills, and configuration files for each supported AI coding tool.
// Ported from the TypeScript adapters in src/adapters/.
package adapter

import (
	"io/fs"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// AdapterContext carries all the information an adapter needs to perform
// installation. It is the Go equivalent of the TypeScript AdapterContext.
type AdapterContext struct {
	// TargetDir is the project root directory where files will be installed.
	TargetDir string
	// SetupScope is the installation scope (global, workspace, project).
	SetupScope types.SetupScope
	// HomeDir is the user's home directory, used for global scope path resolution.
	HomeDir string
	// LibraryDir is the path to the library containing source templates.
	LibraryDir string
	// LibraryFS is the filesystem for reading library data (embedded or disk).
	LibraryFS fs.FS
	// FileRecords accumulates records of all files written during installation.
	FileRecords []types.TrackedFile
	// EnableServers lists MCP servers to enable (e.g. "orchestrator").
	EnableServers []string
	// Force overwrites existing files without prompting.
	Force bool
	// DryRun reports what would be done without writing files.
	DryRun bool
	// DriveCLI, when true, asks adapters that support it to delegate
	// scaffolding to the tool's own CLI (e.g. `gemini mcp add`) instead of
	// direct-write. Falls back silently to direct-write when the binary is
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
}

// AdapterSelections holds the user's selections for which items to install.
type AdapterSelections struct {
	Agents           []types.AgentId
	Skills           []types.SkillId
	Prompts          []types.PromptId
	Commands         []types.CommandId
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
}

// toAdapterContext builds a minimal AdapterContext suitable for calling
// ResolveToolRoot / ResolveCodexRoots from a CompileContext.
func (c CompileContext) toAdapterContext() *AdapterContext {
	return &AdapterContext{
		TargetDir:  c.TargetDir,
		HomeDir:    c.HomeDir,
		SetupScope: c.SetupScope,
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
}
