// Package adapter provides the Claude Code adapter implementation.
// Ported from the TypeScript ClaudeCodeAdapter.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/configmerge"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ClaudeCodeAdapter implements ToolAdapter for Claude Code (claude CLI).
type ClaudeCodeAdapter struct{}

func (a *ClaudeCodeAdapter) ID() types.ToolId  { return types.ToolIdClaudeCode }
func (a *ClaudeCodeAdapter) Name() string      { return "Claude Code" }
func (a *ClaudeCodeAdapter) ConfigDir() string { return ".claude" }

func (a *ClaudeCodeAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	claudeDir, err := ResolveToolRoot(types.ToolIdClaudeCode, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}
	isGlobal := ctx.SetupScope == types.SetupScopeGlobal

	settingsPath := filepath.Join(claudeDir, "settings.json")
	rulesDir := filepath.Join(claudeDir, "rules")
	hooksDir := filepath.Join(claudeDir, "hooks")
	sampleRulePath := filepath.Join(rulesDir, "typescript.md")

	_ = files.EnsureDir(claudeDir)
	_ = files.EnsureDir(rulesDir)
	_ = files.EnsureDir(hooksDir)
	_ = files.EnsureDir(filepath.Join(claudeDir, "skills"))
	_ = files.EnsureDir(filepath.Join(claudeDir, "agents"))

	// Spec 012 / Task 001: pre-spec installs at global scope wrote agents
	// flat at ~/.claude/<agent>.md, where Claude Code does not discover them.
	// Migrate any such files into ~/.claude/agents/ before re-installing so
	// existing global installs converge on the canonical layout.
	if isGlobal {
		migrateLegacyGlobalAgents(ctx, claudeDir)
	}

	// Merge default settings into settings.json, preserving user keys.
	defaultSettings := map[string]any{
		"permissions": map[string]any{
			"allow": []any{},
			"deny":  []any{},
		},
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{
							"command": "${CLAUDE_PROJECT_DIR:-$PWD}/.claude/hooks/block-destructive-shell.sh",
							"type":    "command",
						},
					},
					"matcher": "Bash",
				},
			},
			"Stop": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{
							"command": "${CLAUDE_PROJECT_DIR:-$PWD}/.claude/hooks/objective-workflow-gate.sh",
							"type":    "command",
						},
					},
				},
			},
		},
	}
	preExisted := files.FileExists(settingsPath)
	if _, err := configmerge.MergeJSONFile(settingsPath, defaultSettings); err != nil {
		return nil, err
	}
	if !preExisted {
		relPath, _ := filepath.Rel(ctx.TargetDir, settingsPath)
		relPath = filepath.ToSlash(relPath)
		hash, _ := files.FileHash(settingsPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	// Write sample rule if it doesn't exist.
	if !files.FileExists(sampleRulePath) {
		content, err := ReadSampleRuleContent(ctx)
		if err != nil {
			return nil, fmt.Errorf("read sample rule from library: %w", err)
		}
		if err := files.WriteFile(sampleRulePath, content, 0o644); err != nil {
			return nil, err
		}
		relPath, _ := filepath.Rel(ctx.TargetDir, sampleRulePath)
		relPath = filepath.ToSlash(relPath)
		hash, _ := files.FileHash(sampleRulePath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	adapterLog.Info("installing tools", "adapter", "claude-code")

	// Copy agents from the canonical seven-agent baseline. Source frontmatter
	// carries LazyAI metadata for other uses; generated Claude Code agents emit
	// only name and description to match the baseline surface.
	if err := copyCanonicalDefaultAgent(ctx,
		filepath.Join(claudeDir, "agents", defaultAgentID+".md"),
		claudeDefaultAgentContent,
	); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "canonical/agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(claudeDir, "agents", file)
		},
		IncludeFile: func(file string) bool {
			return !isDefaultAgentFile(file) && isCanonicalAgentFile(file)
		},
		Transform: func(content []byte) []byte {
			out, err := RewriteAgentForClaudeCode(content, ctx)
			if err != nil {
				adapterLog.Warn("claude-code agent rewrite fell back to verbatim copy", "adapter", "claude-code", "error", err)
				return content
			}
			return out
		},
	}); err != nil {
		return nil, err
	}

	// Copy skills.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(claudeDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Copy Claude Code commands (starter set).
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "claudecode/commands",
		SelectionKey: "claudecode/commands",
		ToDestPath: func(file string) string {
			return filepath.Join(claudeDir, "commands", file)
		},
	}); err != nil {
		return nil, err
	}

	// Copy Claude Code output styles (starter set).

	// Copy generated hook scripts.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "claudecode/hooks",
		ToDestPath: func(file string) string {
			return filepath.Join(hooksDir, file)
		},
		Mode: 0o755,
	}); err != nil {
		return nil, err
	}
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "claudecode/output-styles",
		SelectionKey: "claudecode/output-styles",
		ToDestPath: func(file string) string {
			return filepath.Join(claudeDir, "output-styles", file)
		},
	}); err != nil {
		return nil, err
	}

	// Root AGENTS.md placement is handled centrally by scaffold/root.go.
	// (scope-aware); the adapter no longer installs it.

	// CLI-driven MCP registration: when DriveCLI=true, attempt to register MCP
	// servers via `claude mcp add`. Falls back silently to direct-write (via
	// settings.json merge) on any failure.
	if ctx.DriveCLI {
		if ok := installClaudeMCPViaCLI(ctx, claudeDir); ok {
			adapterLog.Info("mcp servers registered via CLI", "adapter", "claude-code")
		}
	}

	// Post-install verification summary (non-fatal on failure).
	displayInstallSummary(ctx, claudeDir, isGlobal)

	return ctx.FileRecords, nil
}

