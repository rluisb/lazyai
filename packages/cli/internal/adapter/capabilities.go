package adapter

// SupportLevel classifies how thoroughly an adapter is verified against its
// host tool's official documentation. It mirrors TECHSPEC §5.3.
type SupportLevel string

const (
	// SupportStable means official docs verified + golden tests + smoke tests.
	SupportStable SupportLevel = "stable"
	// SupportBeta means official docs verified + golden tests, limited runtime smoke.
	SupportBeta SupportLevel = "beta"
	// SupportExperimental means docs partially verified or host tool still moving quickly.
	SupportExperimental SupportLevel = "experimental"
	// SupportDeprecated means the adapter is kept for migration only.
	SupportDeprecated SupportLevel = "deprecated"
)

// Capability declares which canonical surfaces an adapter can emit for its
// host tool, plus the adapter's support level. The boolean fields mirror the
// capability model in TECHSPEC §5.2; the values below are grounded in the
// Official Tool Compliance Matrix (verification date 2026-06-21). Capability
// is declarative metadata consumed by compile/doctor output and conformance
// tests; it does not itself perform compilation.
type Capability struct {
	// Support is the adapter's verification/support level.
	Support SupportLevel

	RootInstructions bool
	Agents           bool
	Subagents        bool
	Skills           bool
	Hooks            bool
	Commands         bool
	PromptTemplates  bool
	ChatModes        bool
	MCP              bool
	Permissions      bool
	Plugins          bool
	Specs            bool
	Steering         bool
	Compaction       bool
	Sessions         bool
	GlobalConfig     bool
}

// IsBeta reports whether the adapter is below stable (beta or experimental),
// i.e. it should be surfaced to users with a maturity warning.
func (c Capability) IsBeta() bool {
	return c.Support == SupportBeta || c.Support == SupportExperimental
}

// Capabilities reports the OpenCode adapter's surfaces. Matrix §1/§2:
// AGENTS.md root instructions, agents, subagents, permissions, MCP, skills,
// commands, plugins.
func (a *OpenCodeAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Agents:           true,
		Subagents:        true,
		Skills:           true,
		Hooks:            true,
		Commands:         true,
		MCP:              true,
		Permissions:      true,
		Plugins:          true,
		GlobalConfig:     true,
	}
}

// Capabilities reports the Claude Code adapter's surfaces. Matrix §1: CLAUDE.md
// root instructions, agents, subagents, skills, hooks, MCP, permissions,
// plugins, commands, managed (global) settings.
func (a *ClaudeCodeAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Agents:           true,
		Subagents:        true,
		Skills:           true,
		Hooks:            true,
		Commands:         true,
		MCP:              true,
		Permissions:      true,
		Plugins:          true,
		GlobalConfig:     true,
	}
}

// Capabilities reports the GitHub Copilot adapter's surfaces. Matrix §1/§4:
// repo + path instructions, custom agents, skills, hooks, MCP, plugins,
// prompt templates, chat modes.
func (a *CopilotAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Agents:           true,
		Skills:           true,
		Hooks:            true,
		PromptTemplates:  true,
		ChatModes:        true,
		MCP:              true,
		Plugins:          true,
	}
}

// Capabilities reports the Pi adapter's surfaces. Matrix §1: AGENTS.md rules,
// skills, prompt templates, hooks, MCP (where supported), compaction,
// packages (plugins), settings (global config).
func (a *PiAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Skills:           true,
		Hooks:            true,
		PromptTemplates:  true,
		MCP:              true,
		Plugins:          true,
		Compaction:       true,
		GlobalConfig:     true,
	}
}

// Capabilities reports the OMP (Oh My Pi) adapter's surfaces. Promoted to stable
// after every emitted surface was verified against the authoritative OMP docs
// (the in-harness `omp://` documentation set); see
// docs/adapters/snapshots/beta-adapter-verification-2026-06.md. Verified emit
// surfaces: root AGENTS.md, .omp/agents, .omp/skills, .omp/hooks/pre/*.ts,
// .omp/commands, .omp/mcp.json. Plugins/Compaction/Sessions/GlobalConfig are
// host-support metadata (OMP supports them) rather than emitted files.
func (a *OmpAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Agents:           true,
		Skills:           true,
		Hooks:            true,
		Commands:         true,
		MCP:              true,
		Plugins:          true,
		Compaction:       true,
		Sessions:         true,
		GlobalConfig:     true,
	}
}

// Capabilities reports the Antigravity/Gemini adapter's surfaces. Promoted to
// stable after the two #486 beta gaps were closed and pinned by conformance
// tests: global-scope skills now write the documented ~/.gemini/config/skills
// root, and root instructions are discovered (GEMINI.md for Gemini CLI via the
// scaffold, .agents/rules/lazyai.md for Antigravity IDE workspaces). See
// docs/adapters/snapshots/beta-adapter-verification-2026-06.md. Surfaces:
// .agents/skills, .agents/hooks.json + .gemini/settings.json hooks,
// ~/.gemini/config/mcp_config.json MCP, plus permissions/plugins/config
// host-support metadata.
func (a *AntigravityAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Skills:           true,
		Hooks:            true,
		MCP:              true,
		Permissions:      true,
		Plugins:          true,
		GlobalConfig:     true,
	}
}

// Capabilities reports the Kiro adapter's verified surfaces. Kiro installs
// agents, skills, prompts, hooks, MCP, permissions, and global config. Hooks
// are native .kiro/hooks/<name>.json files emitted during Install. Specs and
// steering are intentionally absent.
func (a *KiroAdapter) Capabilities() Capability {
	return Capability{
		Support:          SupportStable,
		RootInstructions: true,
		Agents:           true,
		Skills:           true,
		Hooks:            true,
		MCP:              true,
		Permissions:      true,
		PromptTemplates:  true,
		GlobalConfig:     true,
	}
}
