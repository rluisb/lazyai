// Package scaffold provides functions for scaffolding the ai-setup directory
// structure, including agents, skills, prompts, specs, templates, rules,
// infrastructure files, and orchestration definitions.
// Ported from the TypeScript modules in src/scaffold/.
package scaffold

import (
	"io/fs"

	"github.com/rluisb/lazyai/packages/cli/internal/compiler"
	reversa "github.com/rluisb/lazyai/packages/cli/internal/reversa/scout"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldContext carries all the information needed for a full scaffold run.
type ScaffoldContext struct {
	// TargetDir is the project root directory where files will be installed.
	TargetDir string
	// LibraryDir is the path to the library containing source templates.
	LibraryDir string
	// Tools lists the AI tools to scaffold for.
	Tools []types.ToolId
	// CLITools lists CLI tool names that may also be MCP servers.
	CLITools []string
	// EnableServers lists MCP server names to enable.
	EnableServers []string
	// ProjectName is the human-readable project name.
	ProjectName string
	// PlanningDir is the directory for planning artifacts (specs, adrs).
	PlanningDir string
	// SetupScope is the installation scope.
	SetupScope types.SetupScope
	// ExistingSetupPolicy controls how existing setup state is handled.
	ExistingSetupPolicy types.SetupPolicy
	// HomeDir is the user's home directory for global scope resolution.
	HomeDir string
	// Features controls which features are compiled into root files.
	Features *types.FeatureFlags
	// GitConventions defines branch and commit conventions.
	GitConventions *types.GitConventions
	// Strategy controls how file conflicts are handled.
	Strategy types.ConflictStrategy
	// PerFileOverrides allows per-file conflict strategy overrides.
	PerFileOverrides map[string]types.ConflictStrategy
	// Force overwrites existing files without conflict checks.
	Force bool
	// DryRun reports what would be done without writing files.
	DryRun bool
	// DriveCLI, when true, asks adapters that support it to delegate
	// scaffolding to the tool's own CLI.
	DriveCLI bool
	// LocalSecrets, when true, routes Claude Code MCP/settings writes to
	// the gitignored .claude/settings.local.json instead of committed
	// surfaces (.mcp.json / .claude/settings.json). Opt-in; default false.
	LocalSecrets bool
	// Agents lists agent IDs to install.
	Agents []types.AgentId
	// Skills lists skill IDs to install.
	Skills []types.SkillId
	// Prompts lists prompt IDs to install.
	Prompts []types.PromptId
	// ChatModes lists Copilot chat mode IDs to install.
	ChatModes []types.ChatModeId
	// OpenCodeCommands lists opencode slash command IDs to install.
	OpenCodeCommands []types.OpenCodeCommandId
	// OpenCodeModes lists opencode chat mode IDs to install.
	OpenCodeModes []types.OpenCodeModeId
	// OpenCodePlugins lists opencode plugin npm module names to install.
	OpenCodePlugins []string
	// Templates lists template IDs to install.
	Templates []types.TemplateId
	// Rules lists rule IDs to install.
	Rules []types.RuleId
	// Infra lists infrastructure components to install.
	Infra []types.InfraId
	// SpecsDirs lists specs subdirectories to create.
	SpecsDirs []string
	// Housekeeping config controls optional memory/bootstrap scaffolding.
	Housekeeping *types.HousekeepingConfig
	// Repos lists workspace repos for workspace scope.
	Repos []types.RepoInfo
	// PlanningRepoPath is the path to the planning repo (for workspace scope).
	PlanningRepoPath string
	// LibraryFS is the filesystem for reading library data (embedded or disk).
	LibraryFS fs.FS
	// StoreData optionally carries persisted config values used by compiled roots.
	StoreData *types.StoreData
	// SurfaceData optionally carries deterministic Scout analysis results.
	SurfaceData *reversa.SurfaceData
	// Inferred fields from Scout (populated by buildScaffoldContext when Scout runs).
	MigrationsPath     string
	TestPath           string
	StrictMode         string
	InstallCommand     string
	ProtectedBranchGit string // git-detected default branch, used if no wizard override
	// Optional context overrides for compiled root.
	PrimaryLanguage     string
	Framework           string
	WorkspaceType       string
	ProjectInstructions string
	ProjectDescription  string // optional; substituted into AGENTS.md [YOUR_PROJECT_DESCRIPTION]
	Organization        string // optional; substituted into AGENTS.md [YOUR_ORG]
	Team                string // optional; substituted into AGENTS.md [YOUR_TEAM]
	ProjectOverview     string
	Database            string
	ORM                 string
	TestFramework       string
	PackageManager      string
	NamingConventions   string
	ErrorHandling       string
	APIConventions      string
	ImportOrder         string
	ProtectedBranch     string
	TestCommand         string
	LintCommand         string
	BuildCommand        string
	CoverageThreshold   int
	CodebaseMap         []compiler.CodebaseMapEntry
}

// ScaffoldResult holds the outcome of a scaffold run.
type ScaffoldResult struct {
	Files       []types.TrackedFile
	Directories []string
	Errors      []error
}
