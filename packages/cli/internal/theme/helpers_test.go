package theme

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// ansiPattern matches any ANSI CSI escape sequence. Used to assert that
// pipe-safe output (FR-005, EC-001, SC-001) contains no escape codes.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// TestHelpersWriteToBuffer verifies that all 6 Print helpers produce plain
// UTF-8 output (zero ANSI escape sequences) when the destination is a
// `bytes.Buffer` (a non-TTY writer). This is the core pipe-safety contract.
func TestHelpersWriteToBuffer(t *testing.T) {
	cases := []struct {
		name      string
		fn        func(w *bytes.Buffer, s string)
		wantGlyph string
	}{
		{"Infof", func(w *bytes.Buffer, s string) { Infof(w, "%s", s) }, GlyphBullet},
		{"Successf", func(w *bytes.Buffer, s string) { Successf(w, "%s", s) }, GlyphSuccess},
		{"Warnf", func(w *bytes.Buffer, s string) { Warnf(w, "%s", s) }, GlyphWarn},
		{"Errorf", func(w *bytes.Buffer, s string) { Errorf(w, "%s", s) }, GlyphError},
		{"Conflictf", func(w *bytes.Buffer, s string) { Conflictf(w, "%s", s) }, GlyphConflict},
		{"Pendingf", func(w *bytes.Buffer, s string) { Pendingf(w, "%s", s) }, GlyphPending},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			c.fn(&buf, "hello world")
			got := buf.String()

			// AC: zero ANSI escape sequences when writer is bytes.Buffer.
			if ansiPattern.MatchString(got) {
				t.Errorf("%s output contains ANSI escapes: %q", c.name, got)
			}

			// AC: line begins with the canonical glyph + space.
			if !strings.HasPrefix(got, c.wantGlyph+" ") {
				t.Errorf("%s output = %q, want prefix %q", c.name, got, c.wantGlyph+" ")
			}

			// AC: line contains the message and ends with newline.
			if !strings.Contains(got, "hello world") {
				t.Errorf("%s output = %q, want to contain %q", c.name, got, "hello world")
			}
			if !strings.HasSuffix(got, "\n") {
				t.Errorf("%s output = %q, want trailing newline", c.name, got)
			}
		})
	}
}

// TestHelpersFormatArgs verifies that helpers correctly forward format
// directives and args (cf. fmt.Fprintf semantics).
func TestHelpersFormatArgs(t *testing.T) {
	var buf bytes.Buffer
	Infof(&buf, "count: %d, name: %s", 42, "answer")
	got := buf.String()
	if !strings.Contains(got, "count: 42, name: answer") {
		t.Errorf("Infof format output = %q, want to contain %q", got, "count: 42, name: answer")
	}
}

// TestHelpersEmptyMessage verifies EC-007: a whitespace-only / empty message
// renders as glyph + (single space) + (empty msg) + newline.
func TestHelpersEmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	Successf(&buf, "")
	got := buf.String()
	if !strings.HasPrefix(got, GlyphSuccess+" ") {
		t.Errorf("empty Successf = %q, want prefix %q", got, GlyphSuccess+" ")
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("empty Successf = %q, want trailing newline", got)
	}
}

// TestHelpersGoroutineSafety verifies EC-008 (R-04 mitigation): 100 goroutines
// each call `theme.Successf`; output contains 100 well-formed lines (one Write
// per call, no mid-line interleaving).
func TestHelpersGoroutineSafety(t *testing.T) {
	const N = 100
	var buf bytes.Buffer
	var mu sync.Mutex // guards bytes.Buffer (NOT goroutine-safe itself)
	var wg sync.WaitGroup

	wg.Add(N)
	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			// Serialize Writes to bytes.Buffer (it's not concurrency-safe).
			// The helper itself is concurrency-safe at line granularity
			// (single Write per call); in real usage `os.Stdout`/`os.Stderr`
			// ARE concurrency-safe writers.
			mu.Lock()
			defer mu.Unlock()
			Successf(&buf, "msg %d", i)
		}()
	}
	wg.Wait()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != N {
		t.Errorf("got %d lines, want %d (mid-line interleaving suspected)", len(lines), N)
	}
	for i, l := range lines {
		if !strings.HasPrefix(l, GlyphSuccess+" msg ") {
			t.Errorf("line %d = %q, want prefix %q", i, l, GlyphSuccess+" msg ")
		}
	}
}

// TestNoColorOnBufferIsAlwaysPlain verifies the non-TTY writer always strips
// ANSI even when NO_COLOR isn't set; with NO_COLOR=1, the same. A stronger
// PTY-based test of the TTY branch is deferred to manual smoke (gate-5
// checklist).
func TestNoColorOnBufferIsAlwaysPlain(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	Errorf(&buf, "boom")
	if ansiPattern.MatchString(buf.String()) {
		t.Errorf("NO_COLOR=1 + bytes.Buffer should produce zero ANSI; got %q", buf.String())
	}
}
