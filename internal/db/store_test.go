package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open :memory: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func setupStore(t *testing.T) *Store {
	t.Helper()
	db := openTestDB(t)
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	store := NewStore(db)
	if err := store.Initialize("test-version"); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return store
}

func TestOpen_InMemory(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		t.Fatal("db is nil")
	}
	if db.Path() != ":memory:" {
		t.Errorf("Path() = %q, want :memory:", db.Path())
	}
}

func TestRunMigrations_CreatesAllTables(t *testing.T) {
	db := openTestDB(t)
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	expectedTables := []string{"meta", "config", "selections", "tracked_files", "operations", "sync", "feature_flags"}
	for _, table := range expectedTables {
		var count int
		err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("checking table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("table %q not found (count=%d)", table, count)
		}
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db := openTestDB(t)
	if err := RunMigrations(db); err != nil {
		t.Fatalf("first RunMigrations: %v", err)
	}
	if err := RunMigrations(db); err != nil {
		t.Fatalf("second RunMigrations: %v", err)
	}
}

func TestWriteAndReadStoreData_RoundTrip(t *testing.T) {
	store := setupStore(t)

	now := time.Now().UTC().Format(time.RFC3339)
	original := &types.StoreData{
		Meta: types.Meta{
			SchemaVersion: 1,
			CLIVersion:    "1.0.0-test",
			InstalledAt:   now,
			LastUpdatedAt: now,
		},
		Config: types.Config{
			SetupScope:  types.SetupScopeProject,
			Tools:       []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode},
			ProjectName: "test-project",
			TargetDir:   "/tmp/test",
		},
		Selections: types.WizardSelections{
			Templates:    []types.TemplateId{types.TemplateIdAdr},
			Rules:        []types.RuleId{types.RuleIdCodeStyle},
			Agents:       []types.AgentId{types.AgentIdBuilder},
			Skills:       []types.SkillId{types.SkillIdPlan},
			Prompts:      []types.PromptId{types.PromptIdCompact},
			Commands:         []types.CommandId{types.CommandIdRpi},
			ChatModes:        []types.ChatModeId{types.ChatModeIdArchitect},
			OpenCodeCommands: []types.OpenCodeCommandId{types.OpenCodeCommandIdReview},
			OpenCodeModes:    []types.OpenCodeModeId{types.OpenCodeModeIdPlan},
			OpenCodePlugins:  []string{"@opencode/git-tools"},
			Infra:            []types.InfraId{types.InfraIdPreCommit},
			Constitution: []string{"constraints"},
			Features:     &types.FeatureFlags{QualityGates: true},
			GitConventions: &types.GitConventions{
				BranchPattern: "{type}/{ticket}-{description}",
				CommitPattern: "{type}({scope}): {description}",
				Types:         []string{"feat", "fix"},
				RequireTicket: false,
				TicketPattern: "[A-Z]+-[0-9]+",
			},
		},
		Files: []types.TrackedFile{
			{
				Path:        "AGENTS.md",
				Hash:        "abc123",
				Source:      "library",
				Owner:       types.FileOwnerLibrary,
				Status:      types.FileStatusInstalled,
				InstalledAt: now,
			},
		},
		Sync: types.Sync{
			LastSyncAt: now,
			Dirty:      false,
		},
		Operations: []types.Operation{
			{
				ID:            "op-001",
				Type:          "install",
				Timestamp:     now,
				FilesAffected: []string{"AGENTS.md"},
				Result:        types.OperationResultSuccess,
			},
		},
	}

	if err := store.WriteStoreData(original); err != nil {
		t.Fatalf("WriteStoreData: %v", err)
	}

	got, err := store.ReadStoreData()
	if err != nil {
		t.Fatalf("ReadStoreData: %v", err)
	}

	if got.Meta.CLIVersion != original.Meta.CLIVersion {
		t.Errorf("CLIVersion = %q, want %q", got.Meta.CLIVersion, original.Meta.CLIVersion)
	}
	if got.Config.SetupScope != original.Config.SetupScope {
		t.Errorf("SetupScope = %q, want %q", got.Config.SetupScope, original.Config.SetupScope)
	}
	if got.Config.ProjectName != original.Config.ProjectName {
		t.Errorf("ProjectName = %q, want %q", got.Config.ProjectName, original.Config.ProjectName)
	}
	if len(got.Config.Tools) != 2 {
		t.Errorf("Tools length = %d, want 2", len(got.Config.Tools))
	}
	if len(got.Files) != 1 {
		t.Errorf("Files length = %d, want 1", len(got.Files))
	}
	if got.Files[0].Path != "AGENTS.md" {
		t.Errorf("Files[0].Path = %q, want AGENTS.md", got.Files[0].Path)
	}
	if len(got.Operations) != 1 {
		t.Errorf("Operations length = %d, want 1", len(got.Operations))
	}
	if got.Operations[0].ID != "op-001" {
		t.Errorf("Operations[0].ID = %q, want op-001", got.Operations[0].ID)
	}
	if got.Sync.Dirty != false {
		t.Error("Sync.Dirty = true, want false")
	}
	if got.Selections.Features == nil || !got.Selections.Features.QualityGates {
		t.Error("Features.QualityGates not preserved")
	}
	if len(got.Selections.Commands) != 1 || got.Selections.Commands[0] != types.CommandIdRpi {
		t.Errorf("Commands not preserved: got %v", got.Selections.Commands)
	}
	if len(got.Selections.ChatModes) != 1 || got.Selections.ChatModes[0] != types.ChatModeIdArchitect {
		t.Errorf("ChatModes not preserved: got %v", got.Selections.ChatModes)
	}
	if len(got.Selections.OpenCodeCommands) != 1 || got.Selections.OpenCodeCommands[0] != types.OpenCodeCommandIdReview {
		t.Errorf("OpenCodeCommands not preserved: got %v", got.Selections.OpenCodeCommands)
	}
	if len(got.Selections.OpenCodeModes) != 1 || got.Selections.OpenCodeModes[0] != types.OpenCodeModeIdPlan {
		t.Errorf("OpenCodeModes not preserved: got %v", got.Selections.OpenCodeModes)
	}
	if len(got.Selections.OpenCodePlugins) != 1 || got.Selections.OpenCodePlugins[0] != "@opencode/git-tools" {
		t.Errorf("OpenCodePlugins not preserved: got %v", got.Selections.OpenCodePlugins)
	}
}

