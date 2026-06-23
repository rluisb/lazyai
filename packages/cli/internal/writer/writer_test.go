package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/lockfile"
	"github.com/rluisb/lazyai/packages/cli/internal/plan"
)

func diskReader(root string) plan.DiskReader {
	return func(path string) ([]byte, bool) {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, false
		}
		return b, true
	}
}

func TestApplyCreatesFileAndLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	p := plan.Build([]plan.Desired{{Target: "claude", Path: path, SourceHash: "s1", Content: []byte("hello")}}, &lockfile.Lock{}, diskReader(dir))
	lock, results, err := Apply(p, &lockfile.Lock{}, Options{LazyaiVersion: "0.1.0"})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !results[0].Wrote {
		t.Fatal("expected write")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "hello" {
		t.Fatalf("content = %q", got)
	}
	if _, ok := lock.Find(path); !ok {
		t.Fatal("lock entry missing")
	}
}

func TestManagedPreservesUserContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(path, []byte("# My Project\n\nuser notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := plan.Build([]plan.Desired{{Target: "opencode", Path: path, SourceHash: "s1", Content: []byte("LAZYAI MANAGED BODY"), Managed: true}}, &lockfile.Lock{}, diskReader(dir))
	if p.Writes[0].Action != plan.Update {
		t.Fatalf("want update(adopt), got %s", p.Writes[0].Action)
	}
	if _, _, err := Apply(p, &lockfile.Lock{}, Options{}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	got, _ := os.ReadFile(path)
	s := string(got)
	if !strings.Contains(s, "# My Project") || !strings.Contains(s, "user notes") {
		t.Fatalf("user content lost: %q", s)
	}
	if !strings.Contains(s, "LAZYAI MANAGED BODY") || !strings.Contains(s, StartMarker) {
		t.Fatalf("managed region missing: %q", s)
	}
}

func TestDriftRefusedWithoutForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(path, []byte("user authored"), 0o644)
	p := plan.Build([]plan.Desired{{Target: "claude", Path: path, SourceHash: "s1", Content: []byte("generated")}}, &lockfile.Lock{}, diskReader(dir))
	_, results, err := Apply(p, &lockfile.Lock{}, Options{})
	if err == nil {
		t.Fatal("want drift error without force")
	}
	if results[0].Wrote {
		t.Fatal("must not overwrite on drift")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "user authored" {
		t.Fatalf("file overwritten: %q", got)
	}
}

func TestDriftAppliedWithForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(path, []byte("user authored"), 0o644)
	p := plan.Build([]plan.Desired{{Target: "claude", Path: path, SourceHash: "s1", Content: []byte("generated")}}, &lockfile.Lock{}, diskReader(dir))
	_, results, err := Apply(p, &lockfile.Lock{}, Options{Force: true})
	if err != nil {
		t.Fatalf("force apply: %v", err)
	}
	if !results[0].Wrote {
		t.Fatal("force should write")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "generated" {
		t.Fatalf("force write failed: %q", got)
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	p := plan.Build([]plan.Desired{{Target: "claude", Path: path, SourceHash: "s1", Content: []byte("x")}}, &lockfile.Lock{}, diskReader(dir))
	_, results, err := Apply(p, &lockfile.Lock{}, Options{DryRun: true})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if results[0].Wrote {
		t.Fatal("dry-run must not write")
	}
	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatal("dry-run created a file")
	}
}

// Idempotency (FR-004): a second compile after a clean apply is a full no-op.
func TestSecondRunIsNoOp(t *testing.T) {
	dir := t.TempDir()
	wholePath := filepath.Join(dir, "CLAUDE.md")
	managedPath := filepath.Join(dir, "AGENTS.md")
	desired := []plan.Desired{
		{Target: "claude", Path: wholePath, SourceHash: "s1", Content: []byte("body")},
		{Target: "opencode", Path: managedPath, SourceHash: "s2", Content: []byte("managed body"), Managed: true},
	}

	p1 := plan.Build(desired, &lockfile.Lock{}, diskReader(dir))
	lock, _, err := Apply(p1, &lockfile.Lock{}, Options{LazyaiVersion: "0.1.0"})
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}

	p2 := plan.Build(desired, lock, diskReader(dir))
	if p2.Count(plan.Skip) != len(desired) {
		var actions []string
		for _, w := range p2.Writes {
			actions = append(actions, string(w.Action)+":"+w.Reason)
		}
		t.Fatalf("second run not all skips: %v", actions)
	}
}

func TestAtomicWriteDurability(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	// First write: new file via atomic write.
	if err := atomicWrite(path, []byte("first")); err != nil {
		t.Fatalf("first atomicWrite: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "first" {
		t.Fatalf("first write content = %q, want %q", got, "first")
	}

	// Overwrite: existing file replaced atomically.
	if err := atomicWrite(path, []byte("second")); err != nil {
		t.Fatalf("second atomicWrite: %v", err)
	}
	got, _ = os.ReadFile(path)
	if string(got) != "second" {
		t.Fatalf("second write content = %q, want %q", got, "second")
	}

	// Temp file must be cleaned up after success.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "out.txt" {
			t.Fatalf("stale temp file left behind: %s", e.Name())
		}
	}
}
