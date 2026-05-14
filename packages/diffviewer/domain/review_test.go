package domain

import "testing"

func TestRecordResolutionOverwritesAndOrdersByFile(t *testing.T) {
	t.Parallel()

	paths := []string{"file-0.md", "file-1.md", "file-2.md"}
	decisions := map[int]Resolution{}
	decisions = RecordResolution(decisions, len(paths), 0, paths[0], ActionAccept)
	decisions = RecordResolution(decisions, len(paths), 1, paths[1], ActionDeny)
	decisions = RecordResolution(decisions, len(paths), 2, paths[2], ActionSkip)
	decisions = RecordResolution(decisions, len(paths), 0, paths[0], ActionDeny)

	if !AllFilesDecided(len(paths), decisions) {
		t.Fatal("expected all files decided")
	}

	resolutions := OrderedResolutions(paths, decisions)
	if len(resolutions) != len(paths) {
		t.Fatalf("expected %d resolutions, got %d", len(paths), len(resolutions))
	}
	if resolutions[0].Path != paths[0] || resolutions[0].Action != ActionDeny {
		t.Fatalf("expected overwritten first resolution to deny file-0.md, got %#v", resolutions[0])
	}
	if resolutions[1].Path != paths[1] || resolutions[1].Action != ActionDeny {
		t.Fatalf("expected second resolution to deny file-1.md, got %#v", resolutions[1])
	}
	if resolutions[2].Path != paths[2] || resolutions[2].Action != ActionSkip {
		t.Fatalf("expected third resolution to skip file-2.md, got %#v", resolutions[2])
	}
}

func TestReviewResponseForStatusOnlyIncludesConfirmedResolutions(t *testing.T) {
	t.Parallel()

	resolutions := []Resolution{{Path: "AGENTS.md", Action: ActionAccept}}
	confirmed := ReviewResponseForStatus(ReviewStatusConfirmed, resolutions)
	if confirmed.Version != ReviewContractVersion || confirmed.Status != ReviewStatusConfirmed {
		t.Fatalf("unexpected confirmed response metadata: %#v", confirmed)
	}
	if len(confirmed.Resolutions) != 1 {
		t.Fatalf("expected confirmed resolutions to be retained, got %#v", confirmed.Resolutions)
	}

	cancelled := ReviewResponseForStatus(ReviewStatusCancelled, resolutions)
	if cancelled.Version != ReviewContractVersion || cancelled.Status != ReviewStatusCancelled {
		t.Fatalf("unexpected cancelled response metadata: %#v", cancelled)
	}
	if len(cancelled.Resolutions) != 0 {
		t.Fatalf("expected cancelled response to omit resolutions, got %#v", cancelled.Resolutions)
	}
}
