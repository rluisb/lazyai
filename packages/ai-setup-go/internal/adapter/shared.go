// Package adapter provides shared helper functions used by all tool adapters.
// Ported from the TypeScript shared.ts utilities.
package adapter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// selectionSet returns a set (map[T]bool) for the given slice, or nil if the
// slice is empty (meaning "install everything").
func selectionSet[T ~string](items []T) map[T]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[T]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}

// CopyWithRecord copies a file from the library FS to dest, records it in
// ctx.FileRecords, and handles conflict resolution. If transform is non-nil,
// it is applied to the file content before writing.
//
// src is a relative path within ctx.LibraryFS (e.g., "agents/builder.md").
func CopyWithRecord(src, dest string, ctx *AdapterContext, warnOnSkip bool, transform func([]byte) []byte) error {
	relPath, err := filepath.Rel(ctx.TargetDir, dest)
	if err != nil {
		relPath = dest
	}

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
		if warnOnSkip {
			log.Printf("Skipping existing file: %s", relPath)
		}
		return nil
	}

	// Backup if replacing
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return fmt.Errorf("backup %s: %w", dest, err)
		}
	}

	// Source path for TrackedFile.Source — use the library-relative path.
	sourcePath := src

	if ctx.DryRun {
		log.Printf("[dry-run] Would create: %s", relPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   "dry-run",
			Source: sourcePath,
			Owner:  types.FileOwnerLibrary,
		})
		return nil
	}

	if err := files.EnsureDir(filepath.Dir(dest)); err != nil {
		return err
	}

	// Read content from the library FS (or disk FS as fallback).
	var content []byte
	libFS := ctx.LibraryFS
	if libFS == nil {
		// Fallback: read from filesystem using LibraryDir.
		absSrc := filepath.Join(ctx.LibraryDir, src)
		data, err := files.ReadFile(absSrc)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		content = data
	} else {
		data, err := fs.ReadFile(libFS, src)
		if err != nil {
			return fmt.Errorf("read FS %s: %w", src, err)
		}
		content = data
	}

	if transform != nil {
		content = transform(content)
	}

	if err := files.WriteFile(dest, content, 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: sourcePath,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

// WriteContentWithRecord writes content to dest and records it in ctx.FileRecords.
func WriteContentWithRecord(dest string, content []byte, ctx *AdapterContext, source string, warnOnSkip bool) error {
	relPath, err := filepath.Rel(ctx.TargetDir, dest)
	if err != nil {
		relPath = dest
	}

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
		if warnOnSkip {
			log.Printf("Skipping existing file: %s", relPath)
		}
		return nil
	}

	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return fmt.Errorf("backup %s: %w", dest, err)
		}
	}

	if ctx.DryRun {
		log.Printf("[dry-run] Would create: %s", relPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   "dry-run",
			Source: source,
			Owner:  types.FileOwnerLibrary,
		})
		return nil
	}

	if err := files.EnsureDir(filepath.Dir(dest)); err != nil {
		return err
	}
	if err := files.WriteFile(dest, content, 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: source,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

// CopyLibraryDirectoryOption holds options for CopyLibraryDirectory.
type CopyLibraryDirectoryOption struct {
	Ctx          *AdapterContext
	SourceSubdir string // subdirectory within library (e.g. "agents", "skills")
	SelectionKey string // "agents", "skills", "prompts", "opencodeCommands", or "opencodeModes"
	ToDestPath   func(file string) string
	WarnOnSkip   bool
	Transform    func(content []byte) []byte
	IncludeFile  func(file string) bool
}

// CopyLibraryDirectory copies all files from the library subdirectory,
// applying selection filtering and conflict resolution.
// Uses ctx.LibraryFS when available, falls back to ctx.LibraryDir + filesystem.
func CopyLibraryDirectory(opts CopyLibraryDirectoryOption) error {
	libFS := opts.Ctx.LibraryFS

	if libFS != nil {
		return copyLibraryDirectoryFromFS(opts, libFS)
	}
	// Fallback: filesystem mode using LibraryDir.
	sourceDir := filepath.Join(opts.Ctx.LibraryDir, opts.SourceSubdir)
	if !files.DirExists(sourceDir) {
		return nil // directory doesn't exist, nothing to copy
	}
	return copyLibraryDirectoryFromDisk(opts, sourceDir)
}

// copyLibraryDirectoryFromFS copies files from the library fs.FS.
func copyLibraryDirectoryFromFS(opts CopyLibraryDirectoryOption, libFS fs.FS) error {
	entries, err := fs.ReadDir(libFS, opts.SourceSubdir)
	if err != nil {
		// Directory doesn't exist in FS, nothing to copy.
		return nil
	}

	selectedAgents := selectionSet(opts.Ctx.Selections.Agents)
	selectedSkills := selectionSet(opts.Ctx.Selections.Skills)
	selectedPrompts := selectionSet(opts.Ctx.Selections.Prompts)
	selectedOpenCodeCommands := selectionSet(opts.Ctx.Selections.OpenCodeCommands)
	selectedOpenCodeModes := selectionSet(opts.Ctx.Selections.OpenCodeModes)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := entry.Name()

		if opts.IncludeFile != nil && !opts.IncludeFile(file) {
			continue
		}

		fileIDVal := strings.TrimSuffix(file, filepath.Ext(file))
		switch opts.SelectionKey {
		case "agents":
			if selectedAgents != nil && !selectedAgents[types.AgentId(fileIDVal)] {
				continue
			}
		case "skills":
			if selectedSkills != nil && !selectedSkills[types.SkillId(fileIDVal)] {
				continue
			}
		case "prompts":
			if selectedPrompts != nil && !selectedPrompts[types.PromptId(fileIDVal)] {
				continue
			}
		case "opencodeCommands":
			if selectedOpenCodeCommands != nil && !selectedOpenCodeCommands[types.OpenCodeCommandId(fileIDVal)] {
				continue
			}
		case "opencodeModes":
			if selectedOpenCodeModes != nil && !selectedOpenCodeModes[types.OpenCodeModeId(fileIDVal)] {
				continue
			}
		}

		srcPath := opts.SourceSubdir + "/" + file
		dest := opts.ToDestPath(file)
		if err := CopyWithRecord(srcPath, dest, opts.Ctx, opts.WarnOnSkip, opts.Transform); err != nil {
			return err
		}
	}
	return nil
}

