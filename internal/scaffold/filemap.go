package scaffold

import "github.com/ricardoborges-teachable/ai-setup/internal/types"

// RootFileByTool maps each ToolId to the root configuration filename it uses.
// Ported from src/scaffold/root-file-map.ts.
var RootFileByTool = map[types.ToolId]string{
	types.ToolIdOpenCode:   "AGENTS.md",
	types.ToolIdClaudeCode: "CLAUDE.md",
	types.ToolIdGemini:     "GEMINI.md",
	types.ToolIdCopilot:    ".github/copilot-instructions.md",
	types.ToolIdCodex:      "AGENTS.md",
}
