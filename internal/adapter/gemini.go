// Package adapter provides the Gemini adapter implementation.
// Ported from the TypeScript GeminiAdapter.
package adapter

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// GeminiAdapter implements ToolAdapter for Gemini CLI.
type GeminiAdapter struct{}

func (a *GeminiAdapter) ID() types.ToolId  { return types.ToolIdGemini }
func (a *GeminiAdapter) Name() string      { return "Gemini CLI" }
func (a *GeminiAdapter) ConfigDir() string { return ".gemini" }

func (a *GeminiAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	geminiDir, err := ResolveToolRoot(types.ToolIdGemini, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}
	_ = files.EnsureDir(geminiDir)
	_ = files.EnsureDir(filepath.Join(geminiDir, "skills"))

	// Merge default settings.json (preserves user-authored keys like mcpServers, hooks).
	settingsPath := filepath.Join(geminiDir, "settings.json")
	defaultSettings := map[string]any{
		"general": map[string]any{
			"defaultApprovalMode": "default",
		},
		"model": map[string]any{
			"name": "gemini-2.5-pro",
		},
		"context": map[string]any{
			"fileName":             "GEMINI.md",
			"includeDirectoryTree": true,
		},
	}
	preExisted := files.FileExists(settingsPath)
	if _, err := configmerge.MergeJSONFile(settingsPath, defaultSettings); err != nil {
		return nil, err
	}
	if !preExisted {
		relPath, _ := filepath.Rel(ctx.TargetDir, settingsPath)
		if relPath == "" || relPath == "." {
			relPath = settingsPath
		}
		hash, _ := files.FileHash(settingsPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	log.Println("Installing Gemini CLI tools...")

	// Gemini has no agents concept — skip agents entirely.
	// Copy skills.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(geminiDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator as a skill if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorSkillContent(ctx)
		if err := WriteContentWithRecord(
			filepath.Join(geminiDir, "skills", "orchestrator", "SKILL.md"),
			content, ctx, "generated:orchestrator-skill", false,
		); err != nil {
			return nil, err
		}
	}

	// Root GEMINI.md placement is handled centrally by scaffold/root.go.

	// CLI-driven MCP registration: when DriveCLI=true, attempt to register MCP
	// servers via `gemini mcp add`. Falls back to direct-write on any failure.
	if ctx.DriveCLI {
		if ok := installGeminiMCPViaCLI(ctx, geminiDir); ok {
			log.Println("[gemini] MCP servers registered via CLI")
		}
		// On failure (binary absent or command error) we already wrote settings.json
		// above, so no further action is needed.
	}

	return ctx.FileRecords, nil
}

// installGeminiMCPViaCLI calls `gemini mcp add` for each enabled MCP server
// found in the canonical .ai/mcp.json. Returns true only if at least one
// server was registered successfully. Safe to ignore the return value — caller
// already wrote settings.json as the direct-write baseline.
func installGeminiMCPViaCLI(ctx *AdapterContext, geminiDir string) bool {
	geminiBin, err := exec.LookPath("gemini")
	if err != nil {
		log.Println("[gemini] --drive-cli requested but gemini binary not found; using direct-write")
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
		args := []string{"mcp", "add", "--server-name", name}
		if srv.Command != "" {
			args = append(args, "--command", srv.Command)
		}
		for _, a := range srv.Args {
			args = append(args, "--args", a)
		}
		if len(srv.Env) > 0 {
			envJSON, _ := json.Marshal(srv.Env)
			args = append(args, "--env", string(envJSON))
		}

		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		out, runErr := exec.CommandContext(runCtx, geminiBin, args...).CombinedOutput()
		cancel()
		if runErr != nil {
			log.Printf("[gemini] mcp add %q failed: %v\n%s", name, runErr, string(out))
			continue
		}
		log.Printf("[gemini] registered MCP server %q via CLI", name)
		success = true
	}
	_ = geminiDir
	return success
}

func (a *GeminiAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdGemini, ctx)
}

func (a *GeminiAdapter) CanRunHeadless() bool { return false }

func (a *GeminiAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }
