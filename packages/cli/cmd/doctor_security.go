package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/internal/validate"
)

// noSandboxTools are targets that run hooks/commands without an enforced
// sandbox; doctor warns operators to review trust before enabling (FR-011).
var noSandboxTools = map[types.ToolId]string{
	types.ToolIdPi:   "Pi executes hooks/commands with no sandbox — review hook and MCP trust.",
	types.ToolIdKiro: "Kiro executes agent actions with no sandbox — review hook and MCP trust.",
}

// mcpServerSummary describes one MCP server for the security inventory.
type mcpServerSummary struct {
	Name      string
	Transport string // "command" or "url"
	EnvKind   string // "none", "ref", or "inline"
}

// sandboxCaveat is a trust/sandbox advisory for a configured tool.
type sandboxCaveat struct {
	Tool types.ToolId
	Note string
}

// securityReport is the structured FR-011 doctor security report.
type securityReport struct {
	MCPServers  []mcpServerSummary
	HookRisks   []validate.Issue
	SecretRisks []validate.Issue
	PathRisks   []validate.Issue
	Sandbox     []sandboxCaveat
}

// HasFindings reports whether the security report contains anything worth
// printing.
func (s securityReport) HasFindings() bool {
	return len(s.MCPServers) > 0 || len(s.HookRisks) > 0 || len(s.SecretRisks) > 0 ||
		len(s.PathRisks) > 0 || len(s.Sandbox) > 0
}

// buildSecurityReport assembles the FR-011 security report: MCP inventory, hook
// risks, inline-secret risks, path/symlink risks, and sandbox caveats for the
// configured tools.
func buildSecurityReport(root string, tools []types.ToolId, profile validate.Profile) securityReport {
	var report securityReport

	report.MCPServers = inventoryMCPServers(filepath.Join(root, ".ai"))

	// Reuse the consolidated validators for security-relevant findings.
	if info, err := os.Stat(filepath.Join(root, ".ai")); err == nil && info.IsDir() {
		vr := validate.All(validate.Options{Root: root, Profile: profile})
		for _, issue := range vr.Issues {
			switch issue.Rule {
			case "hook":
				report.HookRisks = append(report.HookRisks, issue)
			case "secret":
				report.SecretRisks = append(report.SecretRisks, issue)
			case "path":
				report.PathRisks = append(report.PathRisks, issue)
			}
		}
	}

	for _, tool := range tools {
		if note, ok := noSandboxTools[tool]; ok {
			report.Sandbox = append(report.Sandbox, sandboxCaveat{Tool: tool, Note: note})
		}
	}
	sort.Slice(report.Sandbox, func(i, j int) bool { return report.Sandbox[i].Tool < report.Sandbox[j].Tool })

	return report
}

// inventoryMCPServers reads .ai/mcp.json (or .jsonc) and summarizes each server.
func inventoryMCPServers(aiDir string) []mcpServerSummary {
	path := filepath.Join(aiDir, "mcp.json")
	if _, err := os.Stat(path); err != nil {
		alt := filepath.Join(aiDir, "mcp.jsonc")
		if _, altErr := os.Stat(alt); altErr != nil {
			return nil
		}
		path = alt
	}
	doc, err := jsonc.ReadJSONCFile(path)
	if err != nil {
		return nil
	}
	servers, _ := doc["servers"].(map[string]any)
	if servers == nil {
		servers, _ = doc["mcpServers"].(map[string]any)
	}
	if servers == nil {
		return nil
	}
	var summaries []mcpServerSummary
	for name, raw := range servers {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		transport := "command"
		if _, hasURL := entry["url"]; hasURL {
			transport = "url"
		}
		summaries = append(summaries, mcpServerSummary{
			Name:      name,
			Transport: transport,
			EnvKind:   classifyEnv(entry["env"]),
		})
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Name < summaries[j].Name })
	return summaries
}

// classifyEnv reports whether a server's env block uses references, holds an
// inline literal, or is empty.
func classifyEnv(raw any) string {
	env, ok := raw.(map[string]any)
	if !ok || len(env) == 0 {
		return "none"
	}
	hasInline := false
	for _, v := range env {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if s = strings.TrimSpace(s); s != "" && !strings.Contains(s, "$") {
			hasInline = true
		}
	}
	if hasInline {
		return "inline"
	}
	return "ref"
}

// printSecurityReport renders the security report under doctor's output. It is
// intentionally advisory: it never changes doctor's pass/fail health.
func printSecurityReport(report securityReport, warnStyle, headerStyle styleRenderer) {
	if !report.HasFindings() {
		return
	}
	fmt.Println()
	fmt.Println(headerStyle.Render("🔒 Security Report"))

	if len(report.MCPServers) > 0 {
		fmt.Printf("  %s MCP servers (%d):\n", warnStyle.Render("•"), len(report.MCPServers))
		for _, s := range report.MCPServers {
			fmt.Printf("    %s %s [%s, env: %s]\n", warnStyle.Render("•"), s.Name, s.Transport, s.EnvKind)
		}
	}
	printIssueGroup("Hook risks", report.HookRisks, warnStyle)
	printIssueGroup("Inline-secret risks", report.SecretRisks, warnStyle)
	printIssueGroup("Path/symlink risks", report.PathRisks, warnStyle)

	if len(report.Sandbox) > 0 {
		fmt.Printf("  %s Trust/sandbox caveats:\n", warnStyle.Render("!"))
		for _, c := range report.Sandbox {
			fmt.Printf("    %s %s: %s\n", warnStyle.Render("!"), c.Tool, c.Note)
		}
	}
}

func printIssueGroup(title string, issues []validate.Issue, warnStyle styleRenderer) {
	if len(issues) == 0 {
		return
	}
	fmt.Printf("  %s %s (%d):\n", warnStyle.Render("!"), title, len(issues))
	for _, issue := range issues {
		fmt.Printf("    %s %s: %s\n", warnStyle.Render("!"), issue.File, issue.Message)
	}
}

// styleRenderer is the minimal lipgloss surface used by the security report,
// kept as an interface so tests can pass a no-op renderer.
type styleRenderer interface {
	Render(...string) string
}
