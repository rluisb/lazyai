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
