package lockfile

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestLoad_MissingFileReturnsDefault(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	expected := &Lock{
		Version:       SchemaVersion,
		Adapters:      map[string]AdapterLock{},
		Generated:     []Generated{},
		LazyaiVersion: "",
		CompiledAt:    "",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Load() = %#v, want %#v", got, expected)
	}
}

func TestLoad_MalformedJSONReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "lock.json"), []byte("{\n"), 0o644); err != nil {
		t.Fatalf("write malformed lock: %v", err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected malformed JSON to return an error")
	}
	if !strings.Contains(err.Error(), "unmarshal lockfile") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestLock_SaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lock := &Lock{
		Version:       "1.0",
		LazyaiVersion: "2.0.0",
		CompiledAt:    "2026-06-22T00:00:00Z",
		Adapters: map[string]AdapterLock{
			"opencode": {Version: "1.2.3", DocsSnapshot: "abc"},
		},
		Generated: []Generated{
			{Path: "zeta.txt", Target: "opencode", SourceHash: "sh2", OutputHash: "oh2", Managed: true},
			{Path: "alpha.txt", Target: "opencode", SourceHash: "sh1", OutputHash: "oh1", Managed: false},
		},
	}

	if err := lock.Save(dir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := *lock
	expected.Generated = append([]Generated(nil), expected.Generated...)
	sortGeneratedByPath(expected.Generated)

	if !reflect.DeepEqual(*got, expected) {
		t.Fatalf("round-trip mismatch: got %#v, want %#v", got, expected)
	}
}

func TestHashBytes(t *testing.T) {
	t.Parallel()

	got := HashBytes([]byte("abc"))
	const want = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("HashBytes() = %q, want %q", got, want)
	}
}

func TestUpsertReplacesOrAppendsWithoutSortingEveryInsert(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		initial       []Generated
		insert        Generated
		want          []Generated
		shouldReplace bool
	}{
		{
			name: "replace existing path and preserve length",
			initial: []Generated{
				{Path: "alpha.txt", Target: "opencode", SourceHash: "old", OutputHash: "old", Managed: true},
				{Path: "zeta.txt", Target: "opencode", SourceHash: "x", OutputHash: "x", Managed: false},
			},
			insert:        Generated{Path: "alpha.txt", Target: "copilot", SourceHash: "new", OutputHash: "new", Managed: false},
			shouldReplace: true,
			want: []Generated{
				{Path: "alpha.txt", Target: "copilot", SourceHash: "new", OutputHash: "new", Managed: false},
				{Path: "zeta.txt", Target: "opencode", SourceHash: "x", OutputHash: "x", Managed: false},
			},
		},
		{
			name: "append new path without sorting immediately",
			initial: []Generated{
				{Path: "zeta.txt", Target: "opencode", SourceHash: "x", OutputHash: "x", Managed: false},
			},
			insert: Generated{Path: "beta.txt", Target: "pi", SourceHash: "y", OutputHash: "y", Managed: true},
			want: []Generated{
				{Path: "zeta.txt", Target: "opencode", SourceHash: "x", OutputHash: "x", Managed: false},
				{Path: "beta.txt", Target: "pi", SourceHash: "y", OutputHash: "y", Managed: true},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := &Lock{Generated: append([]Generated(nil), tc.initial...)}
			l.Upsert(tc.insert)

			if len(l.Generated) != len(tc.want) {
				t.Fatalf("len = %d, want %d", len(l.Generated), len(tc.want))
			}
			if !reflect.DeepEqual(l.Generated, tc.want) {
				t.Fatalf("Upsert produced %#v, want %#v", l.Generated, tc.want)
			}
			if tc.shouldReplace && pathCount(l.Generated, tc.insert.Path) != 1 {
				t.Fatalf("expected single %q entry after replace", tc.insert.Path)
			}
			// Save performs the deterministic final sort; Upsert intentionally does
			// not sort on every insert.
		})
	}
}

func TestSaveWritesTrailingNewline(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	l := &Lock{}
	if err := l.Save(dir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepath.Join(dir, "lock.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("saved file is empty")
	}
	if data[len(data)-1] != '\n' {
		t.Fatalf("expected trailing newline, got %q", data[len(data)-1])
	}
}

func sortGeneratedByPath(in []Generated) {
	sort.Slice(in, func(i, j int) bool {
		return in[i].Path < in[j].Path
	})
}

func pathCount(entries []Generated, path string) int {
	count := 0
	for _, entry := range entries {
		if entry.Path == path {
			count++
		}
	}
	return count
}
