// Package adapter provides the Codex adapter implementation.
// Ported from the TypeScript CodexAdapter.
package adapter

import (
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// CodexAdapter implements ToolAdapter for OpenAI Codex CLI.
//
// Structure:
//   - Root: AGENTS.md (shared project notes)
//   - Skills: .agents/skills/{name}/SKILL.md (AgentSkills standard)
//   - No .codex/ project directory (config only in ~/.codex/)
//   - Agents: Inline in AGENTS.md (no separate directory)
type CodexAdapter struct{}

func (a *CodexAdapter) ID() types.ToolId  { return types.ToolIdCodex }
func (a *CodexAdapter) Name() string      { return "Codex CLI" }
func (a *CodexAdapter) ConfigDir() string { return ".agents" }

func (a *CodexAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	isGlobal := ctx.SetupScope == types.SetupScopeGlobal
	agentsDir := filepath.Join(ctx.TargetDir, ".agents")
	if isGlobal {
		agentsDir = filepath.Join(filepath.Dir(ctx.TargetDir), ".agents")
	}

	_ = files.EnsureDir(agentsDir)
	_ = files.EnsureDir(filepath.Join(agentsDir, "skills"))

	log.Println("Installing Codex tools...")

	// Codex uses skills in directory format.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(agentsDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator as a skill if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorSkillContent(ctx)
		if err := WriteContentWithRecord(
			filepath.Join(agentsDir, "skills", "orchestrator", "SKILL.md"),
			content, ctx, "generated:orchestrator-skill", false,
		); err != nil {
			return nil, err
		}
	}

	// Install context files (AGENTS.md references agents inline).
	if err := InstallToolContextFiles(InstallToolContextFilesOption{
		Ctx:             ctx,
		ToolDir:         agentsDir,
		ContextFileName: "AGENTS.md",
		AgentsDestDir:   ".", // Inline - agents referenced in root file.
		SkillsDestDir:   "skills",
	}); err != nil {
		return nil, err
	}

	// Install root AGENTS.md template.
	if err := InstallRootTemplateIfMissing(ctx, "AGENTS.md",
		filepath.Join(ctx.TargetDir, "AGENTS.md"),
		"root/AGENTS.template.md"); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *CodexAdapter) CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	// Codex has no MCP config file format — nothing to compile.
	return fileRecords, nil
}
