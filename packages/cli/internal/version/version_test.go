package version

import "testing"

func TestResolveVersionUsesBuildInfoVersionWhenCurrentVersionIsDev(t *testing.T) {
	got := resolveVersion(DevVersion, "v1.2.3")

	if got != "v1.2.3" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v1.2.3")
	}
}

func TestResolveVersionKeepsExplicitVersionWhenBuildInfoVersionExists(t *testing.T) {
	got := resolveVersion("2.3.4", "v1.2.3")

	if got != "2.3.4" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "2.3.4")
	}
}

func TestResolveVersionKeepsDevVersionWhenBuildInfoVersionIsDevel(t *testing.T) {
	got := resolveVersion(DevVersion, "(devel)")

	if got != DevVersion {
		t.Fatalf("resolveVersion() = %q, want %q", got, DevVersion)
	}
}

func TestResolveVersionUsesBuildInfoVersionWhenCurrentVersionIsEmpty(t *testing.T) {
	got := resolveVersion("", "v1.2.3")

	if got != "v1.2.3" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v1.2.3")
	}
}
