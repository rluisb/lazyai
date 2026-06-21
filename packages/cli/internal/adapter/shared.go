// Package adapter provides shared helper functions used by all tool adapters.
// Ported from the TypeScript shared.ts utilities.
package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// Managed block markers for idempotent root AGENTS.md patching.
const (
	ManagedBlockStartMarker = "<!-- lazyai:managed:start root-agents v1 -->"
	ManagedBlockEndMarker   = "<!-- lazyai:managed:end root-agents v1 -->"
)

// vibe-lab managed-marker contract. Must match bin/inject exactly.
func managedAgentMarker(surface, name string) string {
	return fmt.Sprintf("<!-- vibe-lab:managed kind=agent surface=%s name=%s source=.agents/agents/%s.md -->", surface, name, name)
}

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

const (
	defaultAgentID          = "guide"
	defaultAgentDescription = "Front-door default agent. Answers directly, chats naturally, and only suggests or delegates specialists when that improves the outcome."
)

var canonicalAgentIDs = map[string]struct{}{
	defaultAgentID:      {},
	"implementer":       {},
	"researcher":        {},
	"deployer":          {},
	"responder":         {},
	"planner":           {},
	"reviewer":          {},
	"evidence-verifier": {},
}

func isCanonicalAgentFile(file string) bool {
	_, ok := canonicalAgentIDs[fileID(file)]
	return ok
}
func isDefaultAgentFile(file string) bool {
	return fileID(file) == defaultAgentID
}

func copyCanonicalDefaultAgent(ctx *AdapterContext, dest string, transform func([]byte) []byte) error {
	return CopyWithRecord("canonical/agents/"+defaultAgentID+".md", dest, ctx, true, transform, 0o644)
}

func openCodeDefaultAgentContent(source []byte) []byte {
	fm, body, _ := frontmatter.ExtractFrontmatter(source)
	opts := OpenCodeAgentOpts{
		Description:   inheritedDescription(fm),
		ManagedMarker: managedAgentMarker("opencode", defaultAgentID),
	}
	// Preserve exact body trailing content while adding the marker.
	return BuildOpenCodeAgentFrontmatter(append([]byte{'\n'}, body...), opts)
}

func claudeDefaultAgentContent(source []byte) []byte {
	fm, body, _ := frontmatter.ExtractFrontmatter(source)
	description := inheritedDescription(fm)
	if description == "Agent" {
		description = defaultAgentDescription
	}
	body = trimLeadingNewlines(body)
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: ")
	b.WriteString(defaultAgentID)
	b.WriteByte('\n')
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteString("\n---\n\n")
	b.WriteString(managedAgentMarker("claude", defaultAgentID))
	b.WriteString("\n\n")
	b.Write(body)
	return []byte(b.String())
}

// MergeManagedBlock merges newManagedContent into existing content using managed
// block markers. If existing is empty, returns content wrapped in markers. If
// existing has no markers, appends the managed block. If existing has markers,
// replaces only the content between them while preserving all user content outside.
func MergeManagedBlock(existing, newManagedContent []byte, startMarker, endMarker string) []byte {
	existingStr := string(existing)
	startIdx := strings.Index(existingStr, startMarker)
	endIdx := strings.Index(existingStr, endMarker)

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		// No existing managed block — append it.
		var sb strings.Builder
		if len(existing) > 0 {
			sb.Write(existing)
			if !strings.HasSuffix(existingStr, "\n") {
				sb.WriteByte('\n')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString(startMarker)
		sb.WriteByte('\n')
		sb.Write(newManagedContent)
		if !strings.HasSuffix(string(newManagedContent), "\n") {
			sb.WriteByte('\n')
		}
		sb.WriteString(endMarker)
		sb.WriteByte('\n')
		return []byte(sb.String())
	}

	// Replace content between markers idempotently.
	var sb strings.Builder
	sb.WriteString(existingStr[:startIdx+len(startMarker)])
	sb.WriteByte('\n')
	sb.Write(newManagedContent)
	if !strings.HasSuffix(string(newManagedContent), "\n") {
		sb.WriteByte('\n')
	}
	sb.WriteString(existingStr[endIdx:])
	return []byte(sb.String())
}

// WriteManagedBlockToFile reads destPath (if it exists), merges the managed
// block, and writes back. Returns the final content and whether the file was
// modified (true if created or changed).
func WriteManagedBlockToFile(destPath string, managedContent []byte, startMarker, endMarker string) ([]byte, bool, error) {
	var existing []byte
	if files.FileExists(destPath) {
		data, err := files.ReadFile(destPath)
		if err != nil {
			return nil, false, fmt.Errorf("read %s: %w", destPath, err)
		}
		existing = data
	}

	merged := MergeManagedBlock(existing, managedContent, startMarker, endMarker)
	modified := !bytesEqual(existing, merged)

	if modified {
		if err := files.WriteFile(destPath, merged, 0o644); err != nil {
			return nil, false, fmt.Errorf("write %s: %w", destPath, err)
		}
	}

	return merged, modified, nil
}

func bytesEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// CopyWithRecord copies a file from the library FS to dest, records it in
// ctx.FileRecords, and handles conflict resolution. If transform is non-nil,
// it is applied to the file content before writing.
//
// src is a relative path within ctx.LibraryFS (e.g., "agents/implementer.md").
// perm defaults to 0o644 if zero.
func CopyWithRecord(src, dest string, ctx *AdapterContext, warnOnSkip bool, transform func([]byte) []byte, perm os.FileMode) error {
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
			adapterLog.Info("skipping existing file", "path", relPath)
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
		adapterLog.Info("dry run would create file", "path", relPath)
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

	if perm == 0 {
		perm = 0o644
	}
	if err := files.WriteFile(dest, content, perm); err != nil {
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
			adapterLog.Info("skipping existing file", "path", relPath)
		}
		return nil
	}

	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return fmt.Errorf("backup %s: %w", dest, err)
		}
	}

	if ctx.DryRun {
		adapterLog.Info("dry run would create file", "path", relPath)
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
	SelectionKey string // "agents", "skills", "prompts", "chatmodes", "opencodeCommands", or "opencodeModes"
	ToDestPath   func(file string) string
	WarnOnSkip   bool
	Transform    func(content []byte) []byte
	IncludeFile  func(file string) bool
	Recursive    bool
	Mode         fs.FileMode
}

