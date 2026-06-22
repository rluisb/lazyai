package plan

import (
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/lockfile"
)

func reader(files map[string][]byte) DiskReader {
	return func(path string) ([]byte, bool) {
		b, ok := files[path]
		return b, ok
	}
}

func TestBuildCreateWhenMissing(t *testing.T) {
	d := []Desired{{Target: "claude", Path: "/x/CLAUDE.md", SourceHash: "s1", Content: []byte("hi")}}
	p := Build(d, &lockfile.Lock{}, reader(nil))
	if p.Writes[0].Action != Create {
		t.Fatalf("want create, got %s", p.Writes[0].Action)
	}
}

func TestBuildSkipWhenUpToDateWholeFile(t *testing.T) {
	content := []byte("hello")
	path := "/x/CLAUDE.md"
	lock := &lockfile.Lock{}
	lock.Upsert(lockfile.Generated{Path: path, SourceHash: "s1", OutputHash: lockfile.HashBytes(content)})
	p := Build([]Desired{{Path: path, SourceHash: "s1", Content: content}}, lock, reader(map[string][]byte{path: content}))
	if p.Writes[0].Action != Skip {
		t.Fatalf("want skip, got %s (%s)", p.Writes[0].Action, p.Writes[0].Reason)
	}
}

func TestBuildUpdateOnSourceChange(t *testing.T) {
	path := "/x/CLAUDE.md"
	old := []byte("v1")
	lock := &lockfile.Lock{}
	lock.Upsert(lockfile.Generated{Path: path, SourceHash: "s1", OutputHash: lockfile.HashBytes(old)})
	p := Build([]Desired{{Path: path, SourceHash: "s2", Content: []byte("v2")}}, lock, reader(map[string][]byte{path: old}))
	if p.Writes[0].Action != Update {
		t.Fatalf("want update, got %s", p.Writes[0].Action)
	}
}

func TestBuildDriftOnUntrackedWholeFile(t *testing.T) {
	path := "/x/CLAUDE.md"
	p := Build([]Desired{{Path: path, SourceHash: "s1", Content: []byte("gen")}}, &lockfile.Lock{}, reader(map[string][]byte{path: []byte("user wrote this")}))
	if p.Writes[0].Action != Drift {
		t.Fatalf("want drift, got %s", p.Writes[0].Action)
	}
	if !p.HasDrift() {
		t.Fatal("HasDrift should be true")
	}
}

func TestBuildDriftWholeFileModifiedSinceCompile(t *testing.T) {
	path := "/x/CLAUDE.md"
	gen := []byte("generated")
	lock := &lockfile.Lock{}
	lock.Upsert(lockfile.Generated{Path: path, SourceHash: "s1", OutputHash: lockfile.HashBytes(gen)})
	// disk differs from lock => user edited it
	p := Build([]Desired{{Path: path, SourceHash: "s1", Content: gen}}, lock, reader(map[string][]byte{path: []byte("hand edited")}))
	if p.Writes[0].Action != Drift {
		t.Fatalf("want drift, got %s", p.Writes[0].Action)
	}
}

func TestBuildManagedAdoptAndReMerge(t *testing.T) {
	path := "/x/AGENTS.md"
	// untracked managed => adopt (update)
	p := Build([]Desired{{Path: path, SourceHash: "s1", Content: []byte("block"), Managed: true}}, &lockfile.Lock{}, reader(map[string][]byte{path: []byte("# user")}))
	if p.Writes[0].Action != Update {
		t.Fatalf("want update(adopt), got %s", p.Writes[0].Action)
	}
	// tracked managed, disk edited => re-merge (update), never drift
	lock := &lockfile.Lock{}
	lock.Upsert(lockfile.Generated{Path: path, SourceHash: "s1", OutputHash: lockfile.HashBytes([]byte("merged")), Managed: true})
	p2 := Build([]Desired{{Path: path, SourceHash: "s1", Content: []byte("block"), Managed: true}}, lock, reader(map[string][]byte{path: []byte("# user edited outside region")}))
	if p2.Writes[0].Action != Update {
		t.Fatalf("want update(re-merge), got %s", p2.Writes[0].Action)
	}
}

func TestBuildManagedSkipWhenClean(t *testing.T) {
	path := "/x/AGENTS.md"
	merged := []byte("# user\nmanaged")
	lock := &lockfile.Lock{}
	lock.Upsert(lockfile.Generated{Path: path, SourceHash: "s1", OutputHash: lockfile.HashBytes(merged), Managed: true})
	p := Build([]Desired{{Path: path, SourceHash: "s1", Content: []byte("managed"), Managed: true}}, lock, reader(map[string][]byte{path: merged}))
	if p.Writes[0].Action != Skip {
		t.Fatalf("want skip, got %s (%s)", p.Writes[0].Action, p.Writes[0].Reason)
	}
}