func TestUpsertTrackedFile_AndReadTrackedFiles(t *testing.T) {
	store := setupStore(t)
	now := time.Now().UTC().Format(time.RFC3339)

	file := types.TrackedFile{
		Path:          "specs/rules/access.md",
		Hash:          "deadbeef",
		Source:        "library",
		Owner:         types.FileOwnerLibrary,
		Status:        types.FileStatusInstalled,
		InstalledAt:   now,
		LastCheckedAt: now,
	}

	if err := store.UpsertTrackedFile(file); err != nil {
		t.Fatalf("UpsertTrackedFile: %v", err)
	}

	files, err := store.ReadTrackedFiles()
	if err != nil {
		t.Fatalf("ReadTrackedFiles: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("files length = %d, want 1", len(files))
	}
	if files[0].Path != file.Path {
		t.Errorf("Path = %q, want %q", files[0].Path, file.Path)
	}
	if files[0].Hash != file.Hash {
		t.Errorf("Hash = %q, want %q", files[0].Hash, file.Hash)
	}
	if files[0].Owner != file.Owner {
		t.Errorf("Owner = %q, want %q", files[0].Owner, file.Owner)
	}
}

func TestUpsertTrackedFile_Update(t *testing.T) {
	store := setupStore(t)

	file1 := types.TrackedFile{
		Path:   "test.md",
		Hash:   "hash1",
		Source: "library",
		Owner:  types.FileOwnerLibrary,
		Status: types.FileStatusInstalled,
	}
	file2 := types.TrackedFile{
		Path:   "test.md",
		Hash:   "hash2",
		Source: "user",
		Owner:  types.FileOwnerUser,
		Status: types.FileStatusModified,
	}

	if err := store.UpsertTrackedFile(file1); err != nil {
		t.Fatalf("UpsertTrackedFile first: %v", err)
	}
	if err := store.UpsertTrackedFile(file2); err != nil {
		t.Fatalf("UpsertTrackedFile second: %v", err)
	}

	files, err := store.ReadTrackedFiles()
	if err != nil {
		t.Fatalf("ReadTrackedFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("files length = %d, want 1 (should update not duplicate)", len(files))
	}
	if files[0].Hash != "hash2" {
		t.Errorf("Hash = %q, want hash2", files[0].Hash)
	}
}

func TestAppendOperation_50Cap(t *testing.T) {
	store := setupStore(t)
	now := time.Now().UTC().Format(time.RFC3339)

	// Insert 55 operations.
	for i := 0; i < 55; i++ {
		op := types.Operation{
			ID:            fmt.Sprintf("op-%03d", i),
			Type:          "install",
			Timestamp:     now,
			FilesAffected: []string{fmt.Sprintf("file-%d.md", i)},
			Result:        types.OperationResultSuccess,
		}
		if err := store.AppendOperation(op); err != nil {
			t.Fatalf("AppendOperation %d: %v", i, err)
		}
	}

	ops, err := store.ReadOperations()
	if err != nil {
		t.Fatalf("ReadOperations: %v", err)
	}

	if len(ops) > 50 {
		t.Errorf("operations count = %d, want <= 50", len(ops))
	}
}

func TestAppendOperation_PreservesLatest(t *testing.T) {
	store := setupStore(t)
	now := time.Now().UTC().Format(time.RFC3339)

	// Insert 55 operations.
	for i := 0; i < 55; i++ {
		op := types.Operation{
			ID:            fmt.Sprintf("op-%03d", i),
			Type:          "install",
			Timestamp:     now,
			FilesAffected: []string{},
			Result:        types.OperationResultSuccess,
		}
		if err := store.AppendOperation(op); err != nil {
			t.Fatalf("AppendOperation %d: %v", i, err)
		}
	}

	ops, err := store.ReadOperations()
	if err != nil {
		t.Fatalf("ReadOperations: %v", err)
	}

	// Check that op-054 (the last inserted) is present.
	found := false
	for _, op := range ops {
		if op.ID == "op-054" {
			found = true
			break
		}
	}
	if !found {
		t.Error("latest operation op-054 not found, should be preserved")
	}
}

func TestFeatureFlags(t *testing.T) {
	store := setupStore(t)

	if err := store.SetFeatureFlag("testFlag", true); err != nil {
		t.Fatalf("SetFeatureFlag: %v", err)
	}

	val, err := store.ReadFeatureFlag("testFlag")
	if err != nil {
		t.Fatalf("ReadFeatureFlag: %v", err)
	}
	if !val {
		t.Error("testFlag = false, want true")
	}

	if err := store.SetFeatureFlag("testFlag", false); err != nil {
		t.Fatalf("SetFeatureFlag false: %v", err)
	}

	val, err = store.ReadFeatureFlag("testFlag")
	if err != nil {
		t.Fatalf("ReadFeatureFlag: %v", err)
	}
	if val {
		t.Error("testFlag = true, want false")
	}
}

func TestImportFromJSON(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, ".ai-setup.json")

	storeData := types.DefaultStoreData()
	storeData.Config.ProjectName = "import-test"
	storeData.Meta.CLIVersion = "0.1.0"

	data, err := json.MarshalIndent(storeData, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	if err := writeFile(jsonPath, data); err != nil {
		t.Fatalf("write test JSON: %v", err)
	}

	got, err := ImportFromJSON(jsonPath)
	if err != nil {
		t.Fatalf("ImportFromJSON: %v", err)
	}
	if got.Config.ProjectName != "import-test" {
		t.Errorf("ProjectName = %q, want import-test", got.Config.ProjectName)
	}
	if got.Meta.CLIVersion != "0.1.0" {
		t.Errorf("CLIVersion = %q, want 0.1.0", got.Meta.CLIVersion)
	}
}

// Helper to write a file with standard permissions.
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
