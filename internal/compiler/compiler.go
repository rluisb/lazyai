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
			"- Project config: `opencode.jsonc` at project root\n" +
			"- Agents: `.opencode/agents/<name>.md`\n" +
			"- Skills: `.opencode/skills/<name>/SKILL.md`\n" +
			"- Commands: `.opencode/commands/<name>.md`\n" +
			"- Multiple config sources merged (project -> global -> env)",
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