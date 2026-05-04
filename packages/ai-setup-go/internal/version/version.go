package version

import "strings"

const DevVersion = "0.0.0-dev"

// Version is set at build time via ldflags.
var Version = DevVersion

// ReleaseTag returns the GitHub release tag for a known non-dev version.
func ReleaseTag(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}

	normalized := strings.TrimPrefix(trimmed, "v")
	if normalized == "" || strings.Contains(strings.ToLower(normalized), "dev") {
		return "", false
	}

	return "v" + normalized, true
}
