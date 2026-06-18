// Package adapter provides the GitHub Copilot adapter implementation.
// Ported from the TypeScript CopilotAdapter.
package adapter

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/models"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// CopilotAdapter implements ToolAdapter for GitHub Copilot.
type CopilotAdapter struct{}

func (a *CopilotAdapter) ID() types.ToolId  { return types.ToolIdCopilot }
func (a *CopilotAdapter) Name() string      { return "GitHub Copilot CLI" }
func (a *CopilotAdapter) ConfigDir() string { return ".github" }

func (a *CopilotAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdCopilot, ctx.SetupScope) {
		adapterLog.Info("scope not supported; skipping install", "adapter", "copilot", "scope", ctx.SetupScope)
		return ctx.FileRecords, nil
	}
	// At global scope, check if copilot CLI or ~/.copilot/ is present.
	if ctx.SetupScope == types.SetupScopeGlobal {
		_, found := LookupCopilotBinary()
		if !found {
			// Check if ~/.copilot/ exists in the context's home directory
			home, err := resolveHomeDir(ctx)
			if err != nil || !files.DirExists(filepath.Join(home, ".copilot")) {
				adapterLog.Info("copilot CLI or ~/.copilot/ not found; skipping global install", "adapter", "copilot")
				return ctx.FileRecords, nil
			}
		}
	}
	githubDir, err := ResolveToolRoot(types.ToolIdCopilot, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}
	_ = files.EnsureDir(githubDir)

	agentsDir := filepath.Join(githubDir, "agents")
	_ = files.EnsureDir(agentsDir)
	instructionsDir := filepath.Join(githubDir, "instructions")
	_ = files.EnsureDir(instructionsDir)
	promptsDir := filepath.Join(githubDir, "prompts")
	_ = files.EnsureDir(promptsDir)

	adapterLog.Info("installing tools", "adapter", "copilot")

	selectedSkills := selectionSet(ctx.Selections.Skills)
	selectedPrompts := selectionSet(ctx.Selections.Prompts)

	// Copy agents from library to .github/agents/.
	if err := a.copyCopilotAgents(ctx, agentsDir); err != nil {
		return nil, err
	}

	// Copy instructions from library to .github/instructions/.
	if err := a.copyCopilotInstructions(ctx, instructionsDir); err != nil {
		return nil, err
	}

	// Copy prompts as Copilot .prompt.md files.
	if err := a.copyLibrarySubdirAsPrompts(ctx, "prompts", selectedPrompts, promptsDir); err != nil {
		return nil, err
	}

	// Copy skills as Copilot agent files.
	if err := a.copySkillsAsAgents(ctx, agentsDir, selectedSkills); err != nil {
		return nil, err
	}

	// Copy chat modes to .github/chatmodes/<name>.chatmode.md.
	chatmodesDir := filepath.Join(githubDir, "chatmodes")
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "chatmodes",
		SelectionKey: "chatmodes",
		ToDestPath: func(file string) string {
			return filepath.Join(chatmodesDir, filepath.Base(file))
		},
	}); err != nil {
		return nil, err
	}

	// Memory docs (AGENTS.md and .github/copilot-instructions.md) are placed
	// by scaffold/root.go, which knows the scope-correct destination.

	return ctx.FileRecords, nil
}

// copyLibrarySubdirAsPrompts copies files from a library subdirectory as prompt files.
func (a *CopilotAdapter) copyLibrarySubdirAsPrompts(ctx *AdapterContext, subdir string, selected map[types.PromptId]bool, promptsDir string) error {
	libFS := ctx.LibraryFS
	if libFS != nil {
		return a.copySubdirAsPromptsFromFS(ctx, libFS, subdir, selected, promptsDir)
	}
	// Fallback: disk mode
	srcDir := filepath.Join(ctx.LibraryDir, subdir)
	if !files.DirExists(srcDir) {
		return nil
	}
	for _, file := range files.ListDir(srcDir) {
		srcPath := filepath.Join(srcDir, file)
		if files.IsDirectory(srcPath) {
			continue
		}
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.PromptId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".prompt.md"
		if err := a.copyFileWithRecord(srcPath, filepath.Join(promptsDir, destFile), ctx, subdir+"/"+file); err != nil {
			return err
		}
	}
	return nil
}

// copySubdirAsPromptsFromFS copies files from the library FS as prompt files.
func (a *CopilotAdapter) copySubdirAsPromptsFromFS(ctx *AdapterContext, libFS fs.FS, subdir string, selected map[types.PromptId]bool, promptsDir string) error {
	entries, err := fs.ReadDir(libFS, subdir)
	if err != nil {
		return nil // directory doesn't exist
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := entry.Name()
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.PromptId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".prompt.md"
		dest := filepath.Join(promptsDir, destFile)
		if err := CopyWithRecord(subdir+"/"+file, dest, ctx, false, nil, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// copyFileWithRecord copies a file from disk with conflict resolution and tracking.
// src is an absolute filesystem path; sourcePath is the library-relative path for tracking.
func (a *CopilotAdapter) copyFileWithRecord(src, dest string, ctx *AdapterContext, sourcePath string) error {
	relPath, _ := filepath.Rel(ctx.TargetDir, dest)

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[dest]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(dest, relPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return err
		}
	}

	if err := files.CopyFile(src, dest); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: sourcePath, Owner: types.FileOwnerLibrary,
	})
	return nil
}

