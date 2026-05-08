package cmd

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
)

// GetMcpNextSteps aggregates setup instructions and returns a styled Next Steps block.
func GetMcpNextSteps(activeServers map[string]adapter.McpServer) string {
	hasEnv := false
	var hints []string

	// Sort keys to ensure deterministic output
	var keys []string
	for name := range activeServers {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		server := activeServers[name]
		if len(server.Env) > 0 {
			hasEnv = true
		}
		if server.SetupHint != "" {
			hints = append(hints, server.SetupHint)
		}
	}

	if !hasEnv && len(hints) == 0 {
		return ""
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	bulletStyle := lipgloss.NewStyle().Foreground(theme.Success)
	textStyle := lipgloss.NewStyle().Foreground(theme.Text)

	var sb strings.Builder
	sb.WriteString(headerStyle.Render("🚀 Next Steps: MCP Server Configuration"))
	sb.WriteString("\n\n")

	if hasEnv {
		sb.WriteString(fmt.Sprintf("  %s %s\n", bulletStyle.Render("•"), textStyle.Render("Fill in any required environment variables (e.g. ${API_KEY}) in your .ai/mcp.json.")))
		sb.WriteString(fmt.Sprintf("    %s\n\n", lipgloss.NewStyle().Foreground(theme.Dimmed).Render("Run 'lazyai-cli compile' again if you update variables that tools need at compile time.")))
	}

	if len(hints) > 0 {
		for _, hint := range hints {
			sb.WriteString(fmt.Sprintf("  %s %s\n", bulletStyle.Render("•"), textStyle.Render(hint)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// PrintMcpNextSteps aggregates setup instructions and prints a styled Next Steps block.
func PrintMcpNextSteps(activeServers map[string]adapter.McpServer) {
	steps := GetMcpNextSteps(activeServers)
	if steps != "" {
		fmt.Println()
		fmt.Print(steps)
	}
}
