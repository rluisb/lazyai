package adapter

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

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

// trackedRecordPath computes a normalized path for a TrackedFile record.
//
// It attempts to make the path relative to workspaceRoot (the canonical
// workspace or project root). When the file is under workspaceRoot, the
// result is a slash-normalized relative path — byte-stable across platforms
// and consistent with install-side records (CopyWithRecord,
// WriteContentWithRecord).
//
// When the file is outside workspaceRoot (e.g. a global-scope config under
// the user's home directory), filepath.Rel fails and the absolute path is
// used instead. This is an intentional, documented exception: global/user-owned
// paths are not under the project's control and must remain absolute to be
// meaningful. Even in this fallback, filepath.ToSlash is applied so the
// record is still slash-normalized.
func trackedRecordPath(workspaceRoot, filePath string) string {
	rel, err := filepath.Rel(workspaceRoot, filePath)
	if err != nil || rel == "" || strings.HasPrefix(rel, "..") {
		// Outside workspaceRoot or degenerate — keep absolute, slash-normalized.
		return filepath.ToSlash(filePath)
	}
	return filepath.ToSlash(rel)
}
