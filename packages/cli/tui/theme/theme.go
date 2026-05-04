// Package theme provides the centralized Lip Gloss theme and reusable style
// functions for the ai-setup CLI.
//
// All TUI-facing code should use this package so that colours, borders and
// spacing remain consistent across the application.
package theme

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// ─── Colour palette ──────────────────────────────────────────────────────────

var (
	// Primary is the main brand colour (purple).
	Primary = lipgloss.Color("#7D56F4")
	// Secondary is the accent colour (teal).
	Secondary = lipgloss.Color("#4ECDC4")
	// Success indicates a positive outcome (green).
	Success = lipgloss.Color("#2ECC71")
	// Warning calls attention to a non-fatal issue (yellow).
	Warning = lipgloss.Color("#F1C40F")
	// Error indicates a failure (red).
	Error = lipgloss.Color("#E74C3C")
	// Dimmed is for de-emphasised or secondary text (gray).
	Dimmed = lipgloss.Color("#6C7086")
	// Highlight draws attention to important text (blue).
	Highlight = lipgloss.Color("#89B4FA")
	// Text is the default foreground colour (light gray).
	Text = lipgloss.Color("#CDD6F4")
	// Background is the default dark background.
	Background = lipgloss.Color("#1E1E2E")

	// Orange is used for conflict indicators.
	Orange = lipgloss.Color("#E8912D")
)

// ─── Reusable styles ────────────────────────────────────────────────────────

var (
	baseStyle = lipgloss.NewStyle().Foreground(Text)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	successStyle = lipgloss.NewStyle().
			Foreground(Success)

	errorStyle = lipgloss.NewStyle().
			Foreground(Error)

	warningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	dimStyle = lipgloss.NewStyle().
			Foreground(Dimmed)

	highlightStyle = lipgloss.NewStyle().
			Foreground(Highlight)

	keyStyle = lipgloss.NewStyle().
			Foreground(Highlight).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(Text)

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true).
				Underline(true)

	dimStyleExport = lipgloss.NewStyle().
			Foreground(Dimmed)

	bulletStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	codeStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(lipgloss.Color("#313244")).
			Padding(0, 1)

	installedStyle = lipgloss.NewStyle().
			Foreground(Success)

	modifiedStyle = lipgloss.NewStyle().
			Foreground(Warning)

	missingStyle = lipgloss.NewStyle().
			Foreground(Error)

	conflictStyle = lipgloss.NewStyle().
			Foreground(Orange)

	pendingStyle = lipgloss.NewStyle().
			Foreground(Dimmed)
)

// ─── Style functions ────────────────────────────────────────────────────────

// Title renders bold text in the primary colour.
func Title(text string) string {
	return titleStyle.Render(text)
}

// Subtitle renders text in the secondary colour.
func Subtitle(text string) string {
	return subtitleStyle.Render(text)
}

// SuccessLabel renders text in green with a checkmark prefix.
func SuccessLabel(text string) string {
	return successStyle.Render("✓ " + text)
}

// ErrorLabel renders text in red with a cross prefix.
func ErrorLabel(text string) string {
	return errorStyle.Render("✗ " + text)
}

// WarningLabel renders text in yellow with a warning prefix.
func WarningLabel(text string) string {
	return warningStyle.Render("⚠ " + text)
}

// DimStyle returns the dimmed lipgloss.Style for use in width-based rendering.
func DimStyle() lipgloss.Style {
	return dimStyleExport
}

// DimText renders de-emphasised text.
func DimText(text string) string {
	return dimStyle.Render(text)
}

// KeyValue renders a key-value pair with the key highlighted.
func KeyValue(key, value string) string {
	return keyStyle.Render(key+":") + " " + valueStyle.Render(value)
}

// SectionHeader renders an underlined section header.
func SectionHeader(text string) string {
	return sectionHeaderStyle.Render(text)
}

