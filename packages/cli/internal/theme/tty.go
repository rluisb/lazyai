package theme

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
)

// wrapWriter wraps w with a colorprofile-aware writer that downsamples ANSI
// escape sequences to the destination's actual capability. Calls to Write on
// the wrapper produce plain text when the destination is not a TTY.
func wrapWriter(w io.Writer) io.Writer {
	return colorprofile.NewWriter(w, os.Environ())
}
