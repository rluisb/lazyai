package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/manifest"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// Execute runs the migration plan, creating files and backups as needed.
// Ported from executeMigrationPlan in src/migration/executor.ts.
func Execute(plan *MigrationPlan, ctx *MigrationContext) (*MigrationResult, error) {
	var executedActions []MigrationAction
	var errors []string
	var warnings []string
	stats := MigrationStats{}

	var backupPath string
	var err error

	// Step 1: Create backup directory.
	if !ctx.Options.SkipBackup && !ctx.Options.Preview {
		backupPath, err = createBackupDir(ctx.TargetPath)
		if err != nil {
			return nil, fmt.Errorf("create backup dir: %w", err)
		}
	}

	// Step 2: Execute actions in order.
	for _, action := range plan.Actions {
		switch action.Type {
		case ActionTypeBackup:
			if err := executeBackupAction(action, ctx, backupPath); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to backup %s: %s", action.SourcePath, err))
				continue
			}
			executedActions = append(executedActions, action)
			stats.FilesBackedUp++

		case ActionTypeCreate:
			if err := executeCreateAction(action, ctx); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to create %s: %s", action.TargetPath, err))
				continue
			}
			executedActions = append(executedActions, action)
			stats.FilesCreated++

		case ActionTypeModify:
			if err := executeModifyAction(action, ctx); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to modify %s: %s", action.TargetPath, err))
				continue
			}
			executedActions = append(executedActions, action)
			stats.FilesModified++

		case ActionTypeSkip:
			executedActions = append(executedActions, action)
			stats.FilesSkipped++
		}
	}

	// Step 3: Count conflicts.
	for _, c := range plan.Conflicts {
		if c.Resolved {
			stats.ConflictsResolved++
		} else {
			stats.ConflictsUnresolved++
		}
	}

	// Step 4: Update the manifest.
	if err := updateManifest(ctx, plan, nil); err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to update manifest: %s", err))
	}

	return &MigrationResult{
		Success:         len(errors) == 0,
		Plan:            plan,
		ExecutedActions: executedActions,
		BackupPath:      backupPath,
		Errors:          errors,
		Warnings:        warnings,
		Stats:           stats,
	}, nil
}