// Bullet renders a bullet point item.
func Bullet(text string) string {
	return bulletStyle.Render("•") + " " + text
}

// CodeBlock renders code-style text with a subtle background.
func CodeBlock(text string) string {
	return codeStyle.Render(text)
}

// ─── Box styles ──────────────────────────────────────────────────────────────

// Box returns a generic bordered box style.
func Box() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Dimmed).
		Padding(0, 1)
}

// SuccessBox returns a green-bordered box style.
func SuccessBox() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Success).
		Padding(0, 1)
}

// ErrorBox returns a red-bordered box style.
func ErrorBox() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Error).
		Padding(0, 1)
}

// WarningBox returns a yellow-bordered box style.
func WarningBox() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Warning).
		Padding(0, 1)
}

// ─── Table helper ────────────────────────────────────────────────────────────

// CreateTable renders a simple table with styled headers using lipgloss.
// It does not depend on bubbles/table so it is suitable for static,
// non-interactive rendering (e.g. in command output).
func CreateTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Calculate column widths.
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Minimum column width of 3.
	for i := range colWidths {
		if colWidths[i] < 3 {
			colWidths[i] = 3
		}
	}

	// Render header row.
	headerCells := make([]string, len(headers))
	for i, h := range headers {
		headerCells[i] = keyStyle.Width(colWidths[i]).Render(h)
	}
	headerRow := strings.Join(headerCells, "  ")

	// Render separator.
	separators := make([]string, len(headers))
	for i := range headers {
		separators[i] = dimStyle.Render(strings.Repeat("─", colWidths[i]))
	}
	separatorRow := strings.Join(separators, "  ")

	// Render data rows.
	dataRows := make([]string, 0, len(rows))
	for _, row := range rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			style := baseStyle
			if i < len(headers) {
				cells[i] = style.Width(colWidths[i]).Render(cell)
			} else {
				cells[i] = style.Render(cell)
			}
		}
		// Pad shorter rows.
		for i := len(row); i < len(headers); i++ {
			cells = append(cells, baseStyle.Width(colWidths[i]).Render(""))
		}
		dataRows = append(dataRows, strings.Join(cells, "  "))
	}

	allRows := []string{headerRow, separatorRow}
	allRows = append(allRows, dataRows...)

	return strings.Join(allRows, "\n")
}

// ─── Progress display ────────────────────────────────────────────────────────

// ProgressBar renders a text-based progress bar.
// current and total represent progress; width is the bar width in characters.
func ProgressBar(current, total, width int) string {
	if total <= 0 {
		total = 1
	}
	if width <= 0 {
		width = 20
	}

	pct := float64(current) / float64(total)
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}

	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	styledBar := successStyle.Render(bar[:filled]) + dimStyle.Render(bar[filled:])

	label := fmt.Sprintf(" %d/%d", current, total)

	return styledBar + label
}

// ─── Status indicators ───────────────────────────────────────────────────────

// KeyBadge renders a keyboard key badge (e.g. "[a]") in the highlight colour.
func KeyBadge(text string) string {
	return keyStyle.Render(text)
}

// ValueText renders value text in the default foreground colour.
func ValueText(text string) string {
	return valueStyle.Render(text)
}

// StatusInstalled renders text with a green checkmark.
func StatusInstalled(text string) string {
	return installedStyle.Render("✓") + " " + text
}

// StatusModified renders text with a yellow warning indicator.
func StatusModified(text string) string {
	return modifiedStyle.Render("⚠") + " " + text
}

// StatusMissing renders text with a red cross.
func StatusMissing(text string) string {
	return missingStyle.Render("✗") + " " + text
}

// StatusConflict renders text with an orange lightning bolt.
func StatusConflict(text string) string {
	return conflictStyle.Render("⚡") + " " + text
}

// StatusPending renders text with a dimmed circle.
func StatusPending(text string) string {
	return pendingStyle.Render("○") + " " + text
}