// installClaudeMCPViaCLI calls `claude mcp add-json` for each enabled MCP server
// found in the canonical .ai/mcp.json. Returns true only if at least one
// server was registered successfully. Falls back silently to direct-write
// (via settings.json merge) on any failure. Safe to ignore the return value —
// settings.json already provides the direct-write baseline.
func installClaudeMCPViaCLI(ctx *AdapterContext, claudeDir string) bool {
	_, found := LookupClaudeBinary()
	if !found {
		adapterLog.Info("claude binary not found; falling back to direct-write for MCP servers", "adapter", "claude-code")
		return false
	}

	mcpRoot := ctx.TargetDir
	if ctx.SetupScope == types.SetupScopeWorkspace && ctx.WorkspaceRoot != "" {
		mcpRoot = ctx.WorkspaceRoot
	}
	catalog := ReadCanonicalMcp(mcpRoot)
	if catalog == nil || len(catalog.Servers) == 0 {
		return false
	}

	runner := &DefaultClaudeCLIRunner{}
	runCtx := context.Background()

	// Map scope to claude mcp add-json scope flag.
	scopeFlag := "project"
	workingDir := mcpRoot
	if ctx.SetupScope == types.SetupScopeGlobal {
		scopeFlag = "user"
		workingDir = ""
	}

	success := false
	for name, srv := range catalog.Servers {
		if !srv.isEnabled() {
			continue
		}

		// Pre-check: is this server already registered?
		checkCtx, cancel := context.WithTimeout(runCtx, 10*time.Second)
		_, _, err := runner.Run(checkCtx, workingDir, "mcp", "get", name)
		cancel()
		if err == nil {
			// Server already exists, skip it
			adapterLog.Info("mcp server already registered, skipping", "adapter", "claude-code", "server", name)
			success = true
			continue
		}

		// Server not found, add it via add-json.
		payload := mcpServerToJSON(srv)
		addCtx, cancel := context.WithTimeout(runCtx, 30*time.Second)
		_, stderr, err := runner.Run(addCtx, workingDir, "mcp", "add-json", name, payload, "-s", scopeFlag)
		cancel()
		if err != nil {
			adapterLog.Error("mcp add-json failed", "adapter", "claude-code", "server", name, "error", err, "stderr", string(stderr))
			// Continue trying other servers, but note the failure.
			continue
		}

		adapterLog.Info("registered MCP server via CLI", "adapter", "claude-code", "server", name)
		success = true
	}
	_ = claudeDir
	return success
}

