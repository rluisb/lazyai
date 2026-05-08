package theme

import (
	"fmt"
	"image/color"
	"io"

	"charm.land/lipgloss/v2"
)

// Print helpers (writer-first signatures, ADR-003).
//
// Each helper renders `<glyph> <colored msg>\n` to w. Output is plain UTF-8
// (no ANSI escape sequences) when w is not a TTY, when NO_COLOR=1 is set, or
// when the terminal lacks color support. The rendered string is written as
// exactly one Write call so output is atomic at line granularity (safe for
// concurrent goroutines without locks; see helpers_test.go).
//
// Use `theme.Infof(os.Stdout, "...")` for informational command output,
// `theme.Errorf(os.Stderr, "...")` for error lines. Callers that compose
// styled fragments into a larger rendering (e.g. table cells) should use the
// string-returning helpers in theme.go (`SuccessLabel`, `Bullet`, etc.)
// instead — those return `string` for embedding.
//
// Naming: helpers carry the `f` suffix to mirror Go stdlib conventions
// (`fmt.Errorf`, `log.Printf`) and to avoid colliding with the existing
// `lipgloss.Color`-typed package vars (`Success`, `Error`, etc.).

// renderLine pre-renders the styled line as a string, wraps the destination
// writer with a colorprofile-aware writer (ANSI is stripped/downsampled to
// match the writer's actual capability), then writes the line in a single
// Write call.
func renderLine(w io.Writer, glyph string, c color.Color, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	style := lipgloss.NewStyle().Foreground(c)
	line := style.Render(glyph+" "+msg) + "\n"

	cw := wrapWriter(w)
	_, _ = cw.Write([]byte(line))
}

// Infof renders a `• <msg>` info line in the secondary color.
func Infof(w io.Writer, format string, args ...any) {
	renderLine(w, GlyphBullet, Secondary, format, args...)
}

// Successf renders a `✓ <msg>` success line in the success color.
func Successf(w io.Writer, format string, args ...any) {
	renderLine(w, GlyphSuccess, Success, format, args...)
}

// Warnf renders a `⚠ <msg>` warning line in the warning color.
func Warnf(w io.Writer, format string, args ...any) {
	renderLine(w, GlyphWarn, Warning, format, args...)
}

// Errorf renders a `✗ <msg>` error line in the error color.
func Errorf(w io.Writer, format string, args ...any) {
	renderLine(w, GlyphError, Error, format, args...)
}

// Conflictf renders a `⚡ <msg>` conflict line in the orange color.
func Conflictf(w io.Writer, format string, args ...any) {
	renderLine(w, GlyphConflict, Orange, format, args...)
}

// Pendingf renders a `○ <msg>` pending/dimmed line in the dimmed color.
func Pendingf(w io.Writer, format string, args ...any) {
	renderLine(w, GlyphPending, Dimmed, format, args...)
}
