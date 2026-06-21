package diffviewer

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewDiffViewerComputesMissingDiffLines(t *testing.T) {
	t.Parallel()

	viewer := NewDiffViewer([]ConflictView{
		{
			FilePath:     "AGENTS.md",
			CurrentLines: []string{"same", "old"},
			NewLines:     []string{"same", "new"},
		},
	})

	if viewer == nil {
		t.Fatal("expected viewer")
	}
	if len(viewer.conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(viewer.conflicts))
	}
	if !HasDiffs(viewer.conflicts[0].DiffLines) {
		t.Fatalf("expected computed diff lines to contain changes: %#v", viewer.conflicts[0].DiffLines)
	}
	if viewer.currentIndex != 0 {
		t.Fatalf("expected current index 0, got %d", viewer.currentIndex)
	}
}

func TestDiffViewerResolveCurrentRecordsAction(t *testing.T) {
	t.Parallel()

	viewer := NewDiffViewer([]ConflictView{{FilePath: "rules.md"}})
	viewer.resolveCurrent(ActionAccept)

	if len(viewer.decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(viewer.decisions))
	}
	resolution := viewer.decisions[0]
	if resolution.Path != "rules.md" {
		t.Fatalf("expected path rules.md, got %q", resolution.Path)
	}
	if resolution.Action != ActionAccept {
		t.Fatalf("expected action %q, got %q", ActionAccept, resolution.Action)
	}
}

func TestDiffViewerResolveCurrentOverwritesRevisitedFileDecision(t *testing.T) {
	t.Parallel()

	viewer := NewDiffViewer([]ConflictView{
		{FilePath: "file-0.md"},
		{FilePath: "file-1.md"},
		{FilePath: "file-2.md"},
	})

	viewer.currentIndex = 0
	viewer.resolveCurrent(ActionAccept)
	viewer.currentIndex = 1
	viewer.resolveCurrent(ActionDeny)
	viewer.currentIndex = 2
	viewer.resolveCurrent(ActionSkip)

	viewer.currentIndex = 0
	viewer.resolveCurrent(ActionDeny)

	if len(viewer.decisions) != 3 {
		t.Fatalf("expected decisions for 3 unique files, got %d", len(viewer.decisions))
	}

	file0Decision := viewer.decisions[0]
	if file0Decision.Path != "file-0.md" {
		t.Fatalf("expected file 0 path file-0.md, got %q", file0Decision.Path)
	}
	if file0Decision.Action != ActionDeny {
		t.Fatalf("expected revisited file 0 action %q, got %q", ActionDeny, file0Decision.Action)
	}

	resolutions := viewer.orderedResolutions()
	if len(resolutions) != 3 {
		t.Fatalf("expected 3 ordered resolutions for unique files, got %d", len(resolutions))
	}
	if resolutions[0].Path != "file-0.md" || resolutions[0].Action != ActionDeny {
		t.Fatalf("expected one overwritten file 0 deny resolution, got %#v", resolutions[0])
	}
}

func TestDiffViewerHunkNavigationKeysMoveBetweenHunks(t *testing.T) {
	t.Parallel()

	viewer := newThreeHunkViewer(t)

	pressKey(viewer, "]")
	pressKey(viewer, "]")

	if viewer.hunkIndex != 2 {
		t.Fatalf("expected hunk index 2 after two next-hunk keys, got %d", viewer.hunkIndex)
	}
	if got, want := viewer.leftVP.YOffset(), viewer.hunkStarts[2]; got != want {
		t.Fatalf("expected left viewport offset %d, got %d", want, got)
	}
	if got, want := viewer.rightVP.YOffset(), viewer.hunkStarts[2]; got != want {
		t.Fatalf("expected right viewport offset %d, got %d", want, got)
	}

	pressKey(viewer, "[")

	if viewer.hunkIndex != 1 {
		t.Fatalf("expected hunk index 1 after previous-hunk key, got %d", viewer.hunkIndex)
	}
}

func TestDiffViewerHunkNavigationKeysStayWithinBoundaries(t *testing.T) {
	t.Parallel()

	viewer := newThreeHunkViewer(t)

	pressKey(viewer, "[")
	if viewer.hunkIndex != 0 {
		t.Fatalf("expected previous hunk on first hunk to stay at 0, got %d", viewer.hunkIndex)
	}

	viewer.hunkIndex = len(viewer.hunkStarts) - 1
	viewer.scrollToCurrentHunk()
	pressKey(viewer, "]")

	if got, want := viewer.hunkIndex, len(viewer.hunkStarts)-1; got != want {
		t.Fatalf("expected next hunk on last hunk to stay at %d, got %d", want, got)
	}
}

func TestDiffViewerTransitionsToSummaryAfterAllFilesDecided(t *testing.T) {
	t.Parallel()

	viewer := NewDiffViewer([]ConflictView{
		{FilePath: "file-0.md"},
		{FilePath: "file-1.md"},
	})

	pressKey(viewer, "a")
	pressKey(viewer, "d")

	if viewer.state != summary {
		t.Fatalf("expected summary state after all files are decided, got %q", viewer.state)
	}

	view := viewer.View().Content
	for _, want := range []string{"file-0.md", string(ActionAccept), "file-1.md", string(ActionDeny)} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected summary view to contain %q, got:\n%s", want, view)
		}
	}
}

func TestDiffViewerConfirmFromSummarySetsConfirmedState(t *testing.T) {
	t.Parallel()

	viewer := newSummaryViewer()
	cmd := pressKey(viewer, "y")

	if viewer.state != confirmed {
		t.Fatalf("expected confirmed state, got %q", viewer.state)
	}
	if cmd == nil {
		t.Fatal("expected confirm from summary to quit")
	}
}

