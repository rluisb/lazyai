package diffreview

import (
	"fmt"
	"strings"
)

const reviewContractVersion = 1

// Action is a user-selected per-file review decision.
type Action string

const (
	ActionAccept Action = "accept"
	ActionDeny   Action = "deny"
	ActionSkip   Action = "skip"
)

// ReviewStatus is the terminal status of an interactive review.
type ReviewStatus string

const (
	ReviewStatusConfirmed ReviewStatus = "confirmed"
	ReviewStatusCancelled ReviewStatus = "cancelled"
)

// ReviewRequest is the JSON v1 request payload supplied to lazyai-diffviewer.
type ReviewRequest struct {
	Version int        `json:"version"`
	Title   *string    `json:"title,omitempty"`
	Files   []FileDiff `json:"files"`
}

// FileDiff is a single file diff input in a review request.
type FileDiff struct {
	Path           string `json:"path"`
	CurrentContent string `json:"currentContent"`
	NewContent     string `json:"newContent"`
}

// ReviewResponse is the JSON v1 response payload emitted by lazyai-diffviewer.
type ReviewResponse struct {
	Version     int          `json:"version"`
	Status      ReviewStatus `json:"status"`
	Resolutions []Resolution `json:"resolutions"`
	Message     *string      `json:"message,omitempty"`
}

// Resolution records the user's decision for a single reviewed file.
type Resolution struct {
	Path   string `json:"path"`
	Action Action `json:"action"`
}

// DiffLineType classifies a diff line.
type DiffLineType string

const (
	DiffLineAdded   DiffLineType = "add"
	DiffLineRemoved DiffLineType = "remove"
	DiffLineContext DiffLineType = "context"
)

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type       DiffLineType
	Content    string
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

// ValidateReviewResponse verifies that a ReviewResponse satisfies the JSON v1 contract.
func ValidateReviewResponse(resp ReviewResponse) error {
	if resp.Version != reviewContractVersion {
		return fmt.Errorf("review response version must be %d, got %d", reviewContractVersion, resp.Version)
	}
	if !resp.Status.valid() {
		return fmt.Errorf("review response status must be %q or %q, got %q", ReviewStatusConfirmed, ReviewStatusCancelled, resp.Status)
	}
	if resp.Resolutions == nil {
		return fmt.Errorf("review response resolutions must be provided")
	}
	if resp.Message != nil && *resp.Message == "" {
		return fmt.Errorf("review response message must not be empty when provided")
	}

	for i, resolution := range resp.Resolutions {
		if resolution.Path == "" {
			return fmt.Errorf("review response resolutions[%d].path must not be empty", i)
		}
		if !resolution.Action.valid() {
			return fmt.Errorf("review response resolutions[%d].action must be %q, %q, or %q, got %q", i, ActionAccept, ActionDeny, ActionSkip, resolution.Action)
		}
	}

	return nil
}

func (a Action) valid() bool {
	switch a {
	case ActionAccept, ActionDeny, ActionSkip:
		return true
	default:
		return false
	}
}

func (s ReviewStatus) valid() bool {
	switch s {
	case ReviewStatusConfirmed, ReviewStatusCancelled:
		return true
	default:
		return false
	}
}

// ComputeDiffResult computes a full DiffResult with stats.
func ComputeDiffResult(original, modified []byte) DiffResult {
	lines := computeDiff(original, modified)
	stats := DiffStats{}
	for _, line := range lines {
		switch line.Type {
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

func computeDiff(original, modified []byte) []DiffLine {
	existingLines := strings.Split(string(original), "\n")
	incomingLines := strings.Split(string(modified), "\n")

	n := len(existingLines)
	m := len(incomingLines)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if existingLines[i-1] == incomingLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

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

	lines := make([]DiffLine, len(reversed))
	for idx, line := range reversed {
		lines[len(reversed)-1-idx] = line
	}
	return lines
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
