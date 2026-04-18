// Package adapter provides the Codex adapter implementation.
// Ported from the TypeScript CodexAdapter.
package adapter

import (
	"context"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/detect"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// CodexAdapter implements ToolAdapter for OpenAI Codex CLI.
//
// Codex uses two roots per upstream convention:
//   - configRoot (.codex/ or ~/.codex/): holds config.toml with
//     [mcp_servers.*] tables.
//   - skillsRoot (.agents/skills/ or ~/.agents/skills/): holds per-skill
//     <name>/SKILL.md. Skills live outside .codex/ in Codex's model.
//
// The root AGENTS.md is placed by scaffold/root.go (scope-aware).
type CodexAdapter struct{}

func (a *CodexAdapter) ID() types.ToolId  { return types.ToolIdCodex }
func (a *CodexAdapter) Name() string      { return "Codex CLI" }
func (a *CodexAdapter) ConfigDir() string { return ".codex" }

func (a *CodexAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	// Check if Codex is installed; print a non-fatal warning if not.
	_ = detect.EnsureCodexOrPrompt()

	configRoot, skillsRoot, err := ResolveCodexRoots(ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(configRoot)
	_ = files.EnsureDir(skillsRoot)

	log.Println("Installing Codex tools...")

	// Emit a minimal config.toml with an empty [mcp_servers] table so Codex
	// recognises the project/global config as trusted. User-authored tables
	// survive via configmerge.
	configPath := filepath.Join(configRoot, "config.toml")
	configPatch := map[string]any{
		"mcp_servers": map[string]any{},
	}
	preExisted := files.FileExists(configPath)
	if _, err := configmerge.MergeTOMLFile(configPath, configPatch); err != nil {
		return nil, err
	}
	if !preExisted {
		relPath, _ := filepath.Rel(ctx.TargetDir, configPath)
		if relPath == "" || relPath == "." {
			relPath = configPath
		}
		hash, _ := files.FileHash(configPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	// Codex uses skills in directory format at the skills root.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(skillsRoot, name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator as a skill if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorSkillContent(ctx)
		if err := WriteContentWithRecord(
			filepath.Join(skillsRoot, "orchestrator", "SKILL.md"),
			content, ctx, "generated:orchestrator-skill", false,
		); err != nil {
			return nil, err
		}
	}

	// Root AGENTS.md placement is handled centrally by scaffold/root.go.

	return ctx.FileRecords, nil
}

func (a *CodexAdapter) CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCodex, targetDir, fileRecords)
}

func (a *CodexAdapter) CanRunHeadless() bool { return true }

func (a *CodexAdapter) RunHeadlessValidation(ctx *AdapterContext) error {
	_, err := exec.LookPath("codex")
	if err != nil {
		log.Printf("[codex] codex not on PATH, skipping headless validation")
		return nil
	}

	log.Printf("[codex] running headless validation: codex exec \"check .agents/ structure\"")
	execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "codex", "exec", "check .agents/ structure")
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil // pipe /dev/null equivalent — no interactive input

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[codex] headless validation completed with warning: %v", err)
		if len(output) > 0 {
			log.Printf("[codex] output: %s", string(output))
		}
		return nil // non-fatal
	}

	log.Printf("[codex] headless validation succeeded")
	return nil
}
