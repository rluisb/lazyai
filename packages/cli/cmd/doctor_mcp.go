package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// staleClaudeMcpEntry describes an MCP server entry in Claude's user-scope
// config (`~/.claude.json`) that points at the legacy ai-setup orchestrator
// binary or npm package. Such entries collide with the project-scope
// `lazyai-orchestrator connect` definition LazyAI writes today, producing
// Claude Code's cross-scope conflict warning on every launch (#209).
type staleClaudeMcpEntry struct {
	// Name is the MCP server name as keyed in mcpServers (e.g., "orchestrator").
	Name string `json:"name"`
	// Scope is always "user" for now — `~/.claude.json` is the user-scope
	// surface. Kept explicit so JSON consumers and remediation text remain
	// unambiguous if we later widen detection to other scopes.
	Scope string `json:"scope"`
	// Command is the legacy command string we matched on (e.g.,
	// "ai-setup-orchestrator"). Empty when the match was via args.
	Command string `json:"command,omitempty"`
	// Reason summarizes why the entry was flagged — surfaced in human output.
	Reason string `json:"reason"`
	// Remediation is the exact CLI command a user can run to remove the stale
	// entry. Pre-formatted so doctor output and JSON consumers share one
	// string of truth.
	Remediation string `json:"remediation"`
}

// legacyOrchestratorCommands lists the binary names LazyAI shipped before the
// `lazyai-orchestrator` rename. Any user-scope MCP entry whose `command` matches
// one of these is a stale carry-over.
var legacyOrchestratorCommands = []string{
	"ai-setup-orchestrator",
}

// legacyOrchestratorArgFragments lists substrings that, when found in an
// entry's `args` array, identify a stale ai-setup orchestrator entry. The
// `@ai-setup/orchestrator` form covers the npm-package shape some users may
// have if they registered the server via `npx`.
var legacyOrchestratorArgFragments = []string{
	"ai-setup-orchestrator",
	"@ai-setup/orchestrator",
}

// findStaleClaudeMcpEntries scans `<homeDir>/.claude.json` for MCP server
// entries that still reference the legacy ai-setup orchestrator binary or
// package. Returns an empty slice (no error) when the config file is absent
// or has no MCP servers — those are normal, healthy states.
//
// The function never returns an error for malformed JSON or missing keys;
// doctor diagnostics should degrade quietly rather than block on a config
// surface owned by another tool. Hard errors are reserved for I/O failures
// the caller may want to surface (e.g., permission denied reading $HOME).
func findStaleClaudeMcpEntries(homeDir string) ([]staleClaudeMcpEntry, error) {
	if homeDir == "" {
		return nil, nil
	}
	path := filepath.Join(homeDir, ".claude.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var raw struct {
		McpServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Malformed JSON in a config file we don't own: skip silently.
		return nil, nil
	}

	var stale []staleClaudeMcpEntry
	for name, entry := range raw.McpServers {
		match, matchedCmd := classifyOrchestratorEntry(entry.Command, entry.Args)
		if !match {
			continue
		}
		stale = append(stale, staleClaudeMcpEntry{
			Name:        name,
			Scope:       "user",
			Command:     matchedCmd,
			Reason:      formatStaleReason(name, matchedCmd, entry.Command, entry.Args),
			Remediation: fmt.Sprintf("claude mcp remove %s -s user", name),
		})
	}
	sort.Slice(stale, func(i, j int) bool { return stale[i].Name < stale[j].Name })
	return stale, nil
}

// classifyOrchestratorEntry returns (true, matchedCommand) when the given
// command/args belong to a legacy ai-setup orchestrator entry. The second
// return is the exact string matched, used in the human-facing reason.
func classifyOrchestratorEntry(command string, args []string) (bool, string) {
	cmd := strings.TrimSpace(command)
	for _, legacy := range legacyOrchestratorCommands {
		if cmd == legacy {
			return true, legacy
		}
	}
	for _, arg := range args {
		for _, frag := range legacyOrchestratorArgFragments {
			if strings.Contains(arg, frag) {
				return true, frag
			}
		}
	}
	return false, ""
}

// formatStaleReason produces a one-line explanation for human output. We
// quote the matched substring rather than echoing the full command line to
// keep the diagnostic readable when args are long (npx invocations can be).
func formatStaleReason(name, matched, command string, args []string) string {
	if command != "" && matched == command {
		return fmt.Sprintf("mcpServers.%s.command = %q (legacy ai-setup name)", name, command)
	}
	// Match was via args; surface the first arg that contained the fragment
	// so users can locate it in their config.
	for _, arg := range args {
		if strings.Contains(arg, matched) {
			return fmt.Sprintf("mcpServers.%s.args includes %q (legacy ai-setup reference)", name, arg)
		}
	}
	return fmt.Sprintf("mcpServers.%s references legacy ai-setup orchestrator", name)
}
