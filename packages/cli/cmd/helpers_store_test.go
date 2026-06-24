package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestOpenStore_FailsFastOnCorruptJSON guards against silent config loss:
// when a legacy .ai-setup.json is present but cannot be imported, openStore
// must return an error rather than proceeding with an empty database.
func TestOpenStore_FailsFastOnCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".ai-setup.json"), []byte("{ this is not valid json "), 0o644); err != nil {
		t.Fatalf("write corrupt json: %v", err)
	}

	database, err := openStore(dir)
	if err == nil {
		database.Close()
		t.Fatal("openStore succeeded on corrupt .ai-setup.json; expected fail-fast error")
	}
}

// TestOpenStore_NoJSONIsNotAnError confirms the absent-JSON happy path is
// preserved: a fresh directory opens cleanly.
func TestOpenStore_NoJSONIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	database, err := openStore(dir)
	if err != nil {
		t.Fatalf("openStore on fresh dir: %v", err)
	}
	database.Close()
}

func TestWriteStoreFromScaffoldResult_PersistsEnableServersAndOwnership(t *testing.T) {
	dir := t.TempDir()
	database, err := openStore(dir)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer database.Close()

	features := types.DefaultFeatureFlags()
	gitConventions := types.DefaultGitConventions()
	ctx := &scaffold.ScaffoldContext{
		TargetDir:        dir,
		WorkspaceRoot:    dir,
		PlanningRepoPath: dir,
		Tools:            []types.ToolId{types.ToolIdOpenCode},
		CLITools:         []string{"gh"},
		EnableServers:    []string{"filesystem", "memory"},
		ProjectName:      "demo-app",
		PlanningDir:      "specs",
		SetupScope:       types.SetupScopeProject,
		Features:         &features,
		GitConventions:   &gitConventions,
		CmdRunner:        func(name string, args ...string) ([]byte, error) { return nil, nil },
	}
	result := &scaffold.ScaffoldResult{Files: []types.TrackedFile{{
		Path:   ".ai/mcp.json",
		Hash:   "abc123",
		Source: "mcp/catalog.json",
		Owner:  types.FileOwnerLibrary,
	}}}

	if err := writeStoreFromScaffoldResult(database, ctx, types.PresetLevelMinimal, result); err != nil {
		t.Fatalf("writeStoreFromScaffoldResult: %v", err)
	}

	store := db.NewStore(database)
	storeData, err := store.ReadStoreData()
	if err != nil {
		t.Fatalf("ReadStoreData: %v", err)
	}
	if got, want := storeData.Config.EnableServers, []string{"filesystem", "memory"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("EnableServers = %#v, want %#v", got, want)
	}
	if len(storeData.Files) != 1 || storeData.Files[0].Owner != types.FileOwnerLibrary {
		t.Fatalf("Files = %#v, want one library-owned tracked file", storeData.Files)
	}
	if storeData.Config.WorkspaceRoot != dir {
		t.Fatalf("WorkspaceRoot = %q, want %q", storeData.Config.WorkspaceRoot, dir)
	}
	if storeData.Config.PlanningRepoPath != dir {
		t.Fatalf("PlanningRepoPath = %q, want %q", storeData.Config.PlanningRepoPath, dir)
	}
}
