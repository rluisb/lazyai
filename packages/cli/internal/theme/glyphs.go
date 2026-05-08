package theme

// Canonical Unicode glyphs used across the design system. Status communication
// in the TUI is always glyph + color together — never color alone (Article III
// of the inferred design-system constitution; mirrors the rule in
// .claude/skills/tui-lazy-ai-design-system/README.md §Iconography).
//
// Adding a glyph here without a corresponding entry in the design-system skill
// (or vice versa) is a defect on the side that has not yet caught up.
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
