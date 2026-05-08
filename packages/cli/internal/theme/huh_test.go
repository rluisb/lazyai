package theme

import (
	"fmt"
	"testing"

	"charm.land/huh/v2"
)

// TestHuhThemeNonNil verifies `theme.HuhTheme()` returns a non-nil interface
// value and a non-nil styles object. The returned theme is a `huh.ThemeFunc`
// adapter that builds a `*huh.Styles` from the project tokens.
func TestHuhThemeNonNil(t *testing.T) {
	th := HuhTheme()
	if th == nil {
		t.Fatal("HuhTheme() returned nil")
	}
	styles := th.Theme(true)
	if styles == nil {
		t.Fatal("HuhTheme().Theme(true) returned nil")
	}
}

// TestHuhThemeOverridesBase verifies the project theme actually changes
// styles vs `huh.ThemeBase`. If this fails, the project theme is doing nothing
// — wizard forms would render with Huh defaults regardless of WithTheme.
//
// We compare by stringifying `color.Color` (which `lipgloss.Style.GetForeground`
// returns) via `fmt.Sprintf("%v", ...)`. For the same color, the sprintf
// representation is stable; for different colors, it differs.
func TestHuhThemeOverridesBase(t *testing.T) {
	base := huh.ThemeBase(true)
	project := HuhTheme().Theme(true)

	cases := []struct {
		name       string
		baseValue  string
		projValue  string
	}{
		{
			"Focused.Title.Foreground",
			fmt.Sprintf("%v", base.Focused.Title.GetForeground()),
			fmt.Sprintf("%v", project.Focused.Title.GetForeground()),
		},
		{
			"Focused.SelectSelector.Foreground",
			fmt.Sprintf("%v", base.Focused.SelectSelector.GetForeground()),
			fmt.Sprintf("%v", project.Focused.SelectSelector.GetForeground()),
		},
		{
			"Focused.SelectedOption.Foreground",
			fmt.Sprintf("%v", base.Focused.SelectedOption.GetForeground()),
			fmt.Sprintf("%v", project.Focused.SelectedOption.GetForeground()),
		},
		{
			"Focused.ErrorMessage.Foreground",
			fmt.Sprintf("%v", base.Focused.ErrorMessage.GetForeground()),
			fmt.Sprintf("%v", project.Focused.ErrorMessage.GetForeground()),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.baseValue == c.projValue {
				t.Errorf("%s unchanged from base (%q == %q) — project theme is not overriding", c.name, c.baseValue, c.projValue)
			}
		})
	}
}

// TestNewFormReturnsForm verifies `theme.NewForm(...)` constructs a usable
// `*huh.Form` for each of the 5 form variants the project uses (Select,
// MultiSelect, Input, Confirm, Note). This is the R-01 spike, ported to a
// permanent test.
func TestNewFormReturnsForm(t *testing.T) {
	t.Run("zero groups", func(t *testing.T) {
		f := NewForm()
		if f == nil {
			t.Fatal("NewForm() returned nil")
		}
	})
	t.Run("Select variant", func(t *testing.T) {
		g := huh.NewGroup(
			huh.NewSelect[string]().
				Title("pick one").
				Options(huh.NewOption("a", "a"), huh.NewOption("b", "b")),
		)
		f := NewForm(g)
		if f == nil {
			t.Fatal("NewForm(Select) returned nil")
		}
	})
	t.Run("Input variant", func(t *testing.T) {
		g := huh.NewGroup(huh.NewInput().Title("name"))
		f := NewForm(g)
		if f == nil {
			t.Fatal("NewForm(Input) returned nil")
		}
	})
	t.Run("Confirm variant", func(t *testing.T) {
		g := huh.NewGroup(huh.NewConfirm().Title("are you sure?"))
		f := NewForm(g)
		if f == nil {
			t.Fatal("NewForm(Confirm) returned nil")
		}
	})
	t.Run("Note variant", func(t *testing.T) {
		g := huh.NewGroup(huh.NewNote().Title("info").Description("just a note"))
		f := NewForm(g)
		if f == nil {
			t.Fatal("NewForm(Note) returned nil")
		}
	})
	t.Run("MultiSelect variant", func(t *testing.T) {
		g := huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("pick many").
				Options(huh.NewOption("a", "a"), huh.NewOption("b", "b")),
		)
		f := NewForm(g)
		if f == nil {
			t.Fatal("NewForm(MultiSelect) returned nil")
		}
	})
}