// ExecuteToCanonical writes parsed data into the .ai/ canonical format.
// Ported from executeMigrationToCanonical in src/migration/executor.ts.
func ExecuteToCanonical(ctx *MigrationContext, plan *MigrationPlan, parsedSetups []ParsedSetup) (*MigrationResult, error) {
	var executedActions []MigrationAction
	var errors []string
	var warnings []string
	stats := MigrationStats{}
	var fileRecords []types.TrackedFile

	var backupPath string
	var err error

	if !ctx.Options.SkipBackup && !ctx.Options.Preview {
		backupPath, err = createBackupDir(ctx.TargetPath)
		if err != nil {
			return nil, fmt.Errorf("create backup dir: %w", err)
		}
	}

	// Backup source files.
	if backupPath != "" {
		for _, parsed := range parsedSetups {
			for _, file := range parsed.Files {
				if err := executeBackupAction(MigrationAction{
					Type:       ActionTypeBackup,
					SourcePath: file.Path,
					TargetPath: filepath.Join(".ai-setup-backup", file.Path),
				}, ctx, backupPath); err != nil {
					// Skip silently for virtual files.
					continue
				}
				stats.FilesBackedUp++
			}
		}
	}

	// Write canonical files for each parsed setup.
	for _, parsed := range parsedSetups {
		result, err := WriteCanonical(ctx.TargetPath, &parsed, &fileRecords, ctx.Options.Preview)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Canonical write failed: %s", err))
			continue
		}

		for _, p := range result.Agents {
			executedActions = append(executedActions, MigrationAction{
				Type:        ActionTypeCreate,
				TargetPath:  p,
				Description: fmt.Sprintf("Create %s", p),
				Reason:      fmt.Sprintf("Canonical migration from %s", parsed.Metadata["adapter"]),
			})
		}
		for _, p := range result.Skills {
			executedActions = append(executedActions, MigrationAction{
				Type:        ActionTypeCreate,
				TargetPath:  p,
				Description: fmt.Sprintf("Create %s", p),
				Reason:      fmt.Sprintf("Canonical migration from %s", parsed.Metadata["adapter"]),
			})
		}
		for _, p := range result.Prompts {
			executedActions = append(executedActions, MigrationAction{
				Type:        ActionTypeCreate,
				TargetPath:  p,
				Description: fmt.Sprintf("Create %s", p),
				Reason:      fmt.Sprintf("Canonical migration from %s", parsed.Metadata["adapter"]),
			})
		}
		for _, p := range result.Rules {
			executedActions = append(executedActions, MigrationAction{
				Type:        ActionTypeCreate,
				TargetPath:  p,
				Description: fmt.Sprintf("Create %s", p),
				Reason:      fmt.Sprintf("Canonical migration from %s", parsed.Metadata["adapter"]),
			})
		}
		if result.RootConfig != "" {
			executedActions = append(executedActions, MigrationAction{
				Type:        ActionTypeCreate,
				TargetPath:  result.RootConfig,
				Description: fmt.Sprintf("Create %s", result.RootConfig),
				Reason:      fmt.Sprintf("Canonical migration from %s", parsed.Metadata["adapter"]),
			})
		}

		stats.FilesCreated += len(result.Agents) + len(result.Skills) + len(result.Prompts) + len(result.Rules)
		if result.RootConfig != "" {
			stats.FilesCreated++
		}
		stats.FilesSkipped += len(result.Skipped)
	}

	if err := updateManifest(ctx, plan, fileRecords); err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to update manifest: %s", err))
	}

	return &MigrationResult{
		Success:         len(errors) == 0,
		Plan:            plan,
		ExecutedActions: executedActions,
		BackupPath:      backupPath,
		Errors:          errors,
		Warnings:        warnings,
		Stats:           stats,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func createBackupDir(targetDir string) (string, error) {
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05")
	backupPath := filepath.Join(targetDir, ".ai-setup-backup", fmt.Sprintf("migration-%s", timestamp))
	if err := files.EnsureDir(backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

func executeBackupAction(action MigrationAction, ctx *MigrationContext, backupPath string) error {
	if action.SourcePath == "" {
		return nil
	}

	src := filepath.Join(ctx.SourcePath, action.SourcePath)
	if !files.FileExists(src) {
		return nil
	}

	dst := filepath.Join(backupPath, action.TargetPath)
	if err := files.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	return files.CopyFile(src, dst)
}

func executeCreateAction(action MigrationAction, ctx *MigrationContext) error {
	targetFullPath := filepath.Join(ctx.TargetPath, action.TargetPath)
	return files.EnsureDir(filepath.Dir(targetFullPath))
}

func executeModifyAction(action MigrationAction, ctx *MigrationContext) error {
	targetFullPath := filepath.Join(ctx.TargetPath, action.TargetPath)
	return files.EnsureDir(filepath.Dir(targetFullPath))
}

// updateManifest creates or updates .ai-setup.json with migrated file records.
func updateManifest(ctx *MigrationContext, plan *MigrationPlan, fileRecords []types.TrackedFile) error {
	if ctx.Options.Preview {
		return nil
	}

	store, err := manifest.ReadManifestOptional(ctx.TargetPath)
	if err != nil {
		return err
	}
	if store == nil || store.Meta.SchemaVersion == 0 {
		defaults := types.DefaultStoreData()
		defaults.Config.SetupScope = types.SetupScopeProject
		defaults.Config.TargetDir = ctx.TargetPath
		defaults.Config.ProjectName = filepath.Base(ctx.TargetPath)
		store = &defaults
	}

	now := time.Now().UTC().Format(time.RFC3339)
	store.Meta.LastUpdatedAt = now
	if store.Config.TargetDir == "" {
		store.Config.TargetDir = ctx.TargetPath
	}
	if store.Config.ProjectName == "" {
		store.Config.ProjectName = filepath.Base(ctx.TargetPath)
	}

	existingTools := map[types.ToolId]bool{}
	for _, tool := range store.Config.Tools {
		existingTools[tool] = true
	}
	for _, adapter := range plan.Adapters {
		toolID, ok := adapterToToolID(adapter)
		if ok && !existingTools[toolID] {
			store.Config.Tools = append(store.Config.Tools, toolID)
			existingTools[toolID] = true
		}
	}

	existingFiles := map[string]int{}
	for i, tracked := range store.Files {
		existingFiles[tracked.Path] = i
	}
	for _, record := range fileRecords {
		record.Status = types.FileStatusInstalled
		record.LastCheckedAt = now
		if record.InstalledAt == "" {
			record.InstalledAt = now
		}
		if idx, ok := existingFiles[record.Path]; ok {
			store.Files[idx] = record
			continue
		}
		store.Files = append(store.Files, record)
		existingFiles[record.Path] = len(store.Files) - 1
	}

	store.Sync.LastSyncAt = now
	store.Sync.Dirty = false

	return manifest.WriteManifest(ctx.TargetPath, store)
}

func adapterToToolID(adapter string) (types.ToolId, bool) {
	switch adapter {
	case "opencode":
		return types.ToolIdOpenCode, true
	case "claude-code":
		return types.ToolIdClaudeCode, true
	case "copilot":
		return types.ToolIdCopilot, true
	default:
		return "", false
	}
}

// Rollback is a placeholder for restoring backed-up files.
func Rollback(_ *MigrationContext, _ string) error {
	fmt.Fprintln(os.Stderr, "Rolling back migration...")
	return nil
}
