package migration

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestExecuteToCanonical_WritesParsedMigrationContent(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(sourceDir, "AGENTS.md"), "# Demo Project\n\nShared instructions.")
	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "agents", "reviewer.md"), "# Reviewer\n\nReview carefully.")
	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "commands", "plan.md"), "# Plan\n\nMake a plan.")
	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "templates", "handoff.md"), "# Handoff\n\nSummarize work.")

	ctx := &MigrationContext{
		SourcePath: sourceDir,
		TargetPath: targetDir,
		Options: MigrationOptions{
			Interactive: false,
		},
	}

	detections, err := DetectSetup(sourceDir)
	if err != nil {
		t.Fatalf("DetectSetup error: %v", err)
	}

	parsedSetups, err := ParseDetectedSetups(ctx, detections)
	if err != nil {
		t.Fatalf("ParseDetectedSetups error: %v", err)
	}

	plan, err := BuildCanonicalPlan(ctx, detections, parsedSetups)
	if err != nil {
		t.Fatalf("BuildCanonicalPlan error: %v", err)
	}

	result, err := ExecuteToCanonical(ctx, plan, parsedSetups)
	if err != nil {
		t.Fatalf("ExecuteToCanonical error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got errors: %v", result.Errors)
	}

	assertFileContent(t, filepath.Join(targetDir, ".ai", "agents", "reviewer.md"), "# Reviewer\n\nReview carefully.")
	assertFileContent(t, filepath.Join(targetDir, ".ai", "skills", "plan.md"), "# Plan\n\nMake a plan.")
	assertFileContent(t, filepath.Join(targetDir, ".ai", "prompts", "handoff.md"), "# Handoff\n\nSummarize work.")
	if files.FileExists(filepath.Join(targetDir, ".ai", "constitution", "opencode.md")) {
		t.Fatal("root AGENTS.md should not be adapted into canonical constitution")
	}

	if result.Stats.FilesCreated != 3 {
		t.Fatalf("expected 3 created files, got %d", result.Stats.FilesCreated)
	}
	if result.Stats.FilesBackedUp != 4 {
		t.Fatalf("expected 4 backed up files, got %d", result.Stats.FilesBackedUp)
	}
	if result.BackupPath == "" {
		t.Fatal("expected backup path")
	}
	assertFileContent(t, filepath.Join(result.BackupPath, ".ai-setup-backup", "AGENTS.md"), "# Demo Project\n\nShared instructions.")

	manifestBytes, err := files.ReadFile(filepath.Join(targetDir, ".ai-setup.json"))
	if err != nil {
		t.Fatalf("manifest read error: %v", err)
	}

	var manifest types.StoreData
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest unmarshal error: %v", err)
	}
	if len(manifest.Files) != 3 {
		t.Fatalf("expected 3 tracked files, got %d", len(manifest.Files))
	}
	if len(manifest.Config.Tools) != 1 || manifest.Config.Tools[0] != types.ToolIdOpenCode {
		t.Fatalf("expected opencode tool in manifest, got %+v", manifest.Config.Tools)
	}
}

func TestExecuteToCanonical_MigratesOpenCodeAgentsWithoutAbsorbingRootContext(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(sourceDir, "AGENTS.md"), "# Demo Project\n\nRoot context should remain root context.")
	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "agents", "builder.md"), "# Builder\n\nBuild things.")

	ctx := &MigrationContext{
		SourcePath: sourceDir,
		TargetPath: targetDir,
		Options: MigrationOptions{
			Interactive: false,
		},
	}
	detections, err := DetectSetup(sourceDir)
	if err != nil {
		t.Fatalf("DetectSetup error: %v", err)
	}
	parsedSetups, err := ParseDetectedSetups(ctx, detections)
	if err != nil {
		t.Fatalf("ParseDetectedSetups error: %v", err)
	}
	plan, err := BuildCanonicalPlan(ctx, detections, parsedSetups)
	if err != nil {
		t.Fatalf("BuildCanonicalPlan error: %v", err)
	}

	result, err := ExecuteToCanonical(ctx, plan, parsedSetups)
	if err != nil {
		t.Fatalf("ExecuteToCanonical error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got errors: %v", result.Errors)
	}

	assertFileContent(t, filepath.Join(targetDir, ".ai", "agents", "builder.md"), "# Builder\n\nBuild things.")
	if files.FileExists(filepath.Join(targetDir, ".ai", "constitution", "opencode.md")) {
		t.Fatal("root AGENTS.md should not create canonical constitution output")
	}
	if result.Stats.FilesCreated != 1 {
		t.Fatalf("expected 1 created file, got %d", result.Stats.FilesCreated)
	}
}

