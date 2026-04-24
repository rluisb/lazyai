package scaffold

import "github.com/ricardoborges-teachable/ai-setup/internal/types"

// RootFileByTool maps each ToolId to the root configuration filename it uses.
var RootFileByTool = map[types.ToolId]string{
	types.ToolIdOpenCode: "AGENTS.md",
}
