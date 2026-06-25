// Package adapter provides the GitHub Copilot adapter implementation.
// Ported from the TypeScript CopilotAdapter.
package adapter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	hooksDir := filepath.Join(githubDir, "hooks")
	if ctx.SetupScope != types.SetupScopeGlobal {
		_ = files.EnsureDir(hooksDir)
	}

	adapterLog.Info("installing tools", "adapter", "copilot")

	selectedSkills := selectionSet(ctx.Selections.Skills)
	selectedPrompts := selectionSet(ctx.Selections.Prompts)

	// Copy canonical agents from library to .github/agents/.
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

	skillsDir := filepath.Join(githubDir, "skills")
	_ = files.EnsureDir(skillsDir)

	// Copy selected skills to Agent Skills directories.
	if err := a.cleanupLegacySkillAgentOutputs(ctx, agentsDir, selectedSkills); err != nil {
		return nil, err
	}
	if len(selectedSkills) > 0 {
		if err := a.copySkillsAsSkillDirs(ctx, skillsDir, selectedSkills); err != nil {
			return nil, err
		}
	}

	// Copy Copilot project/workspace hooks. Global Copilot support remains
	// agent/instructions/MCP oriented; no verified user-scope hook surface is
	// emitted there.
	if ctx.SetupScope != types.SetupScopeGlobal {
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "copilot/hooks",
			IncludeFile:  func(file string) bool { return strings.HasSuffix(file, ".json") },
			ToDestPath: func(file string) string {
				return filepath.Join(hooksDir, file)
			},
		}); err != nil {
			return nil, err
		}
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "copilot/hooks",
			IncludeFile:  func(file string) bool { return strings.HasSuffix(file, ".sh") },
			ToDestPath: func(file string) string {
				return filepath.Join(hooksDir, file)
			},
			Mode: 0o755,
		}); err != nil {
			return nil, err
		}
	}
	// Build the set of canonical agent basenames already emitted to agentsDir so
	// that chatmode-derived files never overwrite them. Canonical agents are
	// authoritative: if a chatmode shares a name with a canonical agent, the
	// chatmode copy is silently skipped.
	canonicalAgentNames := map[string]bool{}
	if canonicalEntries, rdErr := os.ReadDir(agentsDir); rdErr == nil {
		for _, e := range canonicalEntries {
			if !e.IsDir() {
				canonicalAgentNames[e.Name()] = true
			}
		}
	}
	// Copy custom agents (formerly chat modes) to .github/agents/<name>.agent.md,
	// skipping any whose destination basename collides with a canonical agent.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "chatmodes",
		SelectionKey: "chatmodes",
		ToDestPath: func(file string) string {
			return filepath.Join(agentsDir, filepath.Base(file))
		},
		IncludeFile: func(file string) bool {
			return !canonicalAgentNames[filepath.Base(file)]
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
	relPath = filepath.ToSlash(relPath)

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

// copyCopilotAgents renders canonical agent markdown into Copilot .agent.md files.
func (a *CopilotAdapter) copyCopilotAgents(ctx *AdapterContext, agentsDir string) error {
	if err := CopyWithRecord(
		filepath.ToSlash(filepath.Join("canonical", "agents", defaultAgentID+".md")),
		filepath.Join(agentsDir, defaultAgentID+".agent.md"),
		ctx,
		true,
		copilotAgentMarkdownContent,
		0o644,
	); err != nil {
		return err
	}

	if ctx.LibraryFS != nil {
		entries, err := fs.ReadDir(ctx.LibraryFS, "canonical/agents")
		if err != nil {
			return nil
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			file := entry.Name()
			if !isCanonicalAgentFile(file) || isDefaultAgentFile(file) {
				continue
			}
			src := filepath.ToSlash(filepath.Join("canonical", "agents", file))
			dest := filepath.Join(agentsDir, fileID(file)+".agent.md")
			if err := CopyWithRecord(src, dest, ctx, true, copilotAgentMarkdownContent, 0o644); err != nil {
				return err
			}
		}
		return nil
	}

	sourceDir := filepath.Join(ctx.LibraryDir, "canonical", "agents")
	if !files.DirExists(sourceDir) {
		return nil
	}
	for _, file := range files.ListDir(sourceDir) {
		srcPath := filepath.Join(sourceDir, file)
		if files.IsDirectory(srcPath) || !isCanonicalAgentFile(file) || isDefaultAgentFile(file) {
			continue
		}
		sourcePath := filepath.ToSlash(filepath.Join("canonical", "agents", file))
		dest := filepath.Join(agentsDir, fileID(file)+".agent.md")
		if err := CopyWithRecord(sourcePath, dest, ctx, true, copilotAgentMarkdownContent, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func copilotAgentMarkdownContent(source []byte) []byte {
	fm, body, err := frontmatter.ExtractFrontmatter(source)
	description := ""
	agentName := "agent"
	if err == nil {
		agentName = strings.TrimSpace(frontmatter.ExtractField(fm, "name"))
		if agentName == "" {
			agentName = "agent"
		}
		description = strings.TrimSpace(frontmatter.ExtractField(fm, "description"))
	}
	if description == "" {
		description = "Agent for the LazyAI runtime."
	}

	bodyStr := strings.TrimLeft(string(body), "\n")
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: ")
	b.WriteString(agentName)
	b.WriteString("\ndescription: ")
	b.WriteString(strconv.Quote(description))
	b.WriteString("\ntools: [\"read\", \"search\", \"edit\", \"shell\"]\n---\n\n")
	b.WriteString(managedAgentMarker("copilot", agentName))
	b.WriteString("\n\n")
	b.WriteString(bodyStr)
	if !strings.HasSuffix(bodyStr, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return []byte(b.String())
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

// copySkillsAsSkillDirs copies selected skill files to Copilot Agent Skills
// directories.
func (a *CopilotAdapter) copySkillsAsSkillDirs(ctx *AdapterContext, skillsDir string, selected map[types.SkillId]bool) error {
	if len(selected) == 0 {
		return nil
	}
	return CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(skillsDir, name, "SKILL.md")
		},
	})
}

// cleanupLegacySkillAgentOutputs removes old, selected Copilot skill outputs that
// used the deprecated skill-as-agent behavior.
func (a *CopilotAdapter) cleanupLegacySkillAgentOutputs(ctx *AdapterContext, agentsDir string, selected map[types.SkillId]bool) error {
	if len(selected) == 0 {
		return nil
	}
	for skill := range selected {
		fileBase := string(skill)
		legacyMarkdown := filepath.Join(agentsDir, fileBase+".agent.md")
		legacyYaml := filepath.Join(agentsDir, fileBase+".agent.yaml")
		if err := a.cleanupLegacySkillAgentArtifact(ctx, legacyMarkdown, fileBase, ".agent.md"); err != nil {
			return err
		}
		if err := a.cleanupLegacySkillAgentArtifact(ctx, legacyYaml, fileBase, ".agent.yaml"); err != nil {
			return err
		}
	}
	return nil
}

func (a *CopilotAdapter) cleanupLegacySkillAgentArtifact(ctx *AdapterContext, path string, skillID string, ext string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !a.isLikelyLegacySkillAgentFile(string(data), skillID, ctx, ext) {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (a *CopilotAdapter) isLikelyLegacySkillAgentFile(content string, skillID string, ctx *AdapterContext, ext string) bool {
	if strings.Contains(content, managedAgentMarker("copilot", skillID)) {
		return true
	}
	sourceContent, ok := a.readLibrarySkillSource(ctx, skillID)
	switch ext {
	case ".agent.md":
		if ok {
			expected, err := skillToCopilotAgentMarkdown(ctx, filepath.ToSlash(filepath.Join("skills", skillID+".md")), string(sourceContent))
			if err == nil {
				return strings.TrimSpace(content) == strings.TrimSpace(expected)
			}
		}
		return false
	case ".agent.yaml":
		if !strings.Contains(content, "name: "+skillID) {
			return false
		}
		if !strings.Contains(content, "model:") || !strings.Contains(content, "prompt:") {
			return false
		}
		if !ok {
			return true
		}
		// Best-effort check: legacy YAML embeds the skill body as an indented prompt.
		_, sourceBody, err := frontmatter.ExtractFrontmatter(sourceContent)
		heading := "# " + skillID
		return err == nil && strings.Contains(string(sourceBody), heading) && strings.Contains(content, heading)
	default:
		return false
	}
}

func (a *CopilotAdapter) readLibrarySkillSource(ctx *AdapterContext, skillID string) ([]byte, bool) {
	path := filepath.ToSlash(filepath.Join("skills", skillID+".md"))
	if ctx.LibraryFS != nil {
		data, err := fs.ReadFile(ctx.LibraryFS, path)
		if err == nil {
			return data, true
		}
		return nil, false
	}
	if ctx.LibraryDir == "" {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(ctx.LibraryDir, "skills", skillID+".md"))
	if err != nil {
		return nil, false
	}
	return data, true
}

// skillToCopilotAgentMarkdown transforms a skill markdown file into Copilot
// custom-agent Markdown format. The model field is resolved against
// `CopilotCatalog` based on the skill's optional tier annotation (defaults to
// Balanced if absent), replacing the hardcoded `model: claude-sonnet-*` value
// that previously required a manual bulk-edit on every Anthropic version bump
// (#199 Bug 2 long-term fix).
//
// Skills opt into Frontier or Speed tier by adding `tier: frontier` (and
// optional `risk:`/`temperature:`/`thinking:`) to their frontmatter. Skills
// with no tier annotation get Balanced — which currently resolves to
// `claude-sonnet-4.6` via `CopilotCatalog.Balanced[0]`.
func skillToCopilotAgentMarkdown(ctx *AdapterContext, skillName string, skillContent string) (string, error) {
	fm, body, err := frontmatter.ParseYamlFrontmatter(skillContent)
	if err != nil {
		return "", fmt.Errorf("parse skill frontmatter %s: %w", skillName, err)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return "", fmt.Errorf("skill %s has no content", skillName)
	}

	// Extract skill ID from filename (e.g., "skills/review.md" → "review").
	basename := filepath.Base(skillName)
	skillID := strings.TrimSuffix(basename, filepath.Ext(basename))

	description := frontmatter.ExtractField(fm, "description")
	if strings.TrimSpace(description) == "" {
		description = fmt.Sprintf("%s skill for the LazyAI runtime.", skillID)
	}

	spec := skillSpecOrDefault([]byte(skillContent), skillID)
	rc := resolveCtxFor(types.ToolIdCopilot, ctx)
	res, err := models.Resolve(spec, rc)
	if err != nil {
		return "", fmt.Errorf("copilot skill %s resolve: %w", skillID, err)
	}

	return fmt.Sprintf(`---
name: %s
description: %s
model: %s
tools:
  - "*"
---

%s
`, skillID, yamlBlockScalar(description, "  "), res.Field, body), nil
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

func yamlBlockScalar(text, indent string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "|-\n" + indent + "LazyAI skill."
	}
	return "|-\n" + indentLines(text, indent)
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

	adapterLog.Warn("Running headless init with elevated permissions. The spawned AI agent will have unrestricted access to files and commands for up to 120 seconds.")
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
