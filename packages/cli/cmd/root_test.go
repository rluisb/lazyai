package cmd

import (
	"testing"

	buildversion "github.com/rluisb/lazyai/packages/cli/internal/version"
)

func TestSyncBuildVersionPropagatesCmdVersion(t *testing.T) {
	originalCmdVersion := Version
	originalBuildVersion := buildversion.Version
	t.Cleanup(func() {
		Version = originalCmdVersion
		buildversion.Version = originalBuildVersion
	})

	Version = "1.2.3"
	buildversion.Version = buildversion.DevVersion

	syncBuildVersion()

	if buildversion.Version != "1.2.3" {
		t.Fatalf("buildversion.Version = %q, want %q", buildversion.Version, "1.2.3")
	}
}

func TestSyncBuildVersionUsesInternalVersionWhenCmdVersionIsDev(t *testing.T) {
	originalCmdVersion := Version
	originalBuildVersion := buildversion.Version
	t.Cleanup(func() {
		Version = originalCmdVersion
		buildversion.Version = originalBuildVersion
	})

	Version = buildversion.DevVersion
	buildversion.Version = "2.3.4"

	syncBuildVersion()

	if Version != "2.3.4" {
		t.Fatalf("Version = %q, want %q", Version, "2.3.4")
	}
}
