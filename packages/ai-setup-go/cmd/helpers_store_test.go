package cmd

import (
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/internal/scaffold"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

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
		TargetDir:      dir,
		Tools:          []types.ToolId{types.ToolIdOpenCode},
		CLITools:       []string{"gh"},
		EnableServers:  []string{"filesystem", "orchestrator"},
		ProjectName:    "demo-app",
		PlanningDir:    "specs",
		SetupScope:     types.SetupScopeProject,
		Features:       &features,
		GitConventions: &gitConventions,
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
	if got, want := storeData.Config.EnableServers, []string{"filesystem", "orchestrator"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("EnableServers = %#v, want %#v", got, want)
	}
	if len(storeData.Files) != 1 || storeData.Files[0].Owner != types.FileOwnerLibrary {
		t.Fatalf("Files = %#v, want one library-owned tracked file", storeData.Files)
	}
}
