package agentmemory

import (
	"database/sql"
	"testing"
)

func seedSearchData(t *testing.T, db *sql.DB) {
	t.Helper()
	now := testNow()
	memoryStore := NewMemoryStore(db)
	artifactStore := NewArtifactStore(db)
	for _, memory := range []Memory{
		{Namespace: "ns", Content: "deterministic continuity keeps resumable agent work", SourceTaskID: "task-1", Tags: "continuity", Importance: ImportanceHigh, CreatedAt: now},
		{Namespace: "ns", Content: "prefix matching supports lexical workflows", SourceTaskID: "task-1", Tags: "search", Importance: ImportanceNormal, CreatedAt: now},
		{Namespace: "other", Content: "deterministic continuity in other namespace", SourceTaskID: "task-2", Tags: "continuity", Importance: ImportanceNormal, CreatedAt: now},
	} {
		if _, err := memoryStore.SaveMemory(t.Context(), memory); err != nil {
			t.Fatalf("SaveMemory() error = %v", err)
		}
	}
	if err := artifactStore.RecordArtifact(t.Context(), Artifact{TaskID: "task-1", Namespace: "ns", Path: "notes/research.md", ContentPreview: "artifact preview contains deterministic continuity details", Tags: "research", CreatedAt: now}); err != nil {
		t.Fatalf("RecordArtifact() error = %v", err)
	}
}

func TestSearchMemoriesExactPrefixPhraseAndSpecialCharacters(t *testing.T) {
	db := testDB(t)
	seedSearchData(t, db)

	exact, err := SearchMemories(t.Context(), db, "ns", "resumable", 10)
	if err != nil {
		t.Fatalf("SearchMemories(exact) error = %v", err)
	}
	if len(exact) != 1 || exact[0].Namespace != "ns" || exact[0].Snippet == "" {
		t.Fatalf("exact results = %+v", exact)
	}

	prefix, err := SearchMemories(t.Context(), db, "ns", "resum*", 10)
	if err != nil {
		t.Fatalf("SearchMemories(prefix) error = %v", err)
	}
	if len(prefix) != 1 {
		t.Fatalf("prefix results len = %d, want 1: %+v", len(prefix), prefix)
	}

	phrase, err := SearchMemories(t.Context(), db, "ns", `"deterministic continuity"`, 10)
	if err != nil {
		t.Fatalf("SearchMemories(phrase) error = %v", err)
	}
	if len(phrase) != 1 {
		t.Fatalf("phrase results len = %d, want 1: %+v", len(phrase), phrase)
	}

	if _, err := SearchMemories(t.Context(), db, "ns", "", 10); err == nil {
		t.Fatal("SearchMemories(empty) error = nil, want error")
	}

	none, err := SearchMemories(t.Context(), db, "ns", "nonexistent", 10)
	if err != nil {
		t.Fatalf("SearchMemories(no results) error = %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("no-results len = %d, want 0", len(none))
	}

	if _, err := SearchMemories(t.Context(), db, "ns", `sk-test* "unterminated`, 10); err != nil {
		t.Fatalf("SearchMemories(special chars) error = %v", err)
	}
}

func TestSearchArtifacts(t *testing.T) {
	db := testDB(t)
	seedSearchData(t, db)

	results, err := SearchArtifacts(t.Context(), db, "ns", "artifact", 10)
	if err != nil {
		t.Fatalf("SearchArtifacts() error = %v", err)
	}
	if len(results) != 1 || results[0].Path != "notes/research.md" || results[0].Snippet == "" {
		t.Fatalf("SearchArtifacts() = %+v", results)
	}
}