// copyLibraryDirectoryFromDisk copies files from the real filesystem (dev mode).
func copyLibraryDirectoryFromDisk(opts CopyLibraryDirectoryOption, sourceDir string) error {
	selectedAgents := selectionSet(opts.Ctx.Selections.Agents)
	selectedSkills := selectionSet(opts.Ctx.Selections.Skills)
	selectedPrompts := selectionSet(opts.Ctx.Selections.Prompts)
	selectedOpenCodeCommands := selectionSet(opts.Ctx.Selections.OpenCodeCommands)
	selectedOpenCodeModes := selectionSet(opts.Ctx.Selections.OpenCodeModes)

	for _, file := range files.ListDir(sourceDir) {
		if opts.IncludeFile != nil && !opts.IncludeFile(file) {
			continue
		}

		srcPath := filepath.Join(sourceDir, file)
		if files.IsDirectory(srcPath) {
			continue
		}

		fileIDVal := strings.TrimSuffix(file, filepath.Ext(file))
		switch opts.SelectionKey {
		case "agents":
			if selectedAgents != nil && !selectedAgents[types.AgentId(fileIDVal)] {
				continue
			}
		case "skills":
			if selectedSkills != nil && !selectedSkills[types.SkillId(fileIDVal)] {
				continue
			}
		case "prompts":
			if selectedPrompts != nil && !selectedPrompts[types.PromptId(fileIDVal)] {
				continue
			}
		case "opencodeCommands":
			if selectedOpenCodeCommands != nil && !selectedOpenCodeCommands[types.OpenCodeCommandId(fileIDVal)] {
				continue
			}
		case "opencodeModes":
			if selectedOpenCodeModes != nil && !selectedOpenCodeModes[types.OpenCodeModeId(fileIDVal)] {
				continue
			}
		}

		dest := opts.ToDestPath(file)
		// Use library-relative path for the source so CopyWithRecord reads from FS.
		libRelPath := opts.SourceSubdir + "/" + file
		if err := CopyWithRecord(libRelPath, dest, opts.Ctx, opts.WarnOnSkip, opts.Transform); err != nil {
			return err
		}
	}
	return nil
}

