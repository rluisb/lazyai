package migration

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

// ---------------------------------------------------------------------------
// Detection patterns for different AI tools
// ---------------------------------------------------------------------------

// DetectionPatterns maps adapter IDs to glob patterns that indicate their presence.
// Ported from DETECTION_PATTERNS in src/migration/detector.ts.
var DetectionPatterns = map[string][]string{
	"opencode": {
		".opencode",
		"opencode.json",
		"AGENTS.md",
	},
	"claude-code": {
		".claude",
		"CLAUDE.md",
	},
	"copilot": {
		".github/copilot-instructions.md",
		".github/instructions",
		".github/skills",
	},
	"pi": {
		".pi",
	},
	"omp": {
		".omp",
	},
	"kiro": {
		".kiro",
	},
	"antigravity": {
		".agents",
	},
}

// AdapterNames maps adapter IDs to human-readable names.
var AdapterNames = map[string]string{
	"opencode":    "OpenCode",
	"claude-code": "Claude Code",
	"copilot":     "GitHub Copilot",
	"pi":          "Pi",
	"omp":         "OMP",
	"kiro":        "Kiro",
	"antigravity": "Antigravity",
}

// fileTypePatterns categorizes files by their path/name.
var fileTypePatterns = map[string]*regexp.Regexp{
	"config":   regexp.MustCompile(`(?i)\.(json|yaml|yml|toml)$`),
	"agent":    regexp.MustCompile(`(?i)agent|\.ai-|copilot-?instruction`),
	"rule":     regexp.MustCompile(`(?i)rule|constraint|guideline`),
	"template": regexp.MustCompile(`(?i)template|prompt`),
	"command":  regexp.MustCompile(`(?i)command|skill`),
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// DetectSetup scans sourceDir for existing AI tool configurations and returns
// detection results sorted by confidence (highest first).
func DetectSetup(sourceDir string) ([]DetectionResult, error) {
	// Check if source exists.
	if !files.DirExists(sourceDir) {
		return nil, nil
	}

	var results []DetectionResult

	for adapterID, patterns := range DetectionPatterns {
		result := detectAdapter(sourceDir, adapterID, patterns)
		if result.Detected {
			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	return results, nil
}

// DetectAdapters returns the adapter IDs detected in sourceDir.
func DetectAdapters(sourceDir string) []string {
	results, err := DetectSetup(sourceDir)
	if err != nil {
		return nil
	}
	var ids []string
	for _, r := range results {
		if r.Confidence > 0.3 {
			ids = append(ids, r.AdapterID)
		}
	}
	return ids
}

// HasAdapter checks whether a specific adapter's configuration exists in sourceDir.
func HasAdapter(sourceDir, adapterID string) bool {
	patterns, ok := DetectionPatterns[adapterID]
	if !ok {
		return false
	}
	result := detectAdapter(sourceDir, adapterID, patterns)
	return result.Detected && result.Confidence > 0.3
}

// DetectClaudeSetup checks for .claude/ directory or CLAUDE.md.
func DetectClaudeSetup(dir string) bool {
	return HasAdapter(dir, "claude-code")
}

// DetectCursorSetup checks for .cursor/ directory (mapped to generic).
func DetectCursorSetup(dir string) bool {
	return files.DirExists(filepath.Join(dir, ".cursor"))
}

// DetectOpenCodeSetup checks for .opencode/ directory or AGENTS.md.
func DetectOpenCodeSetup(dir string) bool {
	return HasAdapter(dir, "opencode")
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func detectAdapter(sourceDir, adapterID string, patterns []string) DetectionResult {
	var detectedFiles []DetectedFile

	for _, pattern := range patterns {
		fullPath := filepath.Join(sourceDir, pattern)

		// Check if it's a direct file or directory.
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			// Walk the directory for files.
			filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				rel, err := filepath.Rel(sourceDir, path)
				if err != nil {
					rel = path
				}
				detectedFiles = append(detectedFiles, DetectedFile{
					Path:     rel,
					Type:     categorizeFile(rel),
					Priority: calculatePriority(rel, adapterID),
				})
				return nil
			})
		} else {
			rel, err := filepath.Rel(sourceDir, fullPath)
			if err != nil {
				rel = pattern
			}
			detectedFiles = append(detectedFiles, DetectedFile{
				Path:     rel,
				Type:     categorizeFile(rel),
				Priority: calculatePriority(rel, adapterID),
			})
		}
	}

	// Sort by priority descending.
	sort.Slice(detectedFiles, func(i, j int) bool {
		return detectedFiles[i].Priority > detectedFiles[j].Priority
	})

	confidence := calculateConfidence(detectedFiles)

	name := adapterID
	if n, ok := AdapterNames[adapterID]; ok {
		name = n
	}

	return DetectionResult{
		Detected:    len(detectedFiles) > 0,
		Confidence:  confidence,
		AdapterID:   adapterID,
		AdapterName: name,
		Files:       detectedFiles,
		Metadata:    map[string]any{},
	}
}

func categorizeFile(filePath string) DetectedFileType {
	lower := strings.ToLower(filePath)
	for typ, re := range fileTypePatterns {
		if re.MatchString(lower) {
			return DetectedFileType(typ)
		}
	}
	return FileTypeOther
}

func calculatePriority(filePath, _ string) int {
	priority := 0
	lower := strings.ToLower(filePath)

	// Root config files are highest priority.
	rootConfigPatterns := []string{"agents.md", "claude.md"}
	for _, p := range rootConfigPatterns {
		if strings.EqualFold(filepath.Base(filePath), p) {
			priority += 100
			break
		}
	}

	// Copilot instructions.
	if strings.HasSuffix(lower, "copilot-instructions.md") {
		priority += 90
	}

	// Agent definitions.
	if strings.Contains(lower, "agent") {
		priority += 50
	}

	// Rules and standards.
	if strings.Contains(lower, "rule") || strings.Contains(lower, "standard") {
		priority += 40
	}

	// Commands/skills.
	if strings.Contains(lower, "command") || strings.Contains(lower, "skill") {
		priority += 30
	}

	// Templates.
	if strings.Contains(lower, "template") {
		priority += 20
	}

	return priority
}

func calculateConfidence(files []DetectedFile) float64 {
	if len(files) == 0 {
		return 0
	}

	// Base confidence on number of files.
	confidence := float64(len(files)) / 5.0
	if confidence > 0.8 {
		confidence = 0.8
	}

	// Boost if we have high-priority files.
	for _, f := range files {
		if f.Priority >= 50 {
			confidence += 0.15
			break
		}
	}

	// Boost if we have root config.
	for _, f := range files {
		base := strings.ToLower(filepath.Base(f.Path))
		if base == "agents.md" || base == "claude.md" ||
			strings.Contains(base, "copilot-instructions") {
			confidence += 0.1
			break
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}