// copyCopilotAgents copies agent files from library/copilot/agents/ to the
// target agents directory, rewriting each yaml's model: line based on the
// tier declared in the source-of-truth library/agents/<name>.md. Agents
// without a matching .md (Copilot-only authoring) keep their hand-pinned
// model.
func (a *CopilotAdapter) copyCopilotAgents(ctx *AdapterContext, agentsDir string) error {
	return CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "copilot/agents",
		ToDestPath: func(file string) string {
			return filepath.Join(agentsDir, filepath.Base(file))
		},
		Transform: func(content []byte) []byte {
			out, err := RewriteCopilotAgent(content, ctx)
			if err != nil {
				adapterLog.Warn("copilot agent model rewrite fell back to verbatim copy", "adapter", "copilot", "error", err)
				return content
			}
			return out
		},
	})
}

// copyCopilotInstructions copies instruction files from library/copilot/instructions/ to the target instructions directory.
func (a *CopilotAdapter) copyCopilotInstructions(ctx *AdapterContext, instructionsDir string) error {
	return CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "copilot/instructions",
		ToDestPath: func(file string) string {
			return filepath.Join(instructionsDir, filepath.Base(file))
		},
	})
}

// copySkillsAsAgents converts skill files to .agent.yaml format in the agents directory.
func (a *CopilotAdapter) copySkillsAsAgents(ctx *AdapterContext, agentsDir string, selected map[types.SkillId]bool) error {
	libFS := ctx.LibraryFS
	if libFS != nil {
		return a.copySkillsAsAgentsFromFS(ctx, libFS, agentsDir, selected)
	}
	// Fallback: disk mode
	srcDir := filepath.Join(ctx.LibraryDir, "skills")
	if !files.DirExists(srcDir) {
		return nil
	}
	for _, file := range files.ListDir(srcDir) {
		srcPath := filepath.Join(srcDir, file)
		if files.IsDirectory(srcPath) {
			continue
		}
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.SkillId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".agent.yaml"
		dest := filepath.Join(agentsDir, destFile)
		libRelPath := "skills/" + file
		if err := a.copySkillAsAgentWithRecord(ctx, srcPath, dest, libRelPath); err != nil {
			return err
		}
	}
	return nil
}

// copySkillsAsAgentsFromFS copies skills from the library FS as agent YAML files.
func (a *CopilotAdapter) copySkillsAsAgentsFromFS(ctx *AdapterContext, libFS fs.FS, agentsDir string, selected map[types.SkillId]bool) error {
	entries, err := fs.ReadDir(libFS, "skills")
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := entry.Name()
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.SkillId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".agent.yaml"
		dest := filepath.Join(agentsDir, destFile)
		libRelPath := "skills/" + file
		if err := a.copySkillAsAgentFromFS(ctx, libFS, libRelPath, dest); err != nil {
			return err
		}
	}
	return nil
}

// copySkillAsAgentWithRecord reads a skill from disk and converts it to agent YAML format.
func (a *CopilotAdapter) copySkillAsAgentWithRecord(ctx *AdapterContext, src, dest string, sourcePath string) error {
	relPath, _ := filepath.Rel(ctx.TargetDir, dest)

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[dest]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(dest, relPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return err
		}
	}

	data, err := files.ReadFile(src)
	if err != nil {
		return err
	}

	// Transform skill to agent YAML format
	transformed, err := skillToAgentYAML(ctx, src, string(data))
	if err != nil {
		return err
	}

	if err := files.EnsureDir(filepath.Dir(dest)); err != nil {
		return err
	}
	if err := files.WriteFile(dest, []byte(transformed), 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: sourcePath, Owner: types.FileOwnerLibrary,
	})
	return nil
}

// copySkillAsAgentFromFS reads a skill from the library FS and converts it to agent YAML.
func (a *CopilotAdapter) copySkillAsAgentFromFS(ctx *AdapterContext, libFS fs.FS, src, dest string) error {
	relPath, _ := filepath.Rel(ctx.TargetDir, dest)

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[dest]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(dest, relPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return err
		}
	}

	data, err := fs.ReadFile(libFS, src)
	if err != nil {
		return fmt.Errorf("read FS %s: %w", src, err)
	}

	// Transform skill to agent YAML format
	transformed, err := skillToAgentYAML(ctx, src, string(data))
	if err != nil {
		return err
	}

	if err := files.EnsureDir(filepath.Dir(dest)); err != nil {
		return err
	}
	if err := files.WriteFile(dest, []byte(transformed), 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: src, Owner: types.FileOwnerLibrary,
	})
	return nil
}

