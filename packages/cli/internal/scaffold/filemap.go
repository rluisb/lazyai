package scaffold

import "github.com/rluisb/lazyai/packages/cli/internal/types"

// RootFileByTool maps each ToolId to the root configuration filename it uses.
// Ported from src/scaffold/root-file-map.ts.
var RootFileByTool = map[types.ToolId]string{
	types.ToolIdOpenCode:    "AGENTS.md",
	types.ToolIdClaudeCode:  "AGENTS.md",
	types.ToolIdCopilot:     ".github/copilot-instructions.md",
	types.ToolIdPi:          "AGENTS.md",
	types.ToolIdOmp:         "AGENTS.md",
	types.ToolIdKiro:        "AGENTS.md",
	types.ToolIdAntigravity: "AGENTS.md",
	types.ToolIdCodex:       "AGENTS.md",
}