// InstallToolContextFilesOption holds options for InstallToolContextFiles.
type InstallToolContextFilesOption struct {
	Ctx              *AdapterContext
	ToolDir          string
	ContextFileName  string
	AgentsDestDir    string
	SkillsDestDir    string
	TemplatesDestDir string
	WarnOnSkip       bool
	// SkipRootIfExists, when true, suppresses the root context file write if
	// <ToolDir>/<ContextFileName> already exists. Used at user/global scope
	// where that path is the user's personal-conventions file (e.g.
	// ~/.claude/CLAUDE.md) and must not be overwritten on re-run.
	SkipRootIfExists bool
}

// InstallToolContextFiles copies the tool-agents context files (agents-dir.md,
// skills-dir.md, root-dir.md) into the appropriate directories.
func InstallToolContextFiles(opts InstallToolContextFilesOption) error {
	// agents directory context file
	if err := CopyWithRecord(
		"tool-agents/agents-dir.md",
		filepath.Join(opts.ToolDir, opts.AgentsDestDir, opts.ContextFileName),
		opts.Ctx, opts.WarnOnSkip, nil,
	); err != nil {
		return err
	}

	// skills directory context file
	if opts.SkillsDestDir != "" {
		if err := CopyWithRecord(
			"tool-agents/skills-dir.md",
			filepath.Join(opts.ToolDir, opts.SkillsDestDir, opts.ContextFileName),
			opts.Ctx, opts.WarnOnSkip, nil,
		); err != nil {
			return err
		}
	}

	// templates directory context file
	if opts.TemplatesDestDir != "" {
		if err := CopyWithRecord(
			"tool-agents/templates-dir.md",
			filepath.Join(opts.ToolDir, opts.TemplatesDestDir, opts.ContextFileName),
			opts.Ctx, opts.WarnOnSkip, nil,
		); err != nil {
			return err
		}
	}

	// root directory context file
	rootPath := filepath.Join(opts.ToolDir, opts.ContextFileName)
	if opts.SkipRootIfExists && files.FileExists(rootPath) {
		return nil
	}
	return CopyWithRecord(
		"tool-agents/root-dir.md",
		rootPath,
		opts.Ctx, opts.WarnOnSkip, nil,
	)
}

