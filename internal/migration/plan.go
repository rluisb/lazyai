package migration

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
)

// BuildPlan creates a migration plan from detected setup results.
// Ported from generateMigrationPlan in src/migration/plan.ts.
func BuildPlan(ctx *MigrationContext, detections []DetectionResult) (*MigrationPlan, error) {
	var actions []MigrationAction
	var conflicts []MergeConflict
	var adapters []string

	existingFiles := getExistingAiSetupFiles(ctx.TargetPath)

	for _, detection := range detections {
		adapters = append(adapters, detection.AdapterID)

		for _, file := range detection.Files {
			targetRel := filepath.Join(".ai-setup-backup", file.Path)

			// Check if the target file already exists in the ai-setup output.
			if existingFiles[file.Path] {
				actions = append(actions, MigrationAction{
					Type:        ActionTypeBackup,
					SourcePath:  file.Path,
					TargetPath:  targetRel,
					Description: fmt.Sprintf("Backup existing %s", file.Path),
					Reason:      "Will be modified by migration",
				})
				actions = append(actions, MigrationAction{
					Type:        ActionTypeModify,
					TargetPath:  file.Path,
					Description: fmt.Sprintf("Update %s", file.Path),
					Reason:      "Merged with existing setup",
				})
			} else {
				actions = append(actions, MigrationAction{
					Type:        ActionTypeCreate,
					TargetPath:  canonicalTargetPath(file),
					Description: fmt.Sprintf("Create %s", canonicalTargetPath(file)),
					Reason:      fmt.Sprintf("Migrated from %s (%s)", detection.AdapterID, file.Type),
				})
			}
		}
	}

	unresolved := 0
	for _, c := range conflicts {
		if !c.Resolved {
			unresolved++
		}
	}

	canProceed := unresolved == 0 || ctx.Options.MergeStrategy != MergeStrategySmart

	return &MigrationPlan{
		SourcePath:         ctx.SourcePath,
		TargetPath:         ctx.TargetPath,
		Adapters:           adapters,
		Actions:            actions,
		Conflicts:          conflicts,
		EstimatedFiles:     countNonSkipActions(actions),
		EstimatedConflicts: unresolved,
		CanProceed:         canProceed,
	}, nil
}

// FormatPlan renders a MigrationPlan as a human-readable string.
// Ported from formatPlan in src/migration/plan.ts.
func FormatPlan(plan *MigrationPlan) string {
	var lines []string

	adapterLabels := map[string]string{
		"opencode":    "OpenCode",
		"claude-code": "Claude Code",
		"gemini":      "Gemini CLI",
		"copilot":     "GitHub Copilot",
	}

	lines = append(lines, "Migration plan")
	lines = append(lines, "==============")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Source: %s", plan.SourcePath))
	lines = append(lines, fmt.Sprintf("Target: %s", plan.TargetPath))

	adapterNames := make([]string, 0, len(plan.Adapters))
	for _, a := range plan.Adapters {
		if label, ok := adapterLabels[a]; ok {
			adapterNames = append(adapterNames, label)
		} else {
			adapterNames = append(adapterNames, a)
		}
	}
	lines = append(lines, fmt.Sprintf("Detected adapters: %s", strings.Join(adapterNames, ", ")))
	lines = append(lines, "")

	// Group actions by type.
	var creates, modifies, backups, skips []MigrationAction
	for _, a := range plan.Actions {
		switch a.Type {
		case ActionTypeCreate:
			creates = append(creates, a)
		case ActionTypeModify:
			modifies = append(modifies, a)
		case ActionTypeBackup:
			backups = append(backups, a)
		case ActionTypeSkip:
			skips = append(skips, a)
		}
	}

	unresolved := 0
	for _, c := range plan.Conflicts {
		if !c.Resolved {
			unresolved++
		}
	}

	lines = append(lines, "Summary:")
	lines = append(lines, fmt.Sprintf("  Create: %d", len(creates)))
	lines = append(lines, fmt.Sprintf("  Modify: %d", len(modifies)))
	lines = append(lines, fmt.Sprintf("  Backup: %d", len(backups)))
	lines = append(lines, fmt.Sprintf("  Skip: %d", len(skips)))
	lines = append(lines, fmt.Sprintf("  Unresolved conflicts: %d", unresolved))
	lines = append(lines, "")

	if len(creates) > 0 {
		lines = append(lines, fmt.Sprintf("Create %d new file(s):", len(creates)))
		for _, a := range creates {
			lines = append(lines, fmt.Sprintf("   + %s", a.TargetPath))
		}
		lines = append(lines, "")
	}

	if len(modifies) > 0 {
		lines = append(lines, fmt.Sprintf("Modify %d file(s):", len(modifies)))
		for _, a := range modifies {
			lines = append(lines, fmt.Sprintf("   ~ %s", a.TargetPath))
		}
		lines = append(lines, "")
	}

	if len(backups) > 0 {
		lines = append(lines, fmt.Sprintf("Backup %d file(s):", len(backups)))
		for _, a := range backups {
			lines = append(lines, fmt.Sprintf("   <- %s -> %s", a.SourcePath, a.TargetPath))
		}
		lines = append(lines, "")
	}

	if unresolved > 0 {
		lines = append(lines, fmt.Sprintf("%d unresolved conflict(s):", unresolved))
		for _, c := range plan.Conflicts {
			if !c.Resolved {
				lines = append(lines, fmt.Sprintf("   ! %s (lines %d-%d)", c.File, c.LineStart, c.LineEnd))
			}
		}
		lines = append(lines, "")
	} else {
		lines = append(lines, "No unresolved conflicts detected.")
		lines = append(lines, "")
	}

	if plan.CanProceed {
		lines = append(lines, "Status: ready to apply")
	} else {
		lines = append(lines, "Status: blocked until conflicts are resolved")
	}

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// canonicalTargetPath maps a detected file to its target path under .ai/.
func canonicalTargetPath(file DetectedFile) string {
	switch file.Type {
	case FileTypeAgent:
		return filepath.Join(".ai", "agents", filepath.Base(file.Path))
	case FileTypeCommand:
		return filepath.Join(".ai", "skills", filepath.Base(file.Path))
	case FileTypeRule:
		return filepath.Join(".ai", "rules", filepath.Base(file.Path))
	case FileTypeTemplate:
		return filepath.Join(".ai", "prompts", filepath.Base(file.Path))
	default:
		return filepath.Join(".ai", filepath.Base(file.Path))
	}
}

// getExistingAiSetupFiles returns a set of paths already tracked by ai-setup.
func getExistingAiSetupFiles(targetDir string) map[string]bool {
	result := make(map[string]bool)

	// Check for .ai-setup.json manifest.
	manifestPath := filepath.Join(targetDir, ".ai-setup.json")
	if data, err := files.ReadFile(manifestPath); err == nil {
		var manifest struct {
			Files []struct {
				Path string `json:"path"`
			} `json:"files"`
		}
		if json.Unmarshal(data, &manifest) == nil {
			for _, f := range manifest.Files {
				result[f.Path] = true
			}
		}
	}

	// Also check for common files.
	commonFiles := []string{
		"AGENTS.md",
		"CLAUDE.md",
		"GEMINI.md",
		".ai-setup.json",
	}
	for _, f := range commonFiles {
		if files.FileExists(filepath.Join(targetDir, f)) {
			result[f] = true
		}
	}

	return result
}

func countNonSkipActions(actions []MigrationAction) int {
	count := 0
	for _, a := range actions {
		if a.Type != ActionTypeSkip {
			count++
		}
	}
	return count
}
