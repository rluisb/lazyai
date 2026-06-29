package theme

// Canonical Unicode glyphs used across the design system. Status communication
// in the TUI is always glyph + color together — never color alone.
//
// These constants are the source of truth for TUI iconography (FR-016); adding
// or changing a glyph here is a deliberate design-system change, and
// TestGlyphConstants guards their codepoints.
const (
	GlyphSuccess  = "✓"
	GlyphError    = "✗"
	GlyphWarn     = "⚠"
	GlyphConflict = "⚡"
	GlyphPending  = "○"
	GlyphBullet   = "•"
	GlyphHRule    = "─"

	GlyphTreeBranch   = "├──"
	GlyphTreeLast     = "└──"
	GlyphTreeVertical = "│"

	GlyphProgressFilled = "█"
	GlyphProgressEmpty  = "░"
)

// SpinnerFrames is the canonical 10-frame Braille spinner sequence used by the
// Bubble Tea TUI. Matches Bubbles' built-in `Dot` style and the design-system
// spinner card at preview/progress-spinner.html.
var SpinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}
