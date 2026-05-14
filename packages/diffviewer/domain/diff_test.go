package domain

import "testing"

func TestComputeDiffResultKeepsParsingPure(t *testing.T) {
	t.Parallel()

	result := ComputeDiffResult([]byte("same\nold"), []byte("same\nnew\nadded"))

	if result.Stats.Additions != 2 {
		t.Fatalf("expected 2 additions, got %d in %#v", result.Stats.Additions, result.Lines)
	}
	if result.Stats.Deletions != 1 {
		t.Fatalf("expected 1 deletion, got %d in %#v", result.Stats.Deletions, result.Lines)
	}
	if result.Stats.Unchanged != 1 {
		t.Fatalf("expected 1 unchanged line, got %d in %#v", result.Stats.Unchanged, result.Lines)
	}
	if !HasDiffs(result.Lines) {
		t.Fatal("expected parsed diff to report changes")
	}
}

func TestHunkStartsGroupsContiguousChanges(t *testing.T) {
	t.Parallel()

	diffLines := []DiffLine{
		{Type: DiffLineContext, Content: "before"},
		{Type: DiffLineRemoved, Content: "old one"},
		{Type: DiffLineAdded, Content: "new one"},
		{Type: DiffLineContext, Content: "between"},
		{Type: DiffLineAdded, Content: "new two"},
		{Type: DiffLineContext, Content: "after"},
	}

	got := HunkStarts(diffLines)
	want := []int{1, 4}
	if len(got) != len(want) {
		t.Fatalf("expected hunk starts %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected hunk starts %v, got %v", want, got)
		}
	}
}