// skillToAgentYAML transforms a skill markdown file into Copilot agent YAML
// format. The model field is resolved against `CopilotCatalog` based on the
// skill's optional tier annotation (defaults to Balanced if absent),
// replacing the hardcoded `model: claude-sonnet-*` value that previously
// required a manual bulk-edit on every Anthropic version bump (#199 Bug 2
// long-term fix).
//
// Skills opt into Frontier or Speed tier by adding `tier: frontier`
// (and optional `risk:`/`temperature:`/`thinking:`) to their frontmatter.
// Skills with no tier annotation get Balanced — which currently resolves
// to `claude-sonnet-4.6` via `CopilotCatalog.Balanced[0]`.
func skillToAgentYAML(ctx *AdapterContext, skillName string, skillContent string) (string, error) {
	_, body := frontmatter.SplitYamlFrontmatter(skillContent)
	body = strings.TrimSpace(body)
	if body == "" {
		return "", fmt.Errorf("skill %s has no content", skillName)
	}

	// Extract skill ID from filename (e.g., "skills/review.md" → "review")
	// First get just the basename in case skillName includes path components
	basename := filepath.Base(skillName)
	skillID := strings.TrimSuffix(basename, filepath.Ext(basename))

	spec := skillSpecOrDefault([]byte(skillContent), skillID)
	rc := resolveCtxFor(types.ToolIdCopilot, ctx)
	res, err := models.Resolve(spec, rc)
	if err != nil {
		return "", fmt.Errorf("copilot skill %s resolve: %w", skillID, err)
	}

	// Build agent YAML
	yaml := fmt.Sprintf(`name: %s
displayName: %s
description: >
  %s skill for the ai-setup orchestrator.
model: %s
tools:
  - "*"
promptParts:
  includeAISafety: true
  includeToolInstructions: true
  includeParallelToolCalling: true
  includeCustomAgentInstructions: false
prompt: |
%s
`, skillID, toDisplayName(skillID), skillID, res.Field, indentLines(body, "  "))

	return yaml, nil
}

// skillSpecOrDefault parses tier metadata from a skill's frontmatter, or
// returns a Balanced default when the skill is unannotated. This is the
// "opt-in tier override" path: skills can opt into Frontier/Speed by
// declaring `tier:` (and optional `risk:`/`temperature:`/`thinking:`) in
// their frontmatter; skills without those annotations default to Balanced
// (which resolves to `claude-sonnet-4.6` via `CopilotCatalog.Balanced[0]`).
//
// Note: `frontmatter.ParseAgentSpec` errors when `tier:` is empty; the
// helper swallows that error and returns the Balanced default — so no
// skill-source migration is required for existing skills.
func skillSpecOrDefault(content []byte, skillID string) models.AgentSpec {
	raw, err := frontmatter.ParseAgentSpec(content)
	if err != nil || raw.Tier == "" {
		return models.AgentSpec{
			Name:        skillID,
			Tier:        models.TierBalanced,
			Temperature: 0.1,
			Thinking:    models.ThinkingLow,
			Risk:        3,
		}
	}
	return models.AgentSpec{
		Name:        raw.Name,
		Tier:        models.Tier(raw.Tier),
		Temperature: raw.Temperature,
		Thinking:    models.Thinking(raw.Thinking),
		Risk:        raw.Risk,
		Multimodal:  raw.Multimodal,
	}
}

// toDisplayName converts "foo-bar" to "Foo Bar".
func toDisplayName(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// indentLines prefixes each line of text with the given indent string.
func indentLines(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" || i < len(lines)-1 {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

func (a *CopilotAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCopilot, ctx)
}

func (a *CopilotAdapter) CanRunHeadless() bool { return true }

func (a *CopilotAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error {
	_, err := exec.LookPath("copilot")
	if err != nil {
		adapterLog.Info("copilot not on PATH, skipping headless init", "adapter", "copilot")
		return nil
	}

	adapterLog.Info("running headless init", "adapter", "copilot", "command", "copilot -p")
	execCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "copilot",
		"-p", prompt,
		"--allow-all",
	)
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil

	output, err := cmd.CombinedOutput()
	if err != nil {
		adapterLog.Warn("headless init completed with warning", "adapter", "copilot", "error", err)
		if len(output) > 0 {
			adapterLog.Info("headless init output", "adapter", "copilot", "output", truncateOutput(string(output), 200))
		}
		return nil // non-fatal
	}

	adapterLog.Info("headless init completed", "adapter", "copilot", "bytes", len(output))
	return nil
}

func (a *CopilotAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }
