package cmd

import (
	"testing"

	"golang.org/x/mod/semver"
)

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		input     string
		expected  string
		wantValid bool
	}{
		{"1.0.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", true},
		{" v2.0.0 ", "v2.0.0", true},
		{"10.0.0", "v10.0.0", true},
		{"0.9.0", "v0.9.0", true},
		{"1.0.0-alpha", "v1.0.0-alpha", true},
		{"release/train/v1.0.0", "release/train/v1.0.0", false},
	}
	for _, c := range cases {
		got := normalizeVersion(c.input)
		if got != c.expected {
			t.Errorf("normalizeVersion(%q) = %q, want %q", c.input, got, c.expected)
		}
		if c.wantValid && !semver.IsValid(got) {
			t.Errorf("normalizeVersion(%q) produced invalid semver %q", c.input, got)
		}
	}
}

func TestVersionComparison(t *testing.T) {
	// The original bug: string comparison treats "2.0.0" > "10.0.0".
	// semver.Compare must correctly order these.
	if semver.Compare(normalizeVersion("2.0.0"), normalizeVersion("10.0.0")) >= 0 {
		t.Errorf("expected 2.0.0 < 10.0.0, got semver.Compare >= 0")
	}

	cases := []struct {
		lesser  string
		greater string
	}{
		{"9.0.0", "10.0.0"},
		{"0.9.0", "0.10.0"},
		{"1.0.0-alpha", "1.0.0"},
	}
	for _, c := range cases {
		l := normalizeVersion(c.lesser)
		g := normalizeVersion(c.greater)
		if semver.Compare(l, g) >= 0 {
			t.Errorf("expected %s < %s, got semver.Compare(%q, %q) = %d",
				c.lesser, c.greater, l, g, semver.Compare(l, g))
		}
	}
}
