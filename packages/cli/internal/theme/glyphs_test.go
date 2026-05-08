package theme

import "testing"

// TestGlyphConstants verifies that every canonical glyph constant is the
// expected Unicode codepoint. Drift here means the design system + the TUI
// have desynchronized — either the constant is wrong, or the design-system
// skill at .claude/skills/tui-lazy-ai-design-system/README.md needs to be
// updated to match (FR-016). Failing this test is a defect on whichever side
// changed without co-update.
func TestGlyphConstants(t *testing.T) {
	cases := []struct {
		name  string
		glyph string
		want  string
	}{
		{"GlyphSuccess", GlyphSuccess, "✓"},
		{"GlyphError", GlyphError, "✗"},
		{"GlyphWarn", GlyphWarn, "⚠"},
		{"GlyphConflict", GlyphConflict, "⚡"},
		{"GlyphPending", GlyphPending, "○"},
		{"GlyphBullet", GlyphBullet, "•"},
		{"GlyphHRule", GlyphHRule, "─"},
		{"GlyphTreeBranch", GlyphTreeBranch, "├──"},
		{"GlyphTreeLast", GlyphTreeLast, "└──"},
		{"GlyphTreeVertical", GlyphTreeVertical, "│"},
		{"GlyphProgressFilled", GlyphProgressFilled, "█"},
		{"GlyphProgressEmpty", GlyphProgressEmpty, "░"},
	}
	for _, c := range cases {
		if c.glyph != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.glyph, c.want)
		}
	}
}

// TestSpinnerFrames verifies the canonical 10-frame Braille spinner sequence.
// Drift here breaks the spinner component visually and the design-system
// preview/progress-spinner.html card.
func TestSpinnerFrames(t *testing.T) {
	want := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if len(SpinnerFrames) != len(want) {
		t.Fatalf("SpinnerFrames length = %d, want %d", len(SpinnerFrames), len(want))
	}
	for i, w := range want {
		if SpinnerFrames[i] != w {
			t.Errorf("SpinnerFrames[%d] = %q, want %q", i, SpinnerFrames[i], w)
		}
	}
}
