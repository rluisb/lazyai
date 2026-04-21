package library

import (
	"testing"
	"testing/fstest"
)

func TestResolveGeminiCommandsSubdir_PrefersNewLayout(t *testing.T) {
	libFS := fstest.MapFS{
		"gemini/commands/commit.toml": &fstest.MapFile{Data: []byte("x = 1\n")},
		"commands/legacy.toml":        &fstest.MapFile{Data: []byte("x = 1\n")},
	}
	got := ResolveGeminiCommandsSubdir(libFS)
	if got != GeminiCommandsSubdir {
		t.Errorf("got %q, want %q (new layout must win when both present)", got, GeminiCommandsSubdir)
	}
}

func TestResolveGeminiCommandsSubdir_FallsBackToLegacy(t *testing.T) {
	libFS := fstest.MapFS{
		"commands/legacy.toml": &fstest.MapFile{Data: []byte("x = 1\n")},
	}
	got := ResolveGeminiCommandsSubdir(libFS)
	if got != GeminiCommandsSubdirLegacy {
		t.Errorf("got %q, want %q (legacy fallback)", got, GeminiCommandsSubdirLegacy)
	}
}

func TestResolveGeminiCommandsSubdir_ReturnsEmptyWhenNeitherPresent(t *testing.T) {
	libFS := fstest.MapFS{
		"agents/builder.md": &fstest.MapFile{Data: []byte("x\n")},
	}
	if got := ResolveGeminiCommandsSubdir(libFS); got != "" {
		t.Errorf("got %q, want empty string when no commands available", got)
	}
}

func TestResolveGeminiCommandsSubdir_IgnoresEmptyDirs(t *testing.T) {
	libFS := fstest.MapFS{
		// gemini/commands exists but is empty; legacy has content
		"commands/legacy.toml": &fstest.MapFile{Data: []byte("x = 1\n")},
	}
	got := ResolveGeminiCommandsSubdir(libFS)
	if got != GeminiCommandsSubdirLegacy {
		t.Errorf("got %q, want legacy when new layout is absent/empty", got)
	}
}

func TestResolveGeminiCommandsSubdir_NilFsReturnsDefault(t *testing.T) {
	got := ResolveGeminiCommandsSubdir(nil)
	if got != GeminiCommandsSubdir {
		t.Errorf("got %q, want %q for nil FS", got, GeminiCommandsSubdir)
	}
}
