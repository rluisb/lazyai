package diffviewer

import "charm.land/lipgloss/v2"

var (
	// Primary is the main brand colour used in headers.
	Primary = lipgloss.Color("#7D56F4")
	// Success indicates a positive outcome (green).
	Success = lipgloss.Color("#2ECC71")
	// Error indicates a failure or removed content (red).
	Error = lipgloss.Color("#E74C3C")
	// Dimmed is for de-emphasised or secondary text (gray).
	Dimmed = lipgloss.Color("#6C7086")
	// Highlight draws attention to keyboard shortcuts.
	Highlight = lipgloss.Color("#89B4FA")
	// Text is the default foreground colour.
	Text = lipgloss.Color("#CDD6F4")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	successStyle = lipgloss.NewStyle().
			Foreground(Success)

	errorStyle = lipgloss.NewStyle().
			Foreground(Error)

	keyStyle = lipgloss.NewStyle().
			Foreground(Highlight).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(Text)
)

// Title renders bold text in the primary colour.
func Title(text string) string {
	return titleStyle.Render(text)
}

// SuccessLabel renders text in green with a checkmark prefix.
func SuccessLabel(text string) string {
	return successStyle.Render("✓ " + text)
}

// ErrorLabel renders text in red with a cross prefix.
func ErrorLabel(text string) string {
	return errorStyle.Render("✗ " + text)
}

// KeyBadge renders a keyboard key badge (e.g. "[a]") in the highlight colour.
func KeyBadge(text string) string {
	return keyStyle.Render(text)
}

// ValueText renders value text in the default foreground colour.
func ValueText(text string) string {
	return valueStyle.Render(text)
}
