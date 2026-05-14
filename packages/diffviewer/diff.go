package diffviewer

import (
	"fmt"
	"os"
	"strings"

	"github.com/rluisb/lazyai/packages/diffviewer/domain"
)

// DiffLineType classifies a diff line.
type DiffLineType = domain.DiffLineType

const (
	DiffLineAdded   = domain.DiffLineAdded
	DiffLineRemoved = domain.DiffLineRemoved
	DiffLineContext = domain.DiffLineContext
)

// DiffLine represents a single line in a diff.
type DiffLine = domain.DiffLine

// DiffStats holds summary counts.
type DiffStats = domain.DiffStats

// DiffResult holds the full diff output.
type DiffResult = domain.DiffResult

// ComputeDiff computes a line-level diff between original and modified content
// using the LCS (Longest Common Subsequence) algorithm.
func ComputeDiff(original, modified []byte) []DiffLine {
	return domain.ComputeDiff(original, modified)
}

// ComputeDiffResult computes a full DiffResult with stats.
func ComputeDiffResult(original, modified []byte) DiffResult {
	return domain.ComputeDiffResult(original, modified)
}

// HasDiffs reports whether the diff contains any changes.
func HasDiffs(diffs []DiffLine) bool {
	return domain.HasDiffs(diffs)
}

// FormatDiff formats a diff as a string. If the output is a terminal, ANSI
// colors are used; otherwise plain text is produced.
func FormatDiff(diffs []DiffLine) string {
	useColor := isTerminal()
	var sb strings.Builder

	for _, line := range diffs {
		switch line.Type {
		case DiffLineAdded:
			if useColor {
				sb.WriteString("\x1b[32m+ ")
			} else {
				sb.WriteString("+ ")
			}
			sb.WriteString(line.Content)
			if useColor {
				sb.WriteString("\x1b[0m")
			}
		case DiffLineRemoved:
			if useColor {
				sb.WriteString("\x1b[31m- ")
			} else {
				sb.WriteString("- ")
			}
			sb.WriteString(line.Content)
			if useColor {
				sb.WriteString("\x1b[0m")
			}
		case DiffLineContext:
			if useColor {
				sb.WriteString("\x1b[2m  ")
			} else {
				sb.WriteString("  ")
			}
			sb.WriteString(line.Content)
			if useColor {
				sb.WriteString("\x1b[0m")
			}
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

// FormatDiffWithStats formats a diff with a header showing file path and stats.
func FormatDiffWithStats(diffs []DiffLine, filePath string) string {
	stats := DiffStats{}
	for _, line := range diffs {
		switch line.Type {
		case DiffLineAdded:
			stats.Additions++
		case DiffLineRemoved:
			stats.Deletions++
		case DiffLineContext:
			stats.Unchanged++
		}
	}

	var sb strings.Builder
	useColor := isTerminal()

	if filePath != "" {
		header := fmt.Sprintf("─── %s ───", filePath)
		if useColor {
			sb.WriteString("\x1b[2m")
		}
		sb.WriteString(header)
		if useColor {
			sb.WriteString("\x1b[0m")
		}
		sb.WriteByte('\n')
	}

	if stats.Additions > 0 || stats.Deletions > 0 {
		statsStr := fmt.Sprintf("+%d -%d", stats.Additions, stats.Deletions)
		if useColor {
			sb.WriteString("\x1b[2m")
		}
		sb.WriteString(statsStr)
		if useColor {
			sb.WriteString("\x1b[0m")
		}
		sb.WriteString("\n\n")
	}

	sb.WriteString(FormatDiff(diffs))
	return sb.String()
}

// RenderSimpleDiff produces a plain-text diff with +/- / space prefixes.
func RenderSimpleDiff(lines []DiffLine) string {
	var sb strings.Builder
	for _, line := range lines {
		switch line.Type {
		case DiffLineContext:
			sb.WriteString("  ")
		case DiffLineAdded:
			sb.WriteString("+ ")
		case DiffLineRemoved:
			sb.WriteString("- ")
		}
		sb.WriteString(line.Content)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// isTerminal reports whether stdout is a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode()&os.ModeCharDevice) != 0 || os.Getenv("FORCE_COLOR") != ""
}
