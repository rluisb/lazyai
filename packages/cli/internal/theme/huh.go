package theme

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// HuhTheme returns the project's huh form theme. Every `huh.Form` constructed
// in this binary MUST register this theme via `theme.NewForm(...)` (or
// `form.WithTheme(theme.HuhTheme())`) so select / multi-select / input /
// confirm / note rendering all match the design system. Forbidden by lint
// (see lint_test.go) to call `huh.NewForm(` outside `internal/theme/huh.go`.
//
// Implementation: returns a `huh.ThemeFunc` adapter that builds a `*huh.Styles`
// from `huh.ThemeBase(isDark)` and overlays project tokens onto the elements
// that matter visually (title, focused/blurred field accents, selectors,
// error indicators).
func HuhTheme() huh.Theme {
	return huh.ThemeFunc(buildHuhStyles)
}

// NewForm returns a new `*huh.Form` with the project theme already applied.
// This is the single-entry-point Huh constructor for the binary. ADR-002
// describes the rationale and the lint that enforces it.
func NewForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithTheme(HuhTheme())
}

// buildHuhStyles overlays project tokens onto huh's base styles. Both light
// and dark variants share the same accent palette today (the design system has
// no light variant — see spec.md §Out of Scope).
func buildHuhStyles(isDark bool) *huh.Styles {
	s := huh.ThemeBase(isDark)

	// Form-level chrome. (FormStyles only exposes `Base` in this Huh
	// version; we leave the form margin to Bubble Tea / terminal defaults.)
	s.FieldSeparator = s.FieldSeparator.Foreground(Dimmed)

	// Focused field — primary accent.
	s.Focused.Title = s.Focused.Title.Foreground(Primary).Bold(true)
	s.Focused.Description = s.Focused.Description.Foreground(Dimmed)
	s.Focused.SelectSelector = s.Focused.SelectSelector.Foreground(Highlight)
	s.Focused.MultiSelectSelector = s.Focused.MultiSelectSelector.Foreground(Highlight)
	s.Focused.SelectedOption = s.Focused.SelectedOption.Foreground(Success)
	s.Focused.SelectedPrefix = s.Focused.SelectedPrefix.Foreground(Success)
	s.Focused.UnselectedOption = s.Focused.UnselectedOption.Foreground(Text)
	s.Focused.UnselectedPrefix = s.Focused.UnselectedPrefix.Foreground(Dimmed)
	s.Focused.Option = s.Focused.Option.Foreground(Text)
	s.Focused.NextIndicator = s.Focused.NextIndicator.Foreground(Highlight)
	s.Focused.PrevIndicator = s.Focused.PrevIndicator.Foreground(Highlight)
	s.Focused.FocusedButton = s.Focused.FocusedButton.
		Foreground(lipgloss.Color("#000000")).
		Background(Primary)
	s.Focused.BlurredButton = s.Focused.BlurredButton.Foreground(Dimmed)
	s.Focused.ErrorIndicator = s.Focused.ErrorIndicator.Foreground(Error)
	s.Focused.ErrorMessage = s.Focused.ErrorMessage.Foreground(Error)
	s.Focused.NoteTitle = s.Focused.NoteTitle.Foreground(Secondary).Bold(true)
	s.Focused.Card = s.Focused.Card.BorderForeground(Dimmed)

	// Blurred (non-focused) field — dimmed.
	s.Blurred.Title = s.Blurred.Title.Foreground(Dimmed)
	s.Blurred.Description = s.Blurred.Description.Foreground(Dimmed)
	s.Blurred.Option = s.Blurred.Option.Foreground(Dimmed)
	s.Blurred.SelectedOption = s.Blurred.SelectedOption.Foreground(Dimmed)
	s.Blurred.UnselectedOption = s.Blurred.UnselectedOption.Foreground(Dimmed)
	s.Blurred.NoteTitle = s.Blurred.NoteTitle.Foreground(Dimmed)

	return s
}
