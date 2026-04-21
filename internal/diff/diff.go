// Package diff provides line-level diff computation and formatting.
// Ported from the TypeScript utilities in src/utils/diff.ts.
package diff

import (
	"fmt"
	"os"
	"strings"
)

// DiffLineType classifies a diff line.
type DiffLineType string

const (
	DiffLineAdded   DiffLineType = "add"
	DiffLineRemoved DiffLineType = "remove"
	DiffLineContext DiffLineType = "context"
)

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type      DiffLineType
	Content   string
	OldLineNum int
	NewLineNum int
}

// DiffStats holds summary counts.
type DiffStats struct {
	Additions int
	Deletions int
	Unchanged int
}

// DiffResult holds the full diff output.
type DiffResult struct {
	Lines []DiffLine
	Stats DiffStats
}

// ComputeDiff computes a line-level diff between original and modified content
// using the LCS (Longest Common Subsequence) algorithm.
func ComputeDiff(original, modified []byte) []DiffLine {
	existingLines := strings.Split(string(original), "\n")
	incomingLines := strings.Split(string(modified), "\n")

	n := len(existingLines)
	m := len(incomingLines)

	// Build LCS table.
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if existingLines[i-1] == incomingLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] >= dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Backtrack to build diff.
	var reversed []DiffLine
	i, j := n, m
	oldLine, newLine := n, m

	for i > 0 && j > 0 {
		if existingLines[i-1] == incomingLines[j-1] {
			reversed = append(reversed, DiffLine{Type: DiffLineContext, Content: existingLines[i-1], OldLineNum: oldLine, NewLineNum: newLine})
			i--
			j--
			oldLine--
			newLine--
			continue
		}
		if dp[i-1][j] >= dp[i][j-1] {
			reversed = append(reversed, DiffLine{Type: DiffLineRemoved, Content: existingLines[i-1], OldLineNum: oldLine})
			i--
			oldLine--
		} else {
			reversed = append(reversed, DiffLine{Type: DiffLineAdded, Content: incomingLines[j-1], NewLineNum: newLine})
			j--
			newLine--
		}
	}

	for i > 0 {
		reversed = append(reversed, DiffLine{Type: DiffLineRemoved, Content: existingLines[i-1], OldLineNum: oldLine})
		i--
		oldLine--
	}

	for j > 0 {
		reversed = append(reversed, DiffLine{Type: DiffLineAdded, Content: incomingLines[j-1], NewLineNum: newLine})
		j--
		newLine--
	}

	// Reverse to get correct order.
	lines := make([]DiffLine, len(reversed))
	for idx, l := range reversed {
		lines[len(reversed)-1-idx] = l
	}

	return lines
}

// ComputeDiffResult computes a full DiffResult with stats.
func ComputeDiffResult(original, modified []byte) DiffResult {
	lines := ComputeDiff(original, modified)
	stats := DiffStats{}
	for _, l := range lines {
		switch l.Type {
		case DiffLineAdded:
			stats.Additions++
		case DiffLineRemoved:
			stats.Deletions++
		case DiffLineContext:
			stats.Unchanged++
		}
	}
	return DiffResult{Lines: lines, Stats: stats}
}

// HasDiffs reports whether the diff contains any changes.
func HasDiffs(diffs []DiffLine) bool {
	for _, l := range diffs {
		if l.Type != DiffLineContext {
			return true
		}
	}
	return false
}

// FormatDiff formats a diff as a string. If the output is a terminal, ANSI
// colors are used; otherwise plain text is produced.
func FormatDiff(diffs []DiffLine) string {
	useColor := isTerminal()
	var sb strings.Builder

	for _, l := range diffs {
		switch l.Type {
		case DiffLineAdded:
			if useColor {
				sb.WriteString("\x1b[32m+ ")
			} else {
				sb.WriteString("+ ")
			}
			sb.WriteString(l.Content)
			if useColor {
				sb.WriteString("\x1b[0m")
			}
		case DiffLineRemoved:
			if useColor {
				sb.WriteString("\x1b[31m- ")
			} else {
				sb.WriteString("- ")
			}
			sb.WriteString(l.Content)
			if useColor {
				sb.WriteString("\x1b[0m")
			}
		case DiffLineContext:
			if useColor {
				sb.WriteString("\x1b[2m  ")
			} else {
				sb.WriteString("  ")
			}
			sb.WriteString(l.Content)
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
	for _, l := range diffs {
		switch l.Type {
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
	for _, l := range lines {
		switch l.Type {
		case DiffLineContext:
			sb.WriteString("  ")
		case DiffLineAdded:
			sb.WriteString("+ ")
		case DiffLineRemoved:
			sb.WriteString("- ")
		}
		sb.WriteString(l.Content)
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
	return (fi.Mode() & os.ModeCharDevice) != 0 || os.Getenv("FORCE_COLOR") != ""
}