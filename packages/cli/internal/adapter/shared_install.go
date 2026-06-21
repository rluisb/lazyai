package adapter

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

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
