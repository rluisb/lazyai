package version

import (
	"runtime/debug"
	"strings"
)

const DevVersion = "0.0.0-dev"

// Version is set at build time via ldflags.
var Version = DevVersion

func init() {
	Version = resolveVersion(Version, mainBuildInfoVersion())
}

func mainBuildInfoVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return info.Main.Version
}

func resolveVersion(current, mainVersion string) string {
	trimmed := strings.TrimSpace(current)
	if trimmed != "" && trimmed != DevVersion {
		return trimmed
	}

	buildInfoVersion := strings.TrimSpace(mainVersion)
	if buildInfoVersion == "" || buildInfoVersion == "(devel)" {
		if trimmed == "" {
			return DevVersion
		}
		return trimmed
	}

	return buildInfoVersion
}

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
