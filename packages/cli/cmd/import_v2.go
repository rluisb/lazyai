package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/migration"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

const migrationReportRelPath = ".ai/migration-report.md"

func canonicalImportDetections(detections []migration.DetectionResult) []migration.DetectionResult {
	parsed := make([]migration.DetectionResult, 0, len(detections))
	for _, detection := range detections {
		if detection.AdapterID == "opencode" {
			parsed = append(parsed, detection)
		}
	}
	return parsed
}

func planRawPreservation(sourcePath, targetPath string, detections []migration.DetectionResult, write bool) ([]string, error) {
	seen := map[string]bool{}
	preserved := make([]string, 0)
	for _, detection := range detections {
		for _, file := range detection.Files {
			key := detection.AdapterID + "\x00" + file.Path
			if seen[key] {
				continue
			}
			seen[key] = true

			source := filepath.Join(sourcePath, file.Path)
			destination := filepath.Join(targetPath, ".ai", "adapters", detection.AdapterID, "raw", file.Path)
			rel, err := filepath.Rel(targetPath, destination)
			if err != nil {
				rel = destination
			}
			preserved = append(preserved, rel)
			if !write {
				continue
			}
			if err := files.CopyFile(source, destination); err != nil {
				return nil, fmt.Errorf("copy %s -> %s: %w", source, destination, err)
			}
		}
	}
	sort.Strings(preserved)
	return preserved, nil
}

func scaffoldImportManifest(targetPath string, detections []migration.DetectionResult, preview bool) error {
	if preview {
		return nil
	}
	tools := toolIDsForDetections(detections)
	if len(tools) == 0 {
		return nil
	}
	return scaffold.ScaffoldManifest(targetPath, tools, nil)
}

func toolIDsForDetections(detections []migration.DetectionResult) []types.ToolId {
	seen := map[types.ToolId]bool{}
	tools := make([]types.ToolId, 0, len(detections))
	for _, detection := range detections {
		tool, ok := toolIDForDetection(detection.AdapterID)
		if !ok || seen[tool] {
			continue
		}
		seen[tool] = true
		tools = append(tools, tool)
	}
	return tools
}

func toolIDForDetection(adapter string) (types.ToolId, bool) {
	switch adapter {
	case "opencode":
		return types.ToolIdOpenCode, true
	case "claude-code":
		return types.ToolIdClaudeCode, true
	case "copilot":
		return types.ToolIdCopilot, true
	case "pi":
		return types.ToolIdPi, true
	case "omp":
		return types.ToolIdOmp, true
	case "kiro":
		return types.ToolIdKiro, true
	case "antigravity":
		return types.ToolIdAntigravity, true
	default:
		return "", false
	}
}

func writeMigrationReport(sourcePath, targetPath string, detections []migration.DetectionResult, result *migration.MigrationResult, rawPaths []string, preview bool) (string, error) {
	if preview {
		return migrationReportRelPath, nil
	}
	lines := []string{
		"# Migration Report",
		"",
		fmt.Sprintf("Generated: %s", time.Now().UTC().Format(time.RFC3339)),
		fmt.Sprintf("Source: `%s`", sourcePath),
		fmt.Sprintf("Target: `%s`", targetPath),
		"",
		"## Summary",
		fmt.Sprintf("- Native detections: %d", len(detections)),
		fmt.Sprintf("- Canonical files written: %d", result.Stats.FilesCreated),
		fmt.Sprintf("- Raw native files preserved: %d", len(rawPaths)),
		"- Native file deletions: 0",
	}
	if result.BackupPath != "" {
		lines = append(lines, fmt.Sprintf("- Backup directory: `%s`", result.BackupPath))
	}
	lines = append(lines,
		"",
		"## Confidence levels",
		"- exact — direct canonical mapping",
		"- high — mapping likely complete, minor target metadata lost",
		"- medium — instructions imported but semantics uncertain",
		"- low — copied as raw target-specific asset",
		"- unsupported — left in native file only",
		"",
		"## Trust model",
		"- Imported canonical files are copied from external AI-tool config and should be reviewed before compile/use.",
		"- Raw native files preserved under `.ai/adapters/<target>/raw/` are snapshots for audit/recovery, not a safety boundary.",
		"- Run `lazyai-cli validate --all` after import and review any MCP command, hook, or secret warnings before enabling outputs.",
		"",
		"## Detections",
	)

	for _, detection := range detections {
		lines = append(lines,
			fmt.Sprintf("### %s (`%s`)", detection.AdapterName, detection.AdapterID),
			fmt.Sprintf("- Detector confidence: %.0f%%", detection.Confidence*100),
			fmt.Sprintf("- Conversion confidence: %s", migrationConfidenceLabel(detection)),
			fmt.Sprintf("- Native files observed: %d", len(detection.Files)),
		)
		if detection.AdapterID == "opencode" {
			lines = append(lines, "- Canonical extraction: yes")
		} else {
			lines = append(lines, "- Canonical extraction: no (raw preserved under `.ai/adapters/<target>/raw/`)")
		}
		lines = append(lines, "- Observed paths:")
		for _, file := range detection.Files {
			lines = append(lines, fmt.Sprintf("  - `%s`", file.Path))
		}
		lines = append(lines, "")
	}

	if len(rawPaths) > 0 {
		lines = append(lines, "## Preserved raw files")
		for _, rel := range rawPaths {
			lines = append(lines, fmt.Sprintf("- `%s`", rel))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "## Notes", "- Unsupported native fields are preserved by copying detected files under `.ai/adapters/<target>/raw/`.", "- Generated regions/native outputs are untouched; import never deletes originals.", "")

	content := strings.Join(lines, "\n")
	path := filepath.Join(targetPath, migrationReportRelPath)
	if err := files.EnsureDir(filepath.Dir(path)); err != nil {
		return "", err
	}
	if err := files.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return migrationReportRelPath, nil
}

func migrationConfidenceLabel(detection migration.DetectionResult) string {
	switch detection.AdapterID {
	case "opencode":
		return "high"
	case "claude-code":
		if len(detection.Files) == 1 && strings.EqualFold(filepath.Base(detection.Files[0].Path), "CLAUDE.md") {
			return "medium"
		}
		return "low"
	case "copilot", "pi", "omp", "kiro", "antigravity":
		return "low"
	default:
		return "unsupported"
	}
}
