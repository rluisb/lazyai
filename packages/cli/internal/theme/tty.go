package theme

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
)

// profileFor returns the color profile that should be used when rendering to w.
//
// Behavior:
//   - For non-TTY writers (`bytes.Buffer`, files, pipes), returns
//     `colorprofile.Ascii` — no color, no escape codes.
//   - For TTY writers, returns the terminal's actual capability
//     (TrueColor / ANSI256 / ANSI / etc.).
//   - `NO_COLOR=1` (env) forces `Ascii` even on a TTY (honored by
//     colorprofile.Detect natively).
//   - Detection is per-call. lipgloss/colorprofile internally memoize the
//     result by file descriptor so the cost amortizes after the first call
//     per writer.
//
// `CLICOLOR_FORCE` is **not** honored — only `NO_COLOR`. Documented limitation.
func profileFor(w io.Writer) colorprofile.Profile {
	return colorprofile.Detect(w, os.Environ())
}

// wrapWriter wraps w with a colorprofile-aware writer that downsamples ANSI
// escape sequences to the destination's actual capability. Calls to Write on
// the wrapper produce plain text when the destination is not a TTY.
func wrapWriter(w io.Writer) io.Writer {
	return colorprofile.NewWriter(w, os.Environ())
}
