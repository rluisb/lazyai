package theme

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestLintNoHardcodedColors asserts that no Go file under packages/cli/
// outside `internal/theme/` (and outside `_test.go` files) constructs a
// `lipgloss.Color("#…")` literal.
//
// This implements SC-004 + ADR-001 / ADR-003 (single source of design truth).
//
// Today (issue #190 partial scope) the migration is incremental: 6 cmd files
// remain on the followup track per the issue's footer. Those files appear in
// `expectedRemainingHardcodedColorSites` with their exact site count. The
// test PASSES as long as:
//
//   - the file-set is exactly the allowlisted set (no new offenders);
//   - the per-file count is unchanged or zero (no regressions; partial
//     migration must be reflected by lowering the count or removing the entry).
//
// When a deferred file is fully migrated, REMOVE its row from the map and
// re-run the test — it should still pass with zero violations there.
//
// When all rows are removed, the lint becomes a hard "zero violations"
// assertion and the migration is complete.
// expectedRemainingHardcodedColorSites — see file header. This baseline
// reflects the state after the FR-007 (boundary.go) + FR-009 partial
// (cmd/init.go) + FR-010 (huh) migrations of #190's first wave. Subsequent
// PRs lower these counts as each cmd file is migrated; an entry going to 0
// MUST be removed from this map (the test enforces it).
// Files cleared in #190's first wave: cmd/init.go (5791cad), cmd/add.go,
// cmd/compile.go, cmd/doctor.go, cmd/info.go, cmd/status.go (this commit).
// FR-009 closed across all 6 high-traffic commands named in the issue.
//
// Remaining (deferred to followup PRs per #190's footer): cmd/create.go,
// cmd/eject.go, cmd/list.go, cmd/mcp_hints.go, cmd/server.go.
var expectedRemainingHardcodedColorSites = map[string]int{
	"cmd/create.go":    3,
	"cmd/eject.go":     3,
	"cmd/list.go":      5,
	"cmd/mcp_hints.go": 4,
	"cmd/server.go":    20,
}

func TestLintNoHardcodedColors(t *testing.T) {
	root := cliRoot(t)

	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep (rg) not found in PATH — skipping lint test")
	}

	cmd := exec.Command(
		"rg",
		`lipgloss\.Color\("#`,
		".",
		"--glob", "!*_test.go",
		"--glob", "!internal/theme/**",
		"--count",
	)
	cmd.Dir = root // glob patterns are relative to the search root (cwd)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		t.Fatalf("rg failed: %v", err)
	}

	found := make(map[string]int)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colon := strings.LastIndex(line, ":")
		if colon < 0 {
			continue
		}
		path := strings.TrimPrefix(line[:colon], "./")
		count, perr := strconv.Atoi(line[colon+1:])
		if perr != nil {
			continue
		}
		found[path] = count
	}

	// Unexpected offenders: files in `found` that aren't in the allowlist.
	var unexpected []string
	for f := range found {
		if _, ok := expectedRemainingHardcodedColorSites[f]; !ok {
			unexpected = append(unexpected, f)
		}
	}
	sort.Strings(unexpected)
	for _, f := range unexpected {
		t.Errorf("file %s contains hardcoded `lipgloss.Color(\"#…\")` literals (%d sites). Use internal/theme/ tokens instead. (Allowlist this file in lint_test.go only if it's intentionally deferred per #190.)", f, found[f])
	}

	// Allowlist regressions: files in `expectedRemainingHardcodedColorSites`
	// whose actual count is HIGHER than allowed.
	for f, allowed := range expectedRemainingHardcodedColorSites {
		got := found[f]
		if got > allowed {
			t.Errorf("file %s grew from %d to %d hardcoded color sites — regressing", f, allowed, got)
		}
		if got > 0 && got < allowed {
			// Partial migration: should we tighten? Inform the implementer
			// without failing — the contract is "no MORE than allowed."
			t.Logf("file %s: %d sites remain (allowance is %d) — consider lowering allowance or removing entry", f, got, allowed)
		}
		if got == 0 && allowed > 0 {
			t.Errorf("file %s has zero hardcoded color sites but allowlist still says %d — REMOVE the entry from expectedRemainingHardcodedColorSites in lint_test.go", f, allowed)
		}
	}
}

// TestLintNoRawHuhNewForm asserts that no Go file under packages/cli/
// outside `internal/theme/huh.go` calls `huh.NewForm(`. Every Huh form must
// go through `theme.NewForm(...)` (ADR-002) so the project theme is
// registered and rendering is consistent across the binary (FR-010).
//
// This is enforced STRICTLY — no allowlist, no exceptions. A failure means a
// new caller bypassed `theme.NewForm`.
func TestLintNoRawHuhNewForm(t *testing.T) {
	root := cliRoot(t)

	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep (rg) not found in PATH — skipping lint test")
	}

	cmd := exec.Command(
		"rg",
		`huh\.NewForm\(`,
		".",
		"--glob", "!*_test.go",
		"--glob", "!internal/theme/huh.go", // the one legitimate site
		"-l",
	)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		// rg exits 1 when no matches — that's the success case.
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		t.Errorf("file %s contains a raw `huh.NewForm(` call. Use `theme.NewForm(...)` instead so the project theme is registered (ADR-002, FR-010).", strings.TrimPrefix(path, "./"))
	}
}

// cliRoot returns the absolute path to packages/cli/.
func cliRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = .../packages/cli/internal/theme/lint_test.go
	root := filepath.Join(filepath.Dir(file), "..", "..")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return abs
}
