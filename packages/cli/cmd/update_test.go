package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestUpdateNonInteractiveRestoresTrackedSetupState(t *testing.T) {
	dir := t.TempDir()
	runSeedInit(t, dir, []types.ToolId{types.ToolIdOpenCode}, types.PresetLevelMinimal)
	withWorkingDir(t, dir)

	targetPath := filepath.Join(dir, "AGENTS.md")
	if err := os.Remove(targetPath); err != nil {
		t.Fatalf("remove AGENTS.md: %v", err)
	}

	cmd := newUpdateCommand(false, true, false)
	// Run without captureOutput to avoid pipe deadlock with child processes
	if err := runUpdate(cmd, nil); err != nil {
		t.Fatalf("runUpdate: %v", err)
	}

	if !fileExists(targetPath) {
		t.Fatal("expected AGENTS.md to be restored by update")
	}

	storeData := readSeededStoreData(t, dir)
	var record *types.TrackedFile
	for i := range storeData.Files {
		if storeData.Files[i].Path == "AGENTS.md" {
			record = &storeData.Files[i]
			break
		}
	}
	if record == nil {
		t.Fatal("expected AGENTS.md to remain tracked after update")
	}
	if record.Hash != fileHashForTest(t, targetPath) {
		t.Fatalf("tracked hash = %q, want current hash", record.Hash)
	}
}

func TestUpdateNonInteractiveRemovesKnownStrayAgentsArtifacts(t *testing.T) {
	dir := t.TempDir()
	runSeedInit(t, dir, []types.ToolId{types.ToolIdOpenCode}, types.PresetLevelMinimal)
	withWorkingDir(t, dir)

	strayPath := filepath.Join(dir, "specs", "adrs", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(strayPath), 0o755); err != nil {
		t.Fatalf("mkdir stray dir: %v", err)
	}
	if err := os.WriteFile(strayPath, []byte("# legacy stray\n"), 0o644); err != nil {
		t.Fatalf("write stray AGENTS: %v", err)
	}

	storeData := readSeededStoreData(t, dir)
	hash, err := files.FileHash(strayPath)
	if err != nil {
		t.Fatalf("FileHash stray AGENTS: %v", err)
	}
	storeData.Files = append(storeData.Files, types.TrackedFile{
		Path:        "specs/adrs/AGENTS.md",
		Hash:        hash,
		Source:      "library/specs-agents/adrs.md",
		Owner:       types.FileOwnerLibrary,
		Status:      types.FileStatusInstalled,
		InstalledAt: "2026-04-17T00:00:00Z",
	})
	seedStoreData(t, dir, func(data *types.StoreData) {
		*data = *storeData
	})

	cmd := newUpdateCommand(false, true, false)
	if _, _ = captureOutput(t, func() {
		if err := runUpdate(cmd, nil); err != nil {
			t.Fatalf("runUpdate: %v", err)
		}
	}); false {
	}

	if fileExists(strayPath) {
		t.Fatal("expected migrated stray AGENTS.md artifact to be removed by update")
	}

	updated := readSeededStoreData(t, dir)
	for _, tracked := range updated.Files {
		if tracked.Path == "specs/adrs/AGENTS.md" {
			t.Fatal("expected migrated stray AGENTS.md artifact to be removed from store tracking")
		}
	}
}

func TestRemoveLegacyAgentsDeletesUnmodifiedLibraryAgents(t *testing.T) {
	dir := t.TempDir()
	legacyAgentPath := filepath.Join(dir, ".opencode", "agents", "builder.md")
	if err := os.MkdirAll(filepath.Dir(legacyAgentPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(legacyAgentPath, []byte("# legacy builder\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	hash, _ := files.FileHash(legacyAgentPath)
	tracked := []types.TrackedFile{
		{
			Path:  ".opencode/agents/builder.md",
			Hash:  hash,
			Owner: types.FileOwnerLibrary,
		},
	}

	if err := removeLegacyAgents(dir, tracked); err != nil {
		t.Fatalf("removeLegacyAgents: %v", err)
	}

	if fileExists(legacyAgentPath) {
		t.Fatal("expected legacy agent to be removed")
	}
}

func TestRemoveLegacyAgentsPreservesUserEditedAgents(t *testing.T) {
	dir := t.TempDir()
	legacyAgentPath := filepath.Join(dir, ".opencode", "agents", "builder.md")
	if err := os.MkdirAll(filepath.Dir(legacyAgentPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(legacyAgentPath, []byte("# user customized builder\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	tracked := []types.TrackedFile{
		{
			Path:  ".opencode/agents/builder.md",
			Hash:  "oldhash123",
			Owner: types.FileOwnerLibrary,
		},
	}

	if err := removeLegacyAgents(dir, tracked); err != nil {
		t.Fatalf("removeLegacyAgents: %v", err)
	}

	if !fileExists(legacyAgentPath) {
		t.Fatal("expected user-edited legacy agent to be preserved")
	}
}

func TestRemoveLegacyAgentsPreservesUserOwnedAgents(t *testing.T) {
	dir := t.TempDir()
	legacyAgentPath := filepath.Join(dir, ".opencode", "agents", "builder.md")
	if err := os.MkdirAll(filepath.Dir(legacyAgentPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(legacyAgentPath, []byte("# user created builder\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	hash, _ := files.FileHash(legacyAgentPath)
	tracked := []types.TrackedFile{
		{
			Path:  ".opencode/agents/builder.md",
			Hash:  hash,
			Owner: types.FileOwnerUser,
		},
	}

	if err := removeLegacyAgents(dir, tracked); err != nil {
		t.Fatalf("removeLegacyAgents: %v", err)
	}

	if !fileExists(legacyAgentPath) {
		t.Fatal("expected user-owned legacy agent to be preserved")
	}
}
