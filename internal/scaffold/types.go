// Package scaffold provides functions for scaffolding the ai-setup directory
// structure, including agents, skills, prompts, specs, templates, rules,
// infrastructure files, and orchestration definitions.
// Ported from the TypeScript modules in src/scaffold/.
package scaffold

import (
	"io/fs"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	// scaffolding to the tool's own CLI (e.g. `gemini mcp add`).
	DriveCLI bool
	// Agents lists agent IDs to install.
	Agents []types.AgentId
	// Skills lists skill IDs to install.
	Skills []types.SkillId
	// Prompts lists prompt IDs to install.
	Prompts []types.PromptId
	// Commands lists Gemini custom command IDs to install.
	Commands []types.CommandId
	// ChatModes lists Copilot chat mode IDs to install.
	ChatModes []types.ChatModeId
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
	// Optional context overrides for compiled root.
	PrimaryLanguage     string
	Framework           string
	WorkspaceType       string
	ProjectInstructions string
}

// ScaffoldResult holds the outcome of a scaffold run.
type ScaffoldResult struct {
	Files       []types.TrackedFile
	Directories []string
	Errors      []error
}