// mcpServerToJSON converts an MCP catalog server entry to a JSON string
// suitable for `claude mcp add-json`.
func mcpServerToJSON(srv McpServer) string {
	// Build a JSON object matching claude's mcp add-json schema.
	// For stdio servers: {"command": "...", "args": [...], "env": {...}}
	// For http/sse servers: {"type":"http", "url": "...", "headers": {...}}

	var buf strings.Builder
	buf.WriteString("{")
	first := true
	// URL (for HTTP/SSE servers)
	if srv.URL != "" {
		buf.WriteString(`"type":"http","url":"`)
		fmt.Fprintf(&buf, `%s"`, srv.URL)
		// Headers object (for HTTP/SSE servers)
		if len(srv.Headers) > 0 {
			buf.WriteString(`,"headers":{`)
			headerFirst := true
			for k, v := range srv.Headers {
				if !headerFirst {
					buf.WriteString(",")
				}
				fmt.Fprintf(&buf, `"%s":"%s"`, k, v)
				headerFirst = false
			}
			buf.WriteString("}")
		}
		buf.WriteString("}")
		return buf.String()
	}

	// Command (for stdio/subprocess servers)
	if srv.Command != "" {
		fmt.Fprintf(&buf, `"command":"%s"`, srv.Command)
		first = false
	}

	// Args array
	if len(srv.Args) > 0 {
		if !first {
			buf.WriteString(",")
		}
		buf.WriteString(`"args":[`)
		for i, arg := range srv.Args {
			if i > 0 {
				buf.WriteString(",")
			}
			fmt.Fprintf(&buf, `"%s"`, arg)
		}
		buf.WriteString("]")
		first = false
	}

	// Env object
	if len(srv.Env) > 0 {
		if !first {
			buf.WriteString(",")
		}
		buf.WriteString(`"env":{`)
		envFirst := true
		for k, v := range srv.Env {
			if !envFirst {
				buf.WriteString(",")
			}
			fmt.Fprintf(&buf, `"%s":"%s"`, k, v)
			envFirst = false
		}
		buf.WriteString("}")
		first = false
	}

	buf.WriteString("}")
	return buf.String()
}

// migrateLegacyGlobalAgents moves any flat-layout agent files at
// <claudeDir>/<agent>.md (pre-spec-012 install bug) into the canonical
// <claudeDir>/agents/<agent>.md location. Only the agent filenames known to
// the embedded library are considered — unrelated files at the root (notably
// the user's personal CLAUDE.md) are never touched.
func migrateLegacyGlobalAgents(ctx *AdapterContext, claudeDir string) {
	agentsDir := filepath.Join(claudeDir, "agents")

	var agentFiles []string
	if ctx.LibraryFS != nil {
		entries, err := fs.ReadDir(ctx.LibraryFS, "canonical/agents")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				agentFiles = append(agentFiles, e.Name())
			}
		}
	} else if ctx.LibraryDir != "" {
		sourceDir := filepath.Join(ctx.LibraryDir, "canonical", "agents")
		for _, f := range files.ListDir(sourceDir) {
			if files.IsDirectory(filepath.Join(sourceDir, f)) {
				continue
			}
			agentFiles = append(agentFiles, f)
		}
	}

	for _, file := range agentFiles {
		legacyPath := filepath.Join(claudeDir, file)
		newPath := filepath.Join(agentsDir, file)
		if !files.FileExists(legacyPath) {
			continue
		}
		if files.FileExists(newPath) {
			adapterLog.Warn("legacy agent remains alongside canonical agent; leaving both for manual review", "adapter", "claude-code", "file", file, "legacy_path", legacyPath, "canonical_path", newPath)
			continue
		}
		if err := os.Rename(legacyPath, newPath); err != nil {
			adapterLog.Error("failed to migrate legacy agent", "adapter", "claude-code", "from", legacyPath, "to", newPath, "error", err)
			continue
		}
		adapterLog.Info("migrated legacy global agent", "adapter", "claude-code", "from", legacyPath, "to", newPath)
	}
}

func (a *ClaudeCodeAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdClaudeCode, ctx)
}

func (a *ClaudeCodeAdapter) CanRunHeadless() bool { return true }

func (a *ClaudeCodeAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error {
	_, err := exec.LookPath("claude")
	if err != nil {
		adapterLog.Info("claude not on PATH, skipping headless init", "adapter", "claude-code")
		return nil
	}

	adapterLog.Info("running headless init", "adapter", "claude-code", "command", "claude -p")
	execCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "claude",
		"-p", prompt,
		"--dangerously-skip-permissions",
		"--max-budget-usd", "1",
		"--output-format", "json",
	)
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil

	output, err := cmd.CombinedOutput()
	if err != nil {
		adapterLog.Warn("headless init completed with warning", "adapter", "claude-code", "error", err)
		if len(output) > 0 {
			adapterLog.Info("headless init output", "adapter", "claude-code", "output", truncateOutput(string(output), 200))
		}
		return nil // non-fatal
	}

	adapterLog.Info("headless init completed", "adapter", "claude-code", "bytes", len(output))
	return nil
}

