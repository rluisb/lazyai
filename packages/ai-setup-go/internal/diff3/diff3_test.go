package diff3

import (
	"reflect"
	"strings"
	"testing"
)

func TestMyersDiff_BothEmpty(t *testing.T) {
	got := MyersDiff(nil, nil)
	if len(got.Added) != 0 || len(got.Removed) != 0 || len(got.Unchanged) != 0 {
		t.Errorf("expected all-empty result, got %+v", got)
	}
}

func TestMyersDiff_EmptyOld(t *testing.T) {
	got := MyersDiff(nil, []string{"a", "b"})
	if len(got.Added) != 2 {
		t.Fatalf("expected 2 added lines, got %d", len(got.Added))
	}
	if got.Added[0].Line != "a" || got.Added[0].Index != 0 {
		t.Errorf("added[0] = %+v, want {a, 0}", got.Added[0])
	}
}

func TestMyersDiff_EmptyNew(t *testing.T) {
	got := MyersDiff([]string{"x"}, nil)
	if len(got.Removed) != 1 {
		t.Fatalf("expected 1 removed line, got %d", len(got.Removed))
	}
	if got.Removed[0].Line != "x" {
		t.Errorf("removed[0] = %+v, want {x, 0}", got.Removed[0])
	}
}

func TestMyersDiff_SimpleChange(t *testing.T) {
	got := MyersDiff([]string{"a", "b", "c"}, []string{"a", "x", "c"})
	// b removed, x added, a and c unchanged
	if len(got.Removed) != 1 || got.Removed[0].Line != "b" {
		t.Errorf("removed = %+v, want single b", got.Removed)
	}
	if len(got.Added) != 1 || got.Added[0].Line != "x" {
		t.Errorf("added = %+v, want single x", got.Added)
	}
	if len(got.Unchanged) != 2 {
		t.Errorf("unchanged = %+v, want 2 lines", got.Unchanged)
	}
}

func TestDiff3_NoConflict_OnlyOursChanged(t *testing.T) {
	base := []string{"line1", "line2", "line3"}
	ours := []string{"line1", "line2-modified", "line3"}
	theirs := []string{"line1", "line2", "line3"}

	result := Diff3(base, ours, theirs)
	if result.HasConflicts {
		t.Errorf("expected no conflicts when only ours changed, got %+v", result.Conflicts)
	}
	if !contains(result.Merged, "line2-modified") {
		t.Errorf("merged missing ours change: %+v", result.Merged)
	}
}

func TestDiff3_NoConflict_OnlyTheirsChanged(t *testing.T) {
	base := []string{"line1", "line2", "line3"}
	ours := []string{"line1", "line2", "line3"}
	theirs := []string{"line1", "line2-other", "line3"}

	result := Diff3(base, ours, theirs)
	if result.HasConflicts {
		t.Errorf("expected no conflicts when only theirs changed, got %+v", result.Conflicts)
	}
	if !contains(result.Merged, "line2-other") {
		t.Errorf("merged missing theirs change: %+v", result.Merged)
	}
}

func TestDiff3_NoConflict_BothChangedSameWay(t *testing.T) {
	base := []string{"line1", "line2", "line3"}
	ours := []string{"line1", "line2-same", "line3"}
	theirs := []string{"line1", "line2-same", "line3"}

	result := Diff3(base, ours, theirs)
	if result.HasConflicts {
		t.Errorf("expected no conflicts when both changed identically, got %+v", result.Conflicts)
	}
}

func TestDiff3_Conflict_BothChangedDifferently(t *testing.T) {
	base := []string{"line1", "line2", "line3"}
	ours := []string{"line1", "OUR-line2", "line3"}
	theirs := []string{"line1", "THEIR-line2", "line3"}

	result := Diff3(base, ours, theirs)
	if !result.HasConflicts {
		t.Fatalf("expected a conflict, got merged=%+v", result.Merged)
	}
	joined := strings.Join(result.Merged, "\n")
	if !strings.Contains(joined, "<<<<<") || !strings.Contains(joined, ">>>>>") || !strings.Contains(joined, "=====") {
		t.Errorf("expected conflict markers, got:\n%s", joined)
	}
	if !strings.Contains(joined, "OUR-line2") || !strings.Contains(joined, "THEIR-line2") {
		t.Errorf("merged missing both conflicting versions:\n%s", joined)
	}
}

func TestMerge2Way_AppliesChangesOnEmpty(t *testing.T) {
	got := Merge2Way(nil, []string{"a", "b"})
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge2Way(nil, [a, b]) = %+v, want %+v", got, want)
	}
}

func TestMerge2Way_KeepsUnchangedLines(t *testing.T) {
	got := Merge2Way([]string{"a", "b", "c"}, []string{"a", "b", "c"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge2Way identical = %+v, want %+v", got, want)
	}
}

func TestHasConflictMarkers(t *testing.T) {
	if !HasConflictMarkers([]string{"x", "<<<<< OURS", "y"}) {
		t.Error("expected true for lines containing <<<<<")
	}
	if HasConflictMarkers([]string{"x", "y", "z"}) {
		t.Error("expected false for clean lines")
	}
	if !HasConflictMarkers([]string{"=====", "a"}) {
		t.Error("expected true for ===== marker")
	}
}

func TestResolveConflicts_TakesOurs(t *testing.T) {
	merged := []string{"line1", "<<<<< OURS", "OUR-line", "=====", "THEIR-line", ">>>>> THEIRS", "line3"}
	got := ResolveConflicts(merged, ResolveOurs)
	want := []string{"line1", "OUR-line", "line3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ResolveConflicts ours = %+v, want %+v", got, want)
	}
}

func TestResolveConflicts_TakesTheirs(t *testing.T) {
	merged := []string{"line1", "<<<<< OURS", "OUR-line", "=====", "THEIR-line", ">>>>> THEIRS", "line3"}
	got := ResolveConflicts(merged, ResolveTheirs)
	want := []string{"line1", "THEIR-line", "line3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ResolveConflicts theirs = %+v, want %+v", got, want)
	}
}

func TestResolveConflicts_NoConflict_Passthrough(t *testing.T) {
	merged := []string{"a", "b", "c"}
	got := ResolveConflicts(merged, ResolveOurs)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ResolveConflicts clean = %+v, want %+v", got, want)
	}
}

func contains(lines []string, target string) bool {
	for _, line := range lines {
		if line == target {
			return true
		}
	}
	return false
}
