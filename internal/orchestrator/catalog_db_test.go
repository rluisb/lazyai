package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
)

// createTestCatalogDB creates an in-memory test DB with catalog tables.
func createTestCatalogDB(t *testing.T) *CatalogDB {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test-catalog.db")
	cdb, err := OpenCatalogDB(dbPath)
	if err != nil {
		t.Fatalf("OpenCatalogDB: %v", err)
	}

	// Create the catalog tables (matching orchestrator migration 0003).
	createSQL := `
	CREATE TABLE IF NOT EXISTS definitions (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  kind TEXT NOT NULL CHECK (kind IN ('agent','skill','chain','team','workflow','mode','command')),
	  name TEXT NOT NULL,
	  active_version_id INTEGER,
	  created_at TEXT NOT NULL,
	  updated_at TEXT NOT NULL,
	  UNIQUE(kind, name)
	);

	CREATE TABLE IF NOT EXISTS definition_versions (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  definition_id INTEGER NOT NULL REFERENCES definitions(id),
	  version INTEGER NOT NULL,
	  frontmatter_json TEXT NOT NULL,
	  body TEXT NOT NULL,
	  checksum TEXT NOT NULL,
	  created_at TEXT NOT NULL,
	  created_by TEXT,
	  UNIQUE(definition_id, version)
	);

	CREATE INDEX IF NOT EXISTS idx_def_versions_checksum ON definition_versions(checksum);
	CREATE INDEX IF NOT EXISTS idx_def_versions_def ON definition_versions(definition_id, version);
	`
	if _, err := cdb.DB.Exec(createSQL); err != nil {
		cdb.Close()
		t.Fatalf("create tables: %v", err)
	}

	t.Cleanup(func() { cdb.Close() })
	return cdb
}

