// Package compiler provides template compilation for the ai-setup project.
// Ported from the TypeScript compiler/ module.
package compiler

// ToolOverrides contains per-tool template metadata.
type ToolOverrides struct {
	Description string
	Notes       string
	RootFile    string
}

// ToolOverrideMap maps tool IDs to their template overrides.
var ToolOverrideMap = map[string]ToolOverrides{
	"opencode": {
		Description: "This project uses OpenCode with ai-setup integration.",
		Notes: "## OpenCode-Specific Notes\n\n" +
			"- Project config: `opencode.json` at project root\n" +
			"- Agents: `.opencode/agents/<name>.md`\n" +
			"- Skills: `.opencode/skills/<name>/SKILL.md`\n" +
			"- Commands: `.opencode/commands/<name>.md`\n" +
			"- Multiple config sources merged (project -> global -> env)",
	},
	"claude-code": {
		Description: "This project uses Claude Code with ai-setup integration.",
		Notes: "## Claude Code-Specific Notes\n\n" +
			"- Project settings: `.claude/settings.json`\n" +
			"- Modular rules: `.claude/rules/<name>.md` (supports `paths` frontmatter for scoping)\n" +
			"- Skills: `.claude/skills/<name>/SKILL.md`\n" +
			"- Agents: `.claude/agents/<name>.md`\n" +
			"- Personal overrides: `CLAUDE.local.md` (gitignore this)",
	},
	"copilot": {
		Description: "This project uses GitHub Copilot with ai-setup integration.",
		RootFile:    "copilot-instructions.md",
		Notes: "## Copilot-Specific Notes\n\n" +
			"- Repository-wide instructions: `.github/copilot-instructions.md`\n" +
			"- Path-specific instructions: `.github/instructions/<name>.instructions.md` with `applyTo` frontmatter\n" +
			"- Reusable prompts: `.github/prompts/<name>.prompt.md`\n" +
			"- Agent instructions: `AGENTS.md` at project root",
	},
	"gemini": {
		Description: "This project uses Gemini CLI with ai-setup integration.",
		RootFile:    "GEMINI.md",
		Notes: "## Gemini CLI-Specific Notes\n\n" +
			"- Project settings: `.gemini/settings.json`\n" +
			"- Skills: `.gemini/skills/<name>/SKILL.md`\n" +
			"- No agents concept (agents are inline in GEMINI.md)\n" +
			"- Context traversal: walks up to .git boundary loading GEMINI.md files",
	},
	"codex": {
		Description: "This project uses OpenAI Codex CLI with ai-setup integration.",
		Notes: "## Codex-Specific Notes\n\n" +
			"- Agents are defined inline in this file (no separate agents directory)\n" +
			"- Skills use AgentSkills standard: `.agents/skills/*/SKILL.md`\n" +
			"- No `.codex/` project directory (global config only in `~/.codex/`)",
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