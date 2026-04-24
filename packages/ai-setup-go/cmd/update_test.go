package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	if _, _ = captureOutput(t, func() {
		if err := runUpdate(cmd, nil); err != nil {
			t.Fatalf("runUpdate: %v", err)
		}
	}); false {
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

// TestUpdateNonInteractiveRemovesKnownStrayAgentsArtifacts was removed when
// `specs/<category>/AGENTS.md` files stopped being "stray" artifacts and
// became legitimate outputs scaffolded by ScaffoldSpecs from
// `library/specs-agents/<category>.md`. The surrounding
// `removeMigratedStrayAgentsArtifacts` cleanup was deleted at the same time.
