package components

import (
	"fmt"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/tui/theme"
)

// SummaryBox renders a styled summary of an operation (e.g. post-install
// or post-update report).
type SummaryBox struct {
	title     string
	successes []string
	warnings  []string
	errors    []string
	stats     summaryStats
	notes     string
}

type summaryStats struct {
	installed int
	modified  int
	skipped   int
}

// NewSummaryBox creates a SummaryBox with the given title.
func NewSummaryBox(title string) *SummaryBox {
	return &SummaryBox{
		title: title,
	}
}

// AddSuccess adds a success message to the summary.
func (s *SummaryBox) AddSuccess(message string) {
	s.successes = append(s.successes, message)
}

// AddWarning adds a warning message to the summary.
func (s *SummaryBox) AddWarning(message string) {
	s.warnings = append(s.warnings, message)
}

// AddError adds an error message to the summary.
func (s *SummaryBox) AddError(message string) {
	s.errors = append(s.errors, message)
}

// SetStats sets the count statistics displayed at the bottom of the box.
func (s *SummaryBox) SetStats(installed, modified, skipped int) {
	s.stats = summaryStats{
		installed: installed,
		modified:  modified,
		skipped:   skipped,
	}
}

// SetNotes sets the extra notes section displayed below the stats.
func (s *SummaryBox) SetNotes(notes string) {
	s.notes = notes
}

// Render returns the summary box as a styled string.
func (s *SummaryBox) Render() string {
	var lines []string

	// Determine box style based on content.
	boxStyle := theme.Box()
	if len(s.errors) > 0 {
		boxStyle = theme.ErrorBox()
	} else if len(s.warnings) > 0 {
		boxStyle = theme.WarningBox()
	} else if len(s.successes) > 0 {
		boxStyle = theme.SuccessBox()
	}

	// Title.
	lines = append(lines, theme.Title(s.title))
	lines = append(lines, "")

	// Success items.
	for _, msg := range s.successes {
		lines = append(lines, theme.StatusInstalled(msg))
	}

	// Warning items.
	for _, msg := range s.warnings {
		lines = append(lines, theme.StatusModified(msg))
	}

	// Error items.
	for _, msg := range s.errors {
		lines = append(lines, theme.StatusMissing(msg))
	}

	// Stats line (only if any stat is non-zero).
	if s.stats.installed > 0 || s.stats.modified > 0 || s.stats.skipped > 0 {
		lines = append(lines, "")
		parts := []string{}
		if s.stats.installed > 0 {
			parts = append(parts, fmt.Sprintf("installed: %d", s.stats.installed))
		}
		if s.stats.modified > 0 {
			parts = append(parts, fmt.Sprintf("modified: %d", s.stats.modified))
		}
		if s.stats.skipped > 0 {
			parts = append(parts, fmt.Sprintf("skipped: %d", s.stats.skipped))
		}
		lines = append(lines, theme.DimText(strings.Join(parts, "  ")))
	}

	if s.notes != "" {
		lines = append(lines, "")
		lines = append(lines, s.notes)
	}

	content := strings.Join(lines, "\n")
	return boxStyle.Render(content)
}
