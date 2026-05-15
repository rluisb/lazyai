package agentmemory

import "testing"

func TestArtifactStoreRecordAndList(t *testing.T) {
	db := testDB(t)
	store := NewArtifactStore(db)
	now := testNow()

	artifacts := []Artifact{
		{TaskID: "task-1", Namespace: "ns", Path: "research.md", ContentPreview: "Authorization Bearer abc.def-123", SizeBytes: 42, ContentHash: "hash", MimeType: "text/markdown", Tags: "research", CreatedAt: now},
		{TaskID: "task-1", Namespace: "other", Path: "other.md", CreatedAt: now},
	}
	for _, artifact := range artifacts {
		if err := store.RecordArtifact(t.Context(), artifact); err != nil {
			t.Fatalf("RecordArtifact() error = %v", err)
		}
	}

	listed, err := store.ListArtifacts(t.Context(), "task-1", "ns")
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}
	if len(listed) != 1 || listed[0].Path != "research.md" {
		t.Fatalf("ListArtifacts() = %+v", listed)
	}
	if listed[0].ContentPreview != "Authorization Bearer REDACTED" {
		t.Fatalf("ContentPreview not redacted: %q", listed[0].ContentPreview)
	}
}