func TestExecuteToCanonical_PreviewDoesNotWriteFiles(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "agents", "reviewer.md"), "# Reviewer\n\nReview carefully.")

	ctx := &MigrationContext{
		SourcePath: sourceDir,
		TargetPath: targetDir,
		Options: MigrationOptions{
			Preview:     true,
			Interactive: false,
		},
	}

	detections, err := DetectSetup(sourceDir)
	if err != nil {
		t.Fatalf("DetectSetup error: %v", err)
	}
	parsedSetups, err := ParseDetectedSetups(ctx, detections)
	if err != nil {
		t.Fatalf("ParseDetectedSetups error: %v", err)
	}
	plan, err := BuildCanonicalPlan(ctx, detections, parsedSetups)
	if err != nil {
		t.Fatalf("BuildCanonicalPlan error: %v", err)
	}

	result, err := ExecuteToCanonical(ctx, plan, parsedSetups)
	if err != nil {
		t.Fatalf("ExecuteToCanonical error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got errors: %v", result.Errors)
	}
	if result.Stats.FilesCreated != 1 {
		t.Fatalf("expected preview create count 1, got %d", result.Stats.FilesCreated)
	}
	if files.FileExists(filepath.Join(targetDir, ".ai", "agents", "reviewer.md")) {
		t.Fatal("preview should not write canonical file")
	}
	if files.FileExists(filepath.Join(targetDir, ".ai-setup.json")) {
		t.Fatal("preview should not write manifest")
	}
	if files.DirExists(filepath.Join(targetDir, ".ai-setup-backup")) {
		t.Fatal("preview should not create backup directory")
	}
	if len(result.ExecutedActions) != 1 || result.ExecutedActions[0].TargetPath != filepath.Join(".ai", "agents", "reviewer.md") {
		t.Fatalf("unexpected preview actions: %+v", result.ExecutedActions)
	}
}

func TestParseDetectedSetupsParsesAgentButSkipsReservedContextDoc(t *testing.T) {
	sourceDir := t.TempDir()
	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "agents", "builder.md"), "# Builder\n\nBuild things.")
	mustWriteTestFile(t, filepath.Join(sourceDir, ".opencode", "agents", "AGENTS.md"), "# Nested Context\n\nDo not parse.")

	ctx := &MigrationContext{SourcePath: sourceDir, TargetPath: t.TempDir()}
	detections, err := DetectSetup(sourceDir)
	if err != nil {
		t.Fatalf("DetectSetup error: %v", err)
	}
	parsedSetups, err := ParseDetectedSetups(ctx, detections)
	if err != nil {
		t.Fatalf("ParseDetectedSetups error: %v", err)
	}

	parsed := parsedSetups[0]
	if len(parsed.Agents) != 1 || parsed.Agents[0].ID != "builder" {
		t.Fatalf("agents = %+v, want only builder", parsed.Agents)
	}
	if len(parsed.Files) != 1 || parsed.Files[0].Path != filepath.Join(".opencode", "agents", "builder.md") {
		t.Fatalf("files = %+v, want only builder file", parsed.Files)
	}
}

func mustWriteTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := files.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()
	data, err := files.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("unexpected content for %s\nwant: %q\n got: %q", path, want, string(data))
	}
}