func TestDiffViewerCancelFromSummarySetsCancelledState(t *testing.T) {
	t.Parallel()

	viewer := newSummaryViewer()
	cmd := pressKey(viewer, "q")

	if viewer.state != cancelled {
		t.Fatalf("expected cancelled state, got %q", viewer.state)
	}
	if cmd == nil {
		t.Fatal("expected cancel from summary to quit")
	}
}

func TestDiffViewerRunReturnsConfirmedResponseForConfirmedState(t *testing.T) {
	t.Parallel()

	viewer := newSummaryViewer()
	pressKey(viewer, "y")

	resp, err := viewer.Run()
	if err != nil {
		t.Fatalf("expected confirmed response, got error: %v", err)
	}
	if resp.Status != ReviewStatusConfirmed {
		t.Fatalf("expected confirmed status, got %q", resp.Status)
	}
	if len(resp.Resolutions) != len(viewer.conflicts) {
		t.Fatalf("expected confirmed response with %d resolutions, got %d", len(viewer.conflicts), len(resp.Resolutions))
	}
}

func TestDiffViewerRunReturnsCancelledResponseForCancelledState(t *testing.T) {
	t.Parallel()

	viewer := newSummaryViewer()
	pressKey(viewer, "q")

	resp, err := viewer.Run()
	if err != nil {
		t.Fatalf("expected cancelled response, got error: %v", err)
	}
	if resp.Status != ReviewStatusCancelled {
		t.Fatalf("expected cancelled status, got %q", resp.Status)
	}
	if len(resp.Resolutions) != 0 {
		t.Fatalf("expected cancelled response with no resolutions, got %d", len(resp.Resolutions))
	}
}

func TestDiffViewerPreviousFromSummaryReopensLastFile(t *testing.T) {
	t.Parallel()

	viewer := newSummaryViewer()
	pressKey(viewer, "p")

	if viewer.state != reviewingFile {
		t.Fatalf("expected reviewingFile state, got %q", viewer.state)
	}
	if got, want := viewer.currentIndex, len(viewer.conflicts)-1; got != want {
		t.Fatalf("expected current index %d, got %d", want, got)
	}
}

func TestDiffViewerRenderSummaryUsesRemainingAsTotalMinusResolved(t *testing.T) {
	t.Parallel()

	viewer := NewDiffViewer([]ConflictView{
		{FilePath: "file-0.md"},
		{FilePath: "file-1.md"},
		{FilePath: "file-2.md"},
	})

	viewer.currentIndex = 2
	viewer.resolveCurrent(ActionAccept)

	if got, want := viewer.renderSummary(), "Conflicts: 3 | Resolved: 1 | Remaining: 2"; got != want {
		t.Fatalf("expected out-of-order summary %q, got %q", want, got)
	}
}

func TestDiffViewerDoesNotIntroduceFileWrites(t *testing.T) {
	t.Parallel()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test filename")
	}
	viewerSource, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "viewer.go"))
	if err != nil {
		t.Fatalf("read viewer source: %v", err)
	}

	for _, forbidden := range []string{"os.WriteFile", "os.Create", "os.OpenFile", "ioutil.WriteFile"} {
		if strings.Contains(string(viewerSource), forbidden) {
			t.Fatalf("viewer.go must not introduce file writes; found %q", forbidden)
		}
	}
}

func newThreeHunkViewer(t *testing.T) *DiffViewer {
	t.Helper()

	viewer := NewDiffViewer([]ConflictView{{
		FilePath:  "file.md",
		DiffLines: threeHunkDiffLines(),
	}})
	viewer.width = 80
	viewer.height = 13
	viewer.syncViewports()

	wantHunkStarts := []int{1, 5, 7}
	if !slices.Equal(viewer.hunkStarts, wantHunkStarts) {
		t.Fatalf("test setup expected hunk starts %v, got %v", wantHunkStarts, viewer.hunkStarts)
	}

	return viewer
}

func newSummaryViewer() *DiffViewer {
	viewer := NewDiffViewer([]ConflictView{
		{FilePath: "file-0.md"},
		{FilePath: "file-1.md"},
	})
	pressKey(viewer, "a")
	pressKey(viewer, "s")
	return viewer
}

func threeHunkDiffLines() []DiffLine {
	return []DiffLine{
		{Type: DiffLineContext, Content: "context before"},
		{Type: DiffLineRemoved, Content: "old hunk one"},
		{Type: DiffLineAdded, Content: "new hunk one"},
		{Type: DiffLineContext, Content: "context one"},
		{Type: DiffLineContext, Content: "context two"},
		{Type: DiffLineAdded, Content: "new hunk two"},
		{Type: DiffLineContext, Content: "context three"},
		{Type: DiffLineRemoved, Content: "old hunk three"},
		{Type: DiffLineAdded, Content: "new hunk three"},
		{Type: DiffLineContext, Content: "trailing context one"},
		{Type: DiffLineContext, Content: "trailing context two"},
		{Type: DiffLineContext, Content: "trailing context three"},
		{Type: DiffLineContext, Content: "trailing context four"},
		{Type: DiffLineContext, Content: "trailing context five"},
		{Type: DiffLineContext, Content: "trailing context six"},
		{Type: DiffLineContext, Content: "trailing context seven"},
		{Type: DiffLineContext, Content: "trailing context eight"},
	}
}

func pressKey(viewer *DiffViewer, key string) tea.Cmd {
	_, cmd := viewer.Update(tea.KeyPressMsg(tea.Key{Text: key, Code: []rune(key)[0]}))
	return cmd
}
