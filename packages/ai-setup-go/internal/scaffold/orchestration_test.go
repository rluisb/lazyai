package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestScaffoldOrchestration_AddsExtensionOrchestrationContent(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createMinimalLibraryFS()
	fileRecords := []types.TrackedFile{}

	localExtDir := filepath.Join(targetDir, ".ai", "extensions", "team-pack")
	if err := os.MkdirAll(filepath.Join(localExtDir, "skills"), 0o755); err != nil {
		t.Fatalf("mkdir local extension skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(localExtDir, "orchestration", "chains"), 0o755); err != nil {
		t.Fatalf("mkdir local extension chains: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localExtDir, "skills", "custom.md"), []byte("# custom"), 0o644); err != nil {
		t.Fatalf("write local extension skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localExtDir, "orchestration", "chains", "release.json"), []byte(`{"name":"release"}`), 0o644); err != nil {
		t.Fatalf("write local extension chain: %v", err)
	}

	sharedExtDir := filepath.Join(targetDir, "shared-ext")
	if err := os.MkdirAll(filepath.Join(sharedExtDir, "skills"), 0o755); err != nil {
		t.Fatalf("mkdir shared extension skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sharedExtDir, "orchestration", "workflows"), 0o755); err != nil {
		t.Fatalf("mkdir shared extension workflows: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedExtDir, "skills", "deploy.md"), []byte("# deploy"), 0o644); err != nil {
		t.Fatalf("write shared extension skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedExtDir, "orchestration", "workflows", "deploy.json"), []byte(`{"name":"deploy"}`), 0o644); err != nil {
		t.Fatalf("write shared extension workflow: %v", err)
	}

	config := "[extensions.shared]\npath = \"./shared-ext\"\n"
	if err := os.WriteFile(filepath.Join(targetDir, ".ai-setup.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write extension config: %v", err)
	}

	if err := ScaffoldOrchestration(targetDir, libFS, &fileRecords, types.ConflictStrategySkip, map[string]types.ConflictStrategy{}); err != nil {
		t.Fatalf("ScaffoldOrchestration failed: %v", err)
	}

	if !filesExist(
		filepath.Join(targetDir, ".ai", "orchestration", "chains", "feature.json"),
		filepath.Join(targetDir, ".ai", "orchestration", "chains", "release.json"),
		filepath.Join(targetDir, ".ai", "orchestration", "workflows", "deploy.json"),
	) {
		t.Fatal("expected built-in and extension orchestration files to be scaffolded")
	}

	if !hasTrackedFile(fileRecords, ".ai/orchestration/chains/release.json") {
		t.Fatal("expected extension chain to be tracked")
	}
	if !hasTrackedFile(fileRecords, ".ai/orchestration/workflows/deploy.json") {
		t.Fatal("expected extension workflow to be tracked")
	}
}

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func hasTrackedFile(records []types.TrackedFile, path string) bool {
	for _, record := range records {
		if record.Path == path {
			return true
		}
	}
	return false
}
