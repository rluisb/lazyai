package components

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

// Table creates a lipgloss-styled table for CLI output.
// This is a static, non-interactive table suitable for rendering in command
// output. For an interactive, selectable table, use the bubbles/table
// component directly.
type Table struct {
	headers   []string
	rows      [][]string
	colWidths []int
	style     tableStyle
}

type tableStyle struct {
	headerStyle lipgloss.Style
	cellStyle   lipgloss.Style
	borderStyle lipgloss.Border
}

// NewTable creates a new Table with the given column headers.
func NewTable(headers ...string) *Table {
	defaultBorder := lipgloss.NormalBorder()

	t := &Table{
		headers: headers,
		colWidths: func() []int {
			widths := make([]int, len(headers))
			for i, h := range headers {
				widths[i] = len(h)
			}
			return widths
		}(),
		style: tableStyle{
			headerStyle: lipgloss.NewStyle().
				Foreground(theme.Highlight).
				Bold(true),
			cellStyle: lipgloss.NewStyle().
				Foreground(theme.Text),
			borderStyle: defaultBorder,
		},
	}

	return t
}

// AddRow appends a row of cells to the table.
func (t *Table) AddRow(cells ...string) *Table {
	t.rows = append(t.rows, cells)
	// Update column widths.
	for i, cell := range cells {
		if i < len(t.colWidths) && len(cell) > t.colWidths[i] {
			t.colWidths[i] = len(cell)
		}
	}
	return t
}

// Render returns the table as a styled string.
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Minimum column width of 3.
	for i := range t.colWidths {
		if t.colWidths[i] < 3 {
			t.colWidths[i] = 3
		}
	}

	// Render header.
	headerCells := make([]string, len(t.headers))
	for i, h := range t.headers {
		headerCells[i] = t.style.headerStyle.Width(t.colWidths[i]).Render(h)
	}
	headerLine := strings.Join(headerCells, "  ")

	// Render separator.
	separators := make([]string, len(t.headers))
	for i := range t.headers {
		separators[i] = theme.DimStyle().Render(strings.Repeat("─", t.colWidths[i]))
	}
	separatorLine := strings.Join(separators, "  ")

	// Render rows.
	dataLines := make([]string, 0, len(t.rows))
	for _, row := range t.rows {
		cells := make([]string, len(t.headers))
		for i := range t.headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cells[i] = t.style.cellStyle.Width(t.colWidths[i]).Render(cell)
		}
		dataLines = append(dataLines, strings.Join(cells, "  "))
	}

	lines := []string{headerLine, separatorLine}
	lines = append(lines, dataLines...)
	return strings.Join(lines, "\n")
}
