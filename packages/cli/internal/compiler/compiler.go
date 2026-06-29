// Package compiler provides template compilation for the ai-setup project.
// Ported from the TypeScript compiler/ module.
package compiler

import "github.com/rluisb/lazyai/packages/cli/internal/types"

// ToolOverrides contains per-tool template metadata.
type ToolOverrides struct {
	Description string
	Notes       string
	RootFile    string
}

// ToolOverrideMap maps tool IDs to their template overrides.
var ToolOverrideMap = map[string]ToolOverrides{
	string(types.ToolIdOpenCode): {
		Description: "This project uses OpenCode with LazyAI integration.",
		Notes: "## OpenCode-Specific Notes\n\n" +
			"- Project config: `opencode.json`\n" +
			"- LazyAI-only MCP extras: `.opencode/lazyai.mcp.jsonc`\n" +
			"- Agents: `.opencode/agents/<name>.md`\n" +
			"- Skills: `.opencode/skills/<name>/SKILL.md`\n" +
			"- Commands: `.opencode/commands/<name>.md`\n" +
			"- Multiple config sources merged (project -> global -> env)",
	},
	string(types.ToolIdClaudeCode): {
		Description: "This project uses Claude Code with LazyAI integration.",
		Notes: "## Claude Code-Specific Notes\n\n" +
			"- Project settings: `.claude/settings.json`\n" +
			"- Modular rules: `.claude/rules/<name>.md` (supports `paths` frontmatter for scoping)\n" +
			"- Skills: `.claude/skills/<name>/SKILL.md`\n" +
			"- Agents: `.claude/agents/<name>.md`\n" +
			"- Personal overrides: `CLAUDE.local.md` (gitignore this)",
	},
	string(types.ToolIdCopilot): {
		Description: "This project uses GitHub Copilot with LazyAI integration.",
		RootFile:    "copilot-instructions.md",
		Notes: "## Copilot-Specific Notes\n\n" +
			"- Repository-wide instructions: `.github/copilot-instructions.md`\n" +
			"- Path-specific instructions: `.github/instructions/<name>.instructions.md` with `applyTo` frontmatter\n" +
			"- Reusable prompts: `.github/prompts/<name>.prompt.md`\n" +
			"- Agent instructions: `AGENTS.md` at project root",
	},
	string(types.ToolIdPi): {
		Description: "This project uses Pi with LazyAI integration.",
		Notes:       "## Pi-Specific Notes\n\n- Agents: `.pi/agents/<name>.md`\n- Skills: `.pi/skills/<name>/SKILL.md`\n- Prompts: `.pi/prompts/<name>.md`\n- Safety hooks ship as `.pi/extensions/*.ts`\n- System prompt: `.pi/SYSTEM.md` (replaces default) / `.pi/APPEND_SYSTEM.md` (appends).",
	},
	string(types.ToolIdOmp): {
		Description: "This project uses OMP with LazyAI integration.",
		Notes:       "## OMP-Specific Notes\n\n- Agents: `.omp/agents/<name>.md`\n- Skills: `.omp/skills/<name>/SKILL.md`\n- Project instructions: `AGENTS.md`.",
	},
	string(types.ToolIdKiro): {
		Description: "This project uses Kiro with LazyAI integration.",
		Notes:       "## Kiro-Specific Notes\n\n- Agents: `.kiro/agents/<name>.md`\n- Skills: `.kiro/skills/<name>/SKILL.md`\n- Prompts: `.kiro/prompts/<name>.md`.",
	},
	string(types.ToolIdAntigravity): {
		Description: "This project uses Antigravity with LazyAI integration.",
		Notes:       "## Antigravity-Specific Notes\n\n- Gemini settings: `.gemini/settings.json`\n- Hooks: `.gemini/hooks/lazyai/*`\n- Skills: `.agents/skills/<name>/SKILL.md`.",
	},
	string(types.ToolIdCodex): {
		Description: "This project uses Codex (OpenAI Codex CLI) with LazyAI integration.",
		Notes: "## Codex-Specific Notes\n\n" +
			"- Project instructions: `AGENTS.md` at project root\n" +
			"- MCP servers: `.codex/config.toml` `[mcp_servers.<name>]`\n" +
			"- Subagents: `.codex/agents/<name>.toml` (name/description/developer_instructions)\n" +
			"- Skills: `.agents/skills/<name>/SKILL.md` (read natively by Codex)\n" +
			"- Hooks: `.codex/hooks.json`.",
	},
}

// CompiledFile represents a single compiled output file.
type CompiledFile struct {
	RelativePath string
	Content      string
}

// CompiledOutput holds the compilation results for a single tool.
type CompiledOutput struct {
	Tool  string
	Files []CompiledFile
}
