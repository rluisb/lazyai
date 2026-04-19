// Package adapter provides the Claude Code adapter implementation.
// Ported from the TypeScript ClaudeCodeAdapter.
package adapter

import (
	"context"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	sampleRulePath := filepath.Join(rulesDir, "typescript.md")

	_ = files.EnsureDir(claudeDir)
	_ = files.EnsureDir(rulesDir)
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
	}
	preExisted := files.FileExists(settingsPath)
	if _, err := configmerge.MergeJSONFile(settingsPath, defaultSettings); err != nil {
		return nil, err
	}
	if !preExisted {
		relPath, _ := filepath.Rel(ctx.TargetDir, settingsPath)
		hash, _ := files.FileHash(settingsPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	// Write sample rule if it doesn't exist.
	if !files.FileExists(sampleRulePath) {
		content := "---\npaths:\n  - \"src/**/*.ts\"\n---\n\n# TypeScript Rules\n\n- Use strict TypeScript\n- Prefer interfaces over types for objects\n"
		if err := files.WriteFile(sampleRulePath, []byte(content), 0o644); err != nil {
			return nil, err
		}
		relPath, _ := filepath.Rel(ctx.TargetDir, sampleRulePath)
		hash, _ := files.FileHash(sampleRulePath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	log.Println("Installing Claude Code tools...")

	// Copy agents.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(claudeDir, "agents", file)
		},
		IncludeFile: func(file string) bool {
			return fileID(file) != "orchestrator"
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator agent if enabled.
	if !isGlobal && IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorAgentContent(ctx)
		if err := CopyWithRecord("agents/orchestrator.md",
			filepath.Join(claudeDir, "agents", "orchestrator.md"),
			ctx, false,
			func([]byte) []byte { return content },
		); err != nil {
			return nil, err
		}
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

	// Install tool context files. At global scope ~/.claude/CLAUDE.md is the
	// user's personal-conventions file; preserve any pre-existing copy
	// (Spec 012 / Q3 decision).
	if err := InstallToolContextFiles(InstallToolContextFilesOption{
		Ctx:                ctx,
		ToolDir:            claudeDir,
		ContextFileName:    "CLAUDE.md",
		AgentsDestDir:      "agents",
		SkillsDestDir:      "skills",
		SkipRootIfExists:   isGlobal,
	}); err != nil {
		return nil, err
	}

	// Root CLAUDE.md placement is handled centrally by scaffold/root.go
	// (scope-aware); the adapter no longer installs it.

	// CLI-driven MCP registration: when DriveCLI=true, attempt to register MCP
	// servers via `claude mcp add`. Falls back silently to direct-write (via
	// settings.json merge) on any failure.
	if ctx.DriveCLI {
		if ok := installClaudeMCPViaCLI(ctx, claudeDir); ok {
			log.Println("[claude] MCP servers registered via CLI")
		}
	}

	return ctx.FileRecords, nil
}

// installClaudeMCPViaCLI calls `claude mcp add` for each enabled MCP server
// found in the canonical .ai/mcp.json. Returns true only if at least one
// server was registered successfully. Safe to ignore the return value —
// settings.json already provides the direct-write baseline.
func installClaudeMCPViaCLI(ctx *AdapterContext, claudeDir string) bool {
	bin, err := exec.LookPath("claude")
	if err != nil {
		log.Println("[claude] --drive-cli requested but claude binary not found; using direct-write")
		return false
	}

	catalog := readCanonicalMcp(ctx.TargetDir)
	if catalog == nil || len(catalog.Servers) == 0 {
		return false
	}

	success := false
	for name, srv := range catalog.Servers {
		if !srv.isEnabled() {
			continue
		}
		// claude mcp add --name <name> --command <cmd> [args...] [--env KEY=VAL]
		args := []string{"mcp", "add", "--name", name}
		if srv.Command != "" {
			args = append(args, "--command", srv.Command)
		}
		for _, a := range srv.Args {
			args = append(args, "--args", a)
		}
		for k, v := range srv.Env {
			args = append(args, "--env", k+"="+v)
		}

		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		out, runErr := exec.CommandContext(runCtx, bin, args...).CombinedOutput()
		cancel()
		if runErr != nil {
			log.Printf("[claude] mcp add %q failed: %v\n%s", name, runErr, string(out))
			continue
		}
		log.Printf("[claude] registered MCP server %q via CLI", name)
		success = true
	}
	_ = claudeDir
	return success
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
		entries, err := fs.ReadDir(ctx.LibraryFS, "agents")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				agentFiles = append(agentFiles, e.Name())
			}
		}
	} else if ctx.LibraryDir != "" {
		sourceDir := filepath.Join(ctx.LibraryDir, "agents")
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
			log.Printf("[claude-code] legacy agent %q remains at %s (canonical %s also exists; leaving both for manual review)", file, legacyPath, newPath)
			continue
		}
		if err := os.Rename(legacyPath, newPath); err != nil {
			log.Printf("[claude-code] failed to migrate legacy agent %s -> %s: %v", legacyPath, newPath, err)
			continue
		}
		log.Printf("[claude-code] migrated legacy global agent %s -> %s", legacyPath, newPath)
	}
}

func (a *ClaudeCodeAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdClaudeCode, ctx)
}

func (a *ClaudeCodeAdapter) CanRunHeadless() bool { return true }

func (a *ClaudeCodeAdapter) RunHeadlessValidation(ctx *AdapterContext) error {
	_, err := exec.LookPath("claude")
	if err != nil {
		log.Printf("[claude-code] claude not on PATH, skipping headless validation")
		return nil
	}

	log.Printf("[claude-code] running headless validation: claude -p \"verify setup structure\"")
	execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "claude", "-p", "verify setup structure")
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil // pipe /dev/null equivalent — no interactive input

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[claude-code] headless validation completed with warning: %v", err)
		if len(output) > 0 {
			log.Printf("[claude-code] output: %s", string(output))
		}
		return nil // non-fatal
	}

	log.Printf("[claude-code] headless validation succeeded")
	return nil
}