func (a *ClaudeCodeAdapter) RunHeadlessValidation(ctx *AdapterContext) error {
	_, err := exec.LookPath("claude")
	if err != nil {
		adapterLog.Info("claude not on PATH, skipping headless validation", "adapter", "claude-code")
		return nil
	}

	adapterLog.Info("running headless validation", "adapter", "claude-code", "command", "claude -p", "prompt", "verify setup structure")
	execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "claude", "-p", "verify setup structure")
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil // pipe /dev/null equivalent — no interactive input

	output, err := cmd.CombinedOutput()
	if err != nil {
		adapterLog.Warn("headless validation completed with warning", "adapter", "claude-code", "error", err)
		if len(output) > 0 {
			adapterLog.Info("headless validation output", "adapter", "claude-code", "output", string(output))
		}
		return nil // non-fatal
	}

	adapterLog.Info("headless validation succeeded", "adapter", "claude-code")
	return nil
}

// displayInstallSummary prints a post-install summary of registered tools.
// Non-fatal on failure — it's informational only.
func displayInstallSummary(ctx *AdapterContext, claudeDir string, isGlobal bool) {
	// Only attempt if claude is on PATH.
	_, found := LookupClaudeBinary()
	if !found {
		return
	}

	// Count installed artifacts by walking directories.
	agents := countDirEntries(filepath.Join(claudeDir, "agents"))
	skills := countDirEntries(filepath.Join(claudeDir, "skills"))
	commands := countDirEntries(filepath.Join(claudeDir, "commands"))
	styles := countDirEntries(filepath.Join(claudeDir, "output-styles"))

	// Try to get MCP server count from settings.json or direct CLI query.
	mcpCount := 0
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if settingsData, err := files.ReadFile(settingsPath); err == nil {
		// Simple heuristic: count occurrences of "mcp" key as estimate.
		mcpCount = strings.Count(string(settingsData), `"mcp"`)
		// More accurate: try to parse and count mcpServers entries.
		var parsed map[string]any
		if err := json.Unmarshal(settingsData, &parsed); err == nil {
			if mcpServersRaw, ok := parsed["mcp"].(map[string]any); ok {
				mcpCount = len(mcpServersRaw)
			}
		}
	}

	// Format scope label
	scopeLabel := "project"
	if isGlobal {
		scopeLabel = "user"
	}

	// Emit summary
	adapterLog.Info("install summary", "adapter", "claude-code", "scope", scopeLabel)
	if mcpCount > 0 {
		adapterLog.Info("mcp servers registered", "adapter", "claude-code", "scope", scopeLabel, "count", mcpCount)
	}
	if agents > 0 {
		adapterLog.Info("agents available", "adapter", "claude-code", "scope", scopeLabel, "count", agents)
	}
	if skills > 0 {
		adapterLog.Info("skills available", "adapter", "claude-code", "scope", scopeLabel, "count", skills)
	}
	if commands > 0 {
		adapterLog.Info("commands available", "adapter", "claude-code", "scope", scopeLabel, "count", commands)
	}
	if styles > 0 {
		adapterLog.Info("output styles available", "adapter", "claude-code", "scope", scopeLabel, "count", styles)
	}
}

// countDirEntries returns the count of files in a directory (non-recursive).
// Returns 0 if dir doesn't exist or can't be read.
func countDirEntries(dirPath string) int {
	entries := files.ListDir(dirPath)
	count := 0
	for _, entry := range entries {
		// Count files, not subdirectories, for most counts.
		// But for skills, each skill is a directory, so include those.
		fullPath := filepath.Join(dirPath, entry)
		if strings.HasSuffix(dirPath, "skills") {
			// Skills are directories; count any entry.
			if !files.IsDirectory(fullPath) {
				continue
			}
		}
		// For others (agents, commands, output-styles), count files.
		if files.IsDirectory(fullPath) {
			continue
		}
		count++
	}
	return count
}
