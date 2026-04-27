package detect

import (
	"runtime"
	"testing"
)

func TestCodexInstallHint_ReturnsNonEmpty(t *testing.T) {
	hint := CodexInstallHint()
	if hint == "" {
		t.Fatal("CodexInstallHint() returned empty string")
	}
}

func TestCodexInstallHint_PlatformSpecific(t *testing.T) {
	hint := CodexInstallHint()

	// The hint should contain platform-relevant instructions.
	// All platforms should at least mention npm as a fallback.
	if hint == "" {
		t.Fatal("expected non-empty install hint")
	}

	// Verify the hint matches the current platform.
	switch runtime.GOOS {
	case "darwin":
		// macOS hint should mention Homebrew.
		if !contains(hint, "brew") {
			t.Errorf("darwin hint should mention Homebrew, got: %s", hint)
		}
	case "linux":
		// Linux hint should mention npm or a download link.
		if !contains(hint, "npm") && !contains(hint, "github") {
			t.Errorf("linux hint should mention npm or github, got: %s", hint)
		}
	case "windows":
		if !contains(hint, "npm") && !contains(hint, "github") {
			t.Errorf("windows hint should mention npm or github, got: %s", hint)
		}
	}
}

func TestEnsureCodexOrPrompt_DoesNotPanic(t *testing.T) {
	// This should never panic regardless of whether codex is installed.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("EnsureCodexOrPrompt() panicked: %v", r)
		}
	}()

	err := EnsureCodexOrPrompt()
	if err != nil {
		t.Errorf("EnsureCodexOrPrompt() should always return nil, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
