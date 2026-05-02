package diffviewer

import "testing"

func TestComputeDiffComputesLineChanges(t *testing.T) {
	t.Parallel()

	original := []byte("line1\nline2\nline3")
	modified := []byte("line1\nchanged\nline3\nline4")

	lines := ComputeDiff(original, modified)

	var additions, removals, context int
	for _, line := range lines {
		switch line.Type {
		case DiffLineAdded:
			additions++
		case DiffLineRemoved:
			removals++
		case DiffLineContext:
			context++
		}
	}

	if additions != 2 {
		t.Fatalf("expected 2 additions, got %d in %#v", additions, lines)
	}
	if removals != 1 {
		t.Fatalf("expected 1 removal, got %d in %#v", removals, lines)
	}
	if context != 2 {
		t.Fatalf("expected 2 context lines, got %d in %#v", context, lines)
	}
	if !HasDiffs(lines) {
		t.Fatal("expected HasDiffs to report changes")
	}
}

func TestComputeDiffResultStats(t *testing.T) {
	t.Parallel()

	result := ComputeDiffResult([]byte("old\nsame"), []byte("new\nsame"))

	if result.Stats.Additions != 1 {
		t.Fatalf("expected 1 addition, got %d", result.Stats.Additions)
	}
	if result.Stats.Deletions != 1 {
		t.Fatalf("expected 1 deletion, got %d", result.Stats.Deletions)
	}
	if result.Stats.Unchanged != 1 {
		t.Fatalf("expected 1 unchanged line, got %d", result.Stats.Unchanged)
	}
}

func TestRenderSimpleDiff(t *testing.T) {
	t.Parallel()

	got := RenderSimpleDiff([]DiffLine{
		{Type: DiffLineContext, Content: "same"},
		{Type: DiffLineRemoved, Content: "old"},
		{Type: DiffLineAdded, Content: "new"},
	})
	want := "  same\n- old\n+ new\n"

	if got != want {
		t.Fatalf("RenderSimpleDiff() = %q, want %q", got, want)
	}
}