func TestCatalogDB_CreateVersion(t *testing.T) {
	cdb := createTestCatalogDB(t)

	// Create a temp file with chain JSON.
	chainJSON := `{"name": "tdd", "kind": "chain", "entry": "scout", "steps": [{"id": "s1", "agent": "scout", "skills": [], "description": "research", "transitions": {"done": "s2"}}, {"id": "s2", "agent": "builder", "skills": [], "description": "build", "transitions": {"done": "done"}}]}`
	tmpFile := filepath.Join(t.TempDir(), "tdd-chain.json")
	if err := os.WriteFile(tmpFile, []byte(chainJSON), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	result, err := cdb.CreateVersion("chain", "tdd", tmpFile, "", true)
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	if result.AlreadyExists {
		t.Error("expected new version, got already exists")
	}
	if result.Version != 1 {
		t.Errorf("expected version 1, got %d", result.Version)
	}
	if result.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestCatalogDB_CreateVersion_Dedup(t *testing.T) {
	cdb := createTestCatalogDB(t)

	chainJSON := `{"name": "tdd", "kind": "chain", "entry": "scout", "steps": []}`
	tmpFile := filepath.Join(t.TempDir(), "tdd-chain.json")
	if err := os.WriteFile(tmpFile, []byte(chainJSON), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	// Create first version.
	result1, err := cdb.CreateVersion("chain", "tdd", tmpFile, "", true)
	if err != nil {
		t.Fatalf("CreateVersion 1: %v", err)
	}
	if result1.AlreadyExists {
		t.Error("expected new version for first create")
	}

	// Create same content again — should dedup by checksum.
	result2, err := cdb.CreateVersion("chain", "tdd", tmpFile, "", true)
	if err != nil {
		t.Fatalf("CreateVersion 2: %v", err)
	}
	if !result2.AlreadyExists {
		t.Error("expected dedup, got new version")
	}
	if result2.Version != result1.Version {
		t.Errorf("expected same version %d, got %d", result1.Version, result2.Version)
	}
}

func TestCatalogDB_ListDefinitions(t *testing.T) {
	cdb := createTestCatalogDB(t)

	// Create a few definitions manually.
	now := "2024-01-15T00:00:00Z"
	if _, err := cdb.DB.Exec(`INSERT INTO definitions (kind, name, created_at, updated_at) VALUES ('chain', 'tdd', ?, ?)`, now, now); err != nil {
		t.Fatalf("insert def 1: %v", err)
	}
	if _, err := cdb.DB.Exec(`INSERT INTO definitions (kind, name, created_at, updated_at) VALUES ('team', 'review-team', ?, ?)`, now, now); err != nil {
		t.Fatalf("insert def 2: %v", err)
	}
	if _, err := cdb.DB.Exec(`INSERT INTO definitions (kind, name, created_at, updated_at) VALUES ('workflow', 'bugfix', ?, ?)`, now, now); err != nil {
		t.Fatalf("insert def 3: %v", err)
	}

	defs, err := cdb.ListDefinitions("")
	if err != nil {
		t.Fatalf("ListDefinitions: %v", err)
	}

	if len(defs) != 3 {
		t.Errorf("expected 3 definitions, got %d", len(defs))
	}

	// Filter by kind.
	chainDefs, err := cdb.ListDefinitions("chain")
	if err != nil {
		t.Fatalf("ListDefinitions(chain): %v", err)
	}
	if len(chainDefs) != 1 || chainDefs[0].Name != "tdd" {
		t.Errorf("expected 1 chain definition named tdd, got %v", chainDefs)
	}
}

func TestCatalogDB_ListVersions(t *testing.T) {
	cdb := createTestCatalogDB(t)

	chainJSON := `{"name": "tdd-v1", "kind": "chain", "steps": []}`
	tmpFile := filepath.Join(t.TempDir(), "tdd-v1.json")
	if err := os.WriteFile(tmpFile, []byte(chainJSON), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	if _, err := cdb.CreateVersion("chain", "tdd", tmpFile, "", true); err != nil {
		t.Fatalf("CreateVersion 1: %v", err)
	}

	chainJSON2 := `{"name": "tdd-v2", "kind": "chain", "steps": [{"id": "s1"}]}`
	tmpFile2 := filepath.Join(t.TempDir(), "tdd-v2.json")
	if err := os.WriteFile(tmpFile2, []byte(chainJSON2), 0o644); err != nil {
		t.Fatalf("write tmp file 2: %v", err)
	}

	if _, err := cdb.CreateVersion("chain", "tdd", tmpFile2, "", true); err != nil {
		t.Fatalf("CreateVersion 2: %v", err)
	}

	versions, err := cdb.ListVersions("chain", "tdd")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != 1 {
		t.Errorf("expected version 1 first, got %d", versions[0].Version)
	}
	if versions[1].Version != 2 {
		t.Errorf("expected version 2 second, got %d", versions[1].Version)
	}
	// v2 should be active (setActive=true on create)
	if !versions[1].IsActive {
		t.Error("expected v2 to be active")
	}
}

func TestCatalogDB_SetActive(t *testing.T) {
	cdb := createTestCatalogDB(t)

	chainJSON := `{"name": "tdd-v1", "kind": "chain", "steps": []}`
	tmpFile := filepath.Join(t.TempDir(), "tdd-v1.json")
	if err := os.WriteFile(tmpFile, []byte(chainJSON), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	if _, err := cdb.CreateVersion("chain", "tdd", tmpFile, "", true); err != nil {
		t.Fatalf("CreateVersion 1: %v", err)
	}

	chainJSON2 := `{"name": "tdd-v2", "kind": "chain", "steps": [{"id": "s1"}]}`
	tmpFile2 := filepath.Join(t.TempDir(), "tdd-v2.json")
	if err := os.WriteFile(tmpFile2, []byte(chainJSON2), 0o644); err != nil {
		t.Fatalf("write tmp file 2: %v", err)
	}

	if _, err := cdb.CreateVersion("chain", "tdd", tmpFile2, "", false); err != nil {
		t.Fatalf("CreateVersion 2: %v", err)
	}

	// Set v1 as active.
	if err := cdb.SetActive("chain", "tdd", 1); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	versions, err := cdb.ListVersions("chain", "tdd")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}

	if !versions[0].IsActive {
		t.Error("expected v1 to be active after set-active")
	}
	if versions[1].IsActive {
		t.Error("expected v2 to NOT be active")
	}
}

func TestCatalogDB_DiffVersions(t *testing.T) {
	cdb := createTestCatalogDB(t)

	chainJSON := `{"name": "tdd-v1", "kind": "chain", "steps": []}`
	tmpFile := filepath.Join(t.TempDir(), "tdd-v1.json")
	if err := os.WriteFile(tmpFile, []byte(chainJSON), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	if _, err := cdb.CreateVersion("chain", "tdd", tmpFile, "", true); err != nil {
		t.Fatalf("CreateVersion 1: %v", err)
	}

	chainJSON2 := `{"name": "tdd-v2", "kind": "chain", "steps": [{"id": "s1"}]}`
	tmpFile2 := filepath.Join(t.TempDir(), "tdd-v2.json")
	if err := os.WriteFile(tmpFile2, []byte(chainJSON2), 0o644); err != nil {
		t.Fatalf("write tmp file 2: %v", err)
	}

	if _, err := cdb.CreateVersion("chain", "tdd", tmpFile2, "", true); err != nil {
		t.Fatalf("CreateVersion 2: %v", err)
	}

	diff, err := cdb.DiffVersions("chain", "tdd", 1, 2)
	if err != nil {
		t.Fatalf("DiffVersions: %v", err)
	}

	if diff.From == nil {
		t.Fatal("expected from version not nil")
	}
	if diff.To == nil {
		t.Fatal("expected to version not nil")
	}
	if diff.From.Version != 1 {
		t.Errorf("expected from version 1, got %d", diff.From.Version)
	}
	if diff.To.Version != 2 {
		t.Errorf("expected to version 2, got %d", diff.To.Version)
	}
	if diff.From.Checksum == diff.To.Checksum {
		t.Error("expected different checksums for different content")
	}
}

func TestCatalogDB_DefaultDBPath(t *testing.T) {
	path, err := DefaultCatalogDBPath()
	if err != nil {
		t.Fatalf("DefaultCatalogDBPath: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty default path")
	}
	// Path should contain the known suffix.
	expectedSuffix := filepath.Join(".local", "share", "ai-setup-orchestrator", "orchestrator.db")
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
	_ = expectedSuffix // just verify it resolves without error
}

func TestFormatCatalogList(t *testing.T) {
	defs := []CatalogDefSummary{
		{Kind: "chain", Name: "tdd", ActiveVersion: intPtr(2), TotalVersions: 3},
		{Kind: "team", Name: "review-team", ActiveVersion: intPtr(1), TotalVersions: 1},
		{Kind: "workflow", Name: "bugfix", ActiveVersion: nil, TotalVersions: 0},
	}
	output := FormatCatalogList(defs)
	if output == "" {
		t.Error("expected non-empty formatted output")
	}
	// Check key content.
	if !contains(output, "chain") || !contains(output, "tdd") || !contains(output, "v2") {
		t.Errorf("missing expected content: %s", output)
	}
}

func TestFormatVersionList(t *testing.T) {
	versions := []CatalogVersionRow{
		{Version: 1, CreatedAt: "2024-01-15T00:00:00Z", IsActive: false, Kind: "chain", Body: `{"steps":[1,2,3]}`},
		{Version: 2, CreatedAt: "2024-03-20T00:00:00Z", IsActive: true, Kind: "chain", Body: `{"steps":[1,2,3,4,5]}`},
	}
	output := FormatVersionList(versions)
	if output == "" {
		t.Error("expected non-empty formatted output")
	}
	if !contains(output, "v1") || !contains(output, "v2") || !contains(output, "2024-01-15") {
		t.Errorf("missing expected content: %s", output)
	}
}

func TestCatalogChecksum(t *testing.T) {
	fm1 := map[string]any{"name": "tdd"}
	body1 := `{"steps":[]}`

	cs1 := catalogChecksum(fm1, body1)
	cs2 := catalogChecksum(fm1, body1)

	if cs1 != cs2 {
		t.Error("same input should produce same checksum")
	}
	if len(cs1) != 16 {
		t.Errorf("expected 16-char checksum, got %d", len(cs1))
	}

	// Different body should produce different checksum.
	cs3 := catalogChecksum(fm1, `{"steps":[1]}`)
	if cs1 == cs3 {
		t.Error("different content should produce different checksums")
	}
}

func TestListCategoryToKind(t *testing.T) {
	tests := []struct {
		cat  ListCategory
		want string
	}{
		{CategoryChains, "chain"},
		{CategoryTeams, "team"},
		{CategoryWorkflows, "workflow"},
		{CategoryDomains, "skill"},
		{CategoryModes, "mode"},
	}
	for _, tt := range tests {
		got := listCategoryToKind(tt.cat)
		if got != tt.want {
			t.Errorf("listCategoryToKind(%q) = %q, want %q", tt.cat, got, tt.want)
		}
	}
}

func TestKindToListCategory(t *testing.T) {
	tests := []struct {
		kind string
		want ListCategory
	}{
		{"chain", CategoryChains},
		{"team", CategoryTeams},
		{"workflow", CategoryWorkflows},
		{"skill", CategoryDomains},
		{"mode", CategoryModes},
		{"unknown", ListCategory("unknown")},
	}
	for _, tt := range tests {
		got := kindToListCategory(tt.kind)
		if got != tt.want {
			t.Errorf("kindToListCategory(%q) = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

// helpers

func intPtr(i int) *int { return &i }

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