// CopyLibraryDirectory copies all files from the library subdirectory,
// applying selection filtering and conflict resolution.
// Uses ctx.LibraryFS when available, falls back to ctx.LibraryDir + filesystem.
func CopyLibraryDirectory(opts CopyLibraryDirectoryOption) error {
	mode := opts.Mode
	if mode == 0 {
		mode = 0o644
	}
	opts.Mode = mode

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
	selectedChatModes := selectionSet(opts.Ctx.Selections.ChatModes)
	selectedOpenCodeCommands := selectionSet(opts.Ctx.Selections.OpenCodeCommands)
	selectedOpenCodeModes := selectionSet(opts.Ctx.Selections.OpenCodeModes)

	for _, entry := range entries {
		if entry.IsDir() {
			if opts.Recursive {
				subDirName := entry.Name()
				dummyDest := opts.ToDestPath(subDirName)
				subDirDest := filepath.Join(filepath.Dir(dummyDest), subDirName)
				if err := files.EnsureDir(subDirDest); err != nil {
					return err
				}
				subOpts := opts
				subOpts.SourceSubdir = opts.SourceSubdir + "/" + subDirName
				subOpts.ToDestPath = func(file string) string {
					return filepath.Join(subDirDest, file)
				}
				if err := copyLibraryDirectoryFromFS(subOpts, libFS); err != nil {
					return err
				}
			}
			continue
		}
		file := entry.Name()

		if opts.IncludeFile != nil && !opts.IncludeFile(file) {
			continue
		}

		// Extract file ID (filename without extension) for selection filtering.
		// Chatmodes use a compound extension ".chatmode.md" — strip it explicitly.
		fileIDVal := strings.TrimSuffix(file, filepath.Ext(file))
		if opts.SelectionKey == "chatmodes" {
			fileIDVal = strings.TrimSuffix(fileIDVal, ".chatmode")
		}
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
		case "chatmodes":
			if selectedChatModes != nil && !selectedChatModes[types.ChatModeId(fileIDVal)] {
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
		if err := CopyWithRecord(srcPath, dest, opts.Ctx, opts.WarnOnSkip, opts.Transform, opts.Mode); err != nil {
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
	selectedChatModes := selectionSet(opts.Ctx.Selections.ChatModes)
	selectedOpenCodeCommands := selectionSet(opts.Ctx.Selections.OpenCodeCommands)
	selectedOpenCodeModes := selectionSet(opts.Ctx.Selections.OpenCodeModes)

	for _, file := range files.ListDir(sourceDir) {
		if opts.IncludeFile != nil && !opts.IncludeFile(file) {
			continue
		}

		srcPath := filepath.Join(sourceDir, file)
		if files.IsDirectory(srcPath) {
			if opts.Recursive {
				dummyDest := opts.ToDestPath(file)
				subDirDest := filepath.Join(filepath.Dir(dummyDest), file)
				if err := files.EnsureDir(subDirDest); err != nil {
					return err
				}
				subOpts := opts
				subOpts.SourceSubdir = opts.SourceSubdir + "/" + file
				subOpts.ToDestPath = func(f string) string {
					return filepath.Join(subDirDest, f)
				}
				if err := copyLibraryDirectoryFromDisk(subOpts, srcPath); err != nil {
					return err
				}
			}
			continue
		}

		// Extract file ID (filename without extension) for selection filtering.
		// Chatmodes use a compound extension ".chatmode.md" — strip it explicitly.
		fileIDVal := strings.TrimSuffix(file, filepath.Ext(file))
		if opts.SelectionKey == "chatmodes" {
			fileIDVal = strings.TrimSuffix(fileIDVal, ".chatmode")
		}
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
		case "chatmodes":
			if selectedChatModes != nil && !selectedChatModes[types.ChatModeId(fileIDVal)] {
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
		if err := CopyWithRecord(libRelPath, dest, opts.Ctx, opts.WarnOnSkip, opts.Transform, opts.Mode); err != nil {
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
		opts.Ctx, opts.WarnOnSkip, nil, 0o644,
	); err != nil {
		return err
	}

	// skills directory context file
	if opts.SkillsDestDir != "" {
		if err := CopyWithRecord(
			"tool-agents/skills-dir.md",
			filepath.Join(opts.ToolDir, opts.SkillsDestDir, opts.ContextFileName),
			opts.Ctx, opts.WarnOnSkip, nil, 0o644,
		); err != nil {
			return err
		}
	}

	// templates directory context file
	if opts.TemplatesDestDir != "" {
		if err := CopyWithRecord(
			"tool-agents/templates-dir.md",
			filepath.Join(opts.ToolDir, opts.TemplatesDestDir, opts.ContextFileName),
			opts.Ctx, opts.WarnOnSkip, nil, 0o644,
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
		opts.Ctx, opts.WarnOnSkip, nil, 0o644,
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

// truncateOutput truncates s to maxLen characters, appending "..." when truncated.
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ChmodScriptsExecutable walks a directory and chmods all .sh files to 0o755.
func ChmodScriptsExecutable(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sh") {
			if err := os.Chmod(path, 0o755); err != nil {
				return err
			}
		}
		return nil
	})
}
