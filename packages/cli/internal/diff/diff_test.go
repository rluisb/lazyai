package diff

import (
	"testing"
)

func TestComputeDiff_DetectsAdditions(t *testing.T) {
	t.Parallel()

	original := []byte("line1\nline2\n")
	modified := []byte("line1\nline2\nline3\n")

	lines := ComputeDiff(original, modified)

	addedCount := 0
	for _, l := range lines {
		if l.Type == DiffLineAdded {
			addedCount++
		}
	}
	if addedCount == 0 {
		t.Error("expected additions in diff")
	}
}

func TestComputeDiff_DetectsRemovals(t *testing.T) {
	t.Parallel()

	original := []byte("line1\nline2\nline3\n")
	modified := []byte("line1\nline3\n")

	lines := ComputeDiff(original, modified)

	removedCount := 0
	for _, l := range lines {
		if l.Type == DiffLineRemoved {
			removedCount++
		}
	}
	if removedCount == 0 {
		t.Error("expected removals in diff")
	}
}

func TestComputeDiff_IdenticalContent_NoDiffs(t *testing.T) {
	t.Parallel()

	content := []byte("same\ncontent\nhere\n")
	lines := ComputeDiff(content, content)

	if HasDiffs(lines) {
		t.Error("HasDiffs should be false for identical content")
	}
}

func TestHasDiffs_TrueWhenChanges(t *testing.T) {
	t.Parallel()

	original := []byte("old\n")
	modified := []byte("new\n")

	lines := ComputeDiff(original, modified)
	if !HasDiffs(lines) {
		t.Error("HasDiffs should be true when content differs")
	}
}

func TestHasDiffs_FalseWhenEmpty(t *testing.T) {
	t.Parallel()

	lines := []DiffLine{}
	if HasDiffs(lines) {
		t.Error("HasDiffs should be false for empty diff")
	}
}

func TestHasDiffs_FalseWhenOnlyContext(t *testing.T) {
	t.Parallel()

	lines := []DiffLine{
		{Type: DiffLineContext, Content: "same"},
		{Type: DiffLineContext, Content: "lines"},
	}
	if HasDiffs(lines) {
		t.Error("HasDiffs should be false for context-only diff")
	}
}

func TestComputeDiffResult_Stats(t *testing.T) {
	t.Parallel()

	original := []byte("line1\nline2\nline3\n")
	modified := []byte("line1\nchanged\nline3\nline4\n")

	result := ComputeDiffResult(original, modified)

	if result.Stats.Additions == 0 && result.Stats.Deletions == 0 {
		t.Error("expected some additions or deletions")
	}
}

func TestComputeDiff_EmptyToContent(t *testing.T) {
	t.Parallel()

	original := []byte("")
	modified := []byte("new line\n")

	lines := ComputeDiff(original, modified)
	if !HasDiffs(lines) {
		t.Error("should detect diff from empty to content")
	}

	// At least one addition must exist.
	hasAdd := false
	for _, l := range lines {
		if l.Type == DiffLineAdded {
			hasAdd = true
			break
		}
	}
	if !hasAdd {
		t.Error("expected at least one addition")
	}
}

func TestComputeDiff_ContentToEmpty(t *testing.T) {
	t.Parallel()

	original := []byte("old line\n")
	modified := []byte("")

	lines := ComputeDiff(original, modified)
	if !HasDiffs(lines) {
		t.Error("should detect diff from content to empty")
	}
}

func TestRenderSimpleDiff(t *testing.T) {
	t.Parallel()

	lines := []DiffLine{
		{Type: DiffLineContext, Content: "same"},
		{Type: DiffLineRemoved, Content: "old"},
		{Type: DiffLineAdded, Content: "new"},
	}

	result := RenderSimpleDiff(lines)
	if result == "" {
		t.Error("RenderSimpleDiff returned empty string")
	}
}