// InstallRootTemplateIfMissing installs a root template file if it hasn't
// already been created during this installation session.
func InstallRootTemplateIfMissing(ctx *AdapterContext, recordPath, destPath, templateSource string) error {
	// Check if already created in this session.
	for _, r := range ctx.FileRecords {
		if r.Path == recordPath {
			return nil
		}
	}

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[destPath]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(destPath, recordPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}

	if action == conflict.ActionReplace && files.FileExists(destPath) {
		if _, err := files.BackupFile(destPath, ctx.TargetDir); err != nil {
			return err
		}
	}

	// Read from library FS or disk.
	var data []byte
	libFS := ctx.LibraryFS
	if libFS != nil {
		d, err := fs.ReadFile(libFS, templateSource)
		if err != nil {
			// Template doesn't exist, skip.
			return nil
		}
		data = d
	} else {
		templatePath := filepath.Join(ctx.LibraryDir, templateSource)
		if !files.FileExists(templatePath) {
			return nil
		}
		d, err := files.ReadFile(templatePath)
		if err != nil {
			return err
		}
		data = d
	}

	if err := files.WriteFile(destPath, data, 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(destPath)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path:   recordPath,
		Hash:   hash,
		Source: templateSource,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

// ---------------------------------------------------------------------------
// Orchestrator helpers
// ---------------------------------------------------------------------------

// Fallback orchestrator tool names when the agent source file is unavailable.
var fallbackOrchestratorTools = []string{
	"list_catalog",
	"compose_agent",
	"start_chain",
	"advance_chain",
	"get_status",
	"get_budget",
	"retry_step",
	"escalate_step",
	"handoff",
}

// IsOrchestratorEnabled reports whether the orchestrator MCP server is enabled.
func IsOrchestratorEnabled(ctx *AdapterContext) bool {
	for _, s := range ctx.EnableServers {
		if s == "orchestrator" {
			return true
		}
	}
	return false
}

// readOrchestratorAgentSource reads the orchestrator agent source file,
// falling back to a hardcoded default if the file is missing.
func readOrchestratorAgentSource(ctx *AdapterContext) string {
	libFS := ctx.LibraryFS
	if libFS != nil {
		data, err := fs.ReadFile(libFS, "agents/orchestrator.md")
		if err == nil {
			return string(data)
		}
	} else if ctx.LibraryDir != "" {
		sourcePath := filepath.Join(ctx.LibraryDir, "agents", "orchestrator.md")
		if files.FileExists(sourcePath) {
			data, err := files.ReadFile(sourcePath)
			if err == nil {
				return string(data)
			}
		}
	}

	// Fallback
	lines := []string{
		"---",
		"name: Orchestrator",
		"model: opus",
		fmt.Sprintf("tools: %s", strings.Join(fallbackOrchestratorTools, " ")),
		"---",
		"",
		"# Orchestrator Agent",
		"",
		"Use the orchestration MCP runtime to coordinate multi-agent execution.",
	}
	return strings.Join(lines, "\n")
}

// ReadSampleRuleContent reads the TypeScript sample rule from the library.
// Returns an error if the file is not found (not a silent skip).
func ReadSampleRuleContent(ctx *AdapterContext) ([]byte, error) {
	libFS := ctx.LibraryFS
	if libFS != nil {
		data, err := fs.ReadFile(libFS, "rules/typescript.md")
		if err == nil {
			return data, nil
		}
	} else if ctx.LibraryDir != "" {
		sourcePath := filepath.Join(ctx.LibraryDir, "rules", "typescript.md")
		if files.FileExists(sourcePath) {
			data, err := files.ReadFile(sourcePath)
			if err == nil {
				return data, nil
			}
		}
	}
	return nil, fmt.Errorf("sample rule library/rules/typescript.md not found")
}

// ExtractTools parses the tools list from YAML frontmatter in content.
func ExtractTools(content string) []string {
	fm, _, err := frontmatter.ExtractFrontmatter([]byte(content))
	if err != nil || fm == nil {
		return nil
	}
	toolsVal, ok := fm["tools"]
	if !ok {
		return nil
	}

	switch v := toolsVal.(type) {
	case string:
		parts := strings.Fields(v)
		var result []string
		for _, p := range parts {
			p = strings.TrimRight(p, ",")
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	case []any:
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// ReadOrchestratorTools returns the tools list from the orchestrator agent.
func ReadOrchestratorTools(ctx *AdapterContext) []string {
	source := readOrchestratorAgentSource(ctx)
	tools := ExtractTools(source)
	if len(tools) > 0 {
		return tools
	}
	result := make([]string, len(fallbackOrchestratorTools))
	copy(result, fallbackOrchestratorTools)
	return result
}

// StripFrontmatterAndInjectModel strips YAML frontmatter and injects the model
// and tools as HTML comments at the top of the content (for OpenCode format).
func StripFrontmatterAndInjectModel(content []byte) []byte {
	str := string(content)
	_, body := frontmatter.SplitYamlFrontmatter(str)
	fm, fmBody, _ := frontmatter.ExtractFrontmatter(content)

	var comments []string
	if fm != nil {
		if model, ok := fm["model"].(string); ok && model != "" {
			comments = append(comments, fmt.Sprintf("<!-- Recommended model: %s -->", model))
		}
		tools := ExtractTools(string(fmBody))
		if len(tools) > 0 {
			comments = append(comments, fmt.Sprintf("<!-- allowed-tools: %s -->", strings.Join(tools, ", ")))
		}
	}

	bodyStr := strings.TrimLeft(body, "\n")
	if len(comments) == 0 {
		return []byte(bodyStr)
	}
	return []byte(strings.Join(comments, "\n") + "\n\n" + bodyStr)
}

// NormalizeToolsFrontmatter rewrites the tools line in frontmatter to use the
// given delimiter format.
func NormalizeToolsFrontmatter(content string, delimiter string) string {
	fmBody, body := frontmatter.SplitYamlFrontmatter(content)
	if fmBody == "" {
		return content
	}

	tools := parseToolsFromFrontmatterBody(fmBody)
	if len(tools) == 0 {
		return content
	}

	var joined string
	switch delimiter {
	case "comma":
		joined = strings.Join(tools, ", ")
	default:
		joined = strings.Join(tools, " ")
	}

	rebuilt := rewriteToolsLine(fmBody, joined)
	bodyStr := body
	if !strings.HasPrefix(bodyStr, "\n") {
		bodyStr = "\n" + bodyStr
	}
	return "---\n" + rebuilt + "\n---\n" + bodyStr
}

// EnsureModeAgentFrontmatter ensures the content has mode: agent in its frontmatter.
func EnsureModeAgentFrontmatter(content string) string {
	fmBody, body := frontmatter.SplitYamlFrontmatter(content)
	if fmBody == "" {
		return "---\nmode: agent\n---\n\n" + content
	}

	lines := strings.Split(fmBody, "\n")
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "mode:") {
			lines[i] = "mode: agent"
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, "mode: agent")
	}

	bodyStr := body
	if !strings.HasPrefix(bodyStr, "\n") {
		bodyStr = "\n" + bodyStr
	}
	return "---\n" + strings.Join(lines, "\n") + "\n---\n" + bodyStr
}

// GetOrchestratorAgentContent returns the orchestrator agent content with
// whitespace-delimited tools in the frontmatter (per Claude Code spec).
func GetOrchestratorAgentContent(ctx *AdapterContext) []byte {
	source := readOrchestratorAgentSource(ctx)
	return []byte(NormalizeToolsFrontmatter(source, "space"))
}

// GetOrchestratorSkillContent returns the orchestrator as a skill file.
func GetOrchestratorSkillContent(ctx *AdapterContext) []byte {
	source := readOrchestratorAgentSource(ctx)
	_, body := frontmatter.SplitYamlFrontmatter(source)
	tools := ReadOrchestratorTools(ctx)

	lines := []string{
		"---",
		"name: orchestrator",
		"description: Orchestration MCP runtime guidance",
		"---",
		"",
		"# Orchestrator Skill",
		"",
		strings.TrimSpace(body),
		"",
		formatAllowedToolsSection(tools),
		"",
	}
	return []byte(strings.Join(lines, "\n"))
}

// GetOrchestratorPromptContent returns the orchestrator as a prompt file.
func GetOrchestratorPromptContent(ctx *AdapterContext) []byte {
	source := readOrchestratorAgentSource(ctx)
	_, body := frontmatter.SplitYamlFrontmatter(source)
	tools := ReadOrchestratorTools(ctx)

	lines := []string{
		"---",
		"mode: agent",
		"---",
		"",
		"# Orchestrator Prompt",
		"",
		strings.TrimSpace(body),
		"",
		formatAllowedToolsSection(tools),
		"",
	}
	return []byte(strings.Join(lines, "\n"))
}

func formatAllowedToolsSection(tools []string) string {
	if len(tools) == 0 {
		return ""
	}
	lines := []string{"## Allowed MCP Tools", ""}
	for _, tool := range tools {
		lines = append(lines, "- "+tool)
	}
	return strings.Join(lines, "\n")
}

// parseToolsFromFrontmatterBody extracts tool names from a YAML frontmatter body.
func parseToolsFromFrontmatterBody(body string) []string {
	lines := strings.Split(body, "\n")
	toolsLineIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "tools:") {
			toolsLineIdx = i
			break
		}
	}
	if toolsLineIdx == -1 {
		return nil
	}

	afterColon := strings.TrimSpace(strings.TrimPrefix(lines[toolsLineIdx], "tools:"))
	if afterColon != "" {
		return splitToolList(afterColon)
	}

	// Parse YAML list items (- item)
	var result []string
	for i := toolsLineIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			break
		}
		item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func splitToolList(value string) []string {
	var result []string
	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	}) {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func rewriteToolsLine(frontmatterBody, joined string) string {
	lines := strings.Split(frontmatterBody, "\n")
	toolsLineIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "tools:") {
			toolsLineIdx = i
			break
		}
	}
	if toolsLineIdx == -1 {
		return frontmatterBody
	}

	var result []string
	result = append(result, lines[:toolsLineIdx]...)
	result = append(result, "tools: "+joined)

	// Skip YAML list items after tools:
	i := toolsLineIdx + 1
	for i < len(lines) {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "- ") {
			i++
			continue
		}
		break
	}
	result = append(result, lines[i:]...)
	return strings.Join(result, "\n")
}

// WriteJSONFile writes a JSON-marshaled struct to a file with indentation.
func WriteJSONFile(path string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return files.WriteFile(path, content, 0o644)
}