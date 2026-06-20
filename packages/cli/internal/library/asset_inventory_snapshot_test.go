package library

import (
	"io/fs"
	"sort"
	"strings"
	"testing"
)

// Snapshot of the canonical library asset inventory. This is intentionally a
// pinned list: adding or removing a canonical agent/skill MUST be a deliberate
// change that updates this snapshot, so accidental asset drift fails loudly in
// CI. Closes KNOWLEDGE_MAP "Snapshot tests for library assets".
var canonicalAgentsSnapshot = []string{
	"deployer.md",
	"evidence-verifier.md",
	"guide.md",
	"implementer.md",
	"planner.md",
	"researcher.md",
	"responder.md",
	"reviewer.md",
}

var canonicalSkillsSnapshot = []string{
	"codebase-exploration.md",
	"diagnose.md",
	"pr-review.md",
	"test-first-change.md",
}

func TestCanonicalAgentsInventorySnapshot(t *testing.T) {
	assertMarkdownInventory(t, "canonical/agents", canonicalAgentsSnapshot)
}

func TestCanonicalSkillsInventorySnapshot(t *testing.T) {
	assertMarkdownInventory(t, "canonical/skills", canonicalSkillsSnapshot)
}

// assertMarkdownInventory reads the top-level *.md files in dir (relative to the
// library FS root) and asserts the sorted list matches want exactly.
func assertMarkdownInventory(t *testing.T, dir string, want []string) {
	t.Helper()
	fsys := testLibFS()

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	var got []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		got = append(got, entry.Name())
	}
	sort.Strings(got)

	if len(got) != len(want) {
		t.Fatalf("%s inventory drift: got %v, want %v", dir, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s inventory drift at index %d: got %q, want %q (full got=%v)", dir, i, got[i], want[i], got)
		}
	}
}
