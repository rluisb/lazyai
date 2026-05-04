// Package conflict provides conflict resolution strategies for file operations.
// Ported from the TypeScript utilities in src/utils/conflicts.ts and
// src/utils/conflict-strategy.ts.
package conflict

import (
	"fmt"

	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// Action defines the outcome of resolving a conflict.
type Action string

const (
	ActionKeep    Action = "keep"
	ActionReplace Action = "replace"
	ActionMerge   Action = "merge"
	ActionSkip    Action = "skip"
)

// Conflict describes a file conflict between current and new content.
type Conflict struct {
	Path           string
	CurrentContent []byte
	NewContent     []byte
	CurrentHash    string
	NewHash        string
	TargetDir      string // directory containing the file, used for backups
}

// Resolution is the result of applying a conflict strategy.
type Resolution struct {
	Path       string
	Action     Action
	Content    []byte
	BackupPath string
}

// ResolveConflict applies a ConflictStrategy to a Conflict and returns a Resolution.
// This is the deterministic, non-interactive version matching the TypeScript applyStrategy.
func ResolveConflict(conflict Conflict, strategy types.ConflictStrategy) (*Resolution, error) {
	exists := files.FileExists(conflict.Path)

	// New file — always write regardless of strategy.
	if !exists {
		return &Resolution{
			Path:    conflict.Path,
			Action:  ActionReplace,
			Content: conflict.NewContent,
		}, nil
	}

	switch strategy {
	case types.ConflictStrategySkip:
		return &Resolution{
			Path:   conflict.Path,
			Action: ActionSkip,
		}, nil

	case types.ConflictStrategyBackupAndReplace:
		backupPath, err := files.BackupFile(conflict.Path, conflict.TargetDir)
		if err != nil {
			return nil, aierror.Unknown(fmt.Sprintf("backup failed for %s", conflict.Path), err)
		}
		return &Resolution{
			Path:       conflict.Path,
			Action:     ActionReplace,
			Content:    conflict.NewContent,
			BackupPath: backupPath,
		}, nil

	case types.ConflictStrategyAlign:
		// For align strategy, backup and write (same as backup-and-replace;
		// the diff preview already happened at the wizard stage).
		backupPath, err := files.BackupFile(conflict.Path, conflict.TargetDir)
		if err != nil {
			return nil, aierror.Unknown(fmt.Sprintf("backup failed for %s", conflict.Path), err)
		}
		return &Resolution{
			Path:       conflict.Path,
			Action:     ActionReplace,
			Content:    conflict.NewContent,
			BackupPath: backupPath,
		}, nil

	default:
		return nil, aierror.ConflictUnresolved(conflict.Path, string(strategy))
	}
}

// ApplyStrategy is a simplified version matching the TypeScript signature.
// It returns "write" or "skip" based on the strategy and file existence.
func ApplyStrategy(destPath string, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy, targetDir string) (string, error) {
	effective := strategy
	if override, ok := perFileOverrides[destPath]; ok {
		effective = override
	}

	exists := files.FileExists(destPath)

	// New file — always write regardless of strategy.
	if !exists {
		return "write", nil
	}

	switch effective {
	case types.ConflictStrategySkip:
		return "skip", nil
	case types.ConflictStrategyBackupAndReplace:
		_, err := files.BackupFile(destPath, targetDir)
		if err != nil {
			return "", err
		}
		return "write", nil
	case types.ConflictStrategyAlign:
		_, err := files.BackupFile(destPath, targetDir)
		if err != nil {
			return "", err
		}
		return "write", nil
	default:
		return "write", nil
	}
}

// ResolveConflictWithOptions is the richer conflict resolver matching the
// TypeScript resolveConflict, supporting force flag and tracked hash comparison.
type ConflictOptions struct {
	Force       bool
	TrackedHash string
	Strategy    types.ConflictStrategy
}

// ResolveConflictWithOptions resolves a conflict with the given options.
// Returns the ConflictResolution action to take.
func ResolveConflictWithOptions(destPath string, displayName string, options ConflictOptions) (Action, error) {
	if !files.FileExists(destPath) {
		return ActionReplace, nil
	}

	// Deterministic strategy bypasses interactive logic.
	if options.Strategy != "" {
		switch options.Strategy {
		case types.ConflictStrategySkip:
			return ActionSkip, nil
		case types.ConflictStrategyBackupAndReplace, types.ConflictStrategyAlign:
			return ActionReplace, nil
		}
	}

	if options.Force {
		return ActionReplace, nil
	}

	if options.TrackedHash != "" {
		currentHash, err := files.FileHash(destPath)
		if err != nil {
			return ActionSkip, err
		}
		if currentHash == options.TrackedHash {
			return ActionReplace, nil
		}
		// File has been customized — in Go (non-interactive), default to skip.
		return ActionSkip, nil
	}

	// File exists with no hash info — in Go (non-interactive), default to skip.
	return ActionSkip, nil
}
