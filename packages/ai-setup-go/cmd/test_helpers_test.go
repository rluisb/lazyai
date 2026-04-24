package cmd

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
	"github.com/ricardoborges-teachable/ai-setup/internal/manifest"
	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/scaffold"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
	"github.com/ricardoborges-teachable/ai-setup/tui/wizard"
)

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

func ensureTestLibraryFS(t *testing.T) {
	t.Helper()
	repoRoot := testRepoRoot(t)
	library.ResetLibraryDir()
	library.SetEmbeddedFS(os.DirFS(filepath.Join(repoRoot, "library")))
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
		library.ResetLibraryDir()
	})
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	stdoutDone := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutReader)
		stdoutDone <- buf.String()
	}()
	stderrDone := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderrReader)
		stderrDone <- buf.String()
	}()

	fn()
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr
	return <-stdoutDone, <-stderrDone
}

func newUpdateCommand(force, nonInteractive, dryRun bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	_ = cmd.Flags().Set("force", boolString(force))
	_ = cmd.Flags().Set("non-interactive", boolString(nonInteractive))
	_ = cmd.Flags().Set("dry-run", boolString(dryRun))
	return cmd
}

func newCompileCommand(dir, tool string, dryRun bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("dir", "", "")
	cmd.Flags().String("tool", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	_ = cmd.Flags().Set("dir", dir)
	if tool != "" {
		_ = cmd.Flags().Set("tool", tool)
	}
	_ = cmd.Flags().Set("dry-run", boolString(dryRun))
	return cmd
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func seedStoreData(t *testing.T, dir string, mutate func(*types.StoreData)) *types.StoreData {
	t.Helper()
	database, err := openStore(dir)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	storeData := types.DefaultStoreData()
	storeData.Config.SetupScope = types.SetupScopeProject
	storeData.Config.ProjectName = filepath.Base(dir)
	storeData.Config.TargetDir = dir
	storeData.Config.PlanningDir = "specs"
	storeData.Meta.CLIVersion = Version
	if mutate != nil {
		mutate(&storeData)
	}

	store := db.NewStore(database)
	if err := store.WriteStoreData(&storeData); err != nil {
		t.Fatalf("WriteStoreData: %v", err)
	}
	return &storeData
}

func readSeededStoreData(t *testing.T, dir string) *types.StoreData {
	t.Helper()
	database, err := openStore(dir)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer database.Close()

	store := db.NewStore(database)
	data, err := store.ReadStoreData()
	if err != nil {
		t.Fatalf("ReadStoreData: %v", err)
	}
	return data
}

func writeCanonicalMCPConfig(t *testing.T, dir string) {
	t.Helper()
	content := []byte(`{
  "servers": {
    "test-server": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}`)
	path := filepath.Join(dir, ".ai", "mcp.json")
	if err := files.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile %q: %v", path, err)
	}
}

func runSeedInit(t *testing.T, dir string, tools []types.ToolId, presetLevel types.PresetLevel) {
	t.Helper()
	ensureTestLibraryFS(t)
	withWorkingDir(t, dir)
	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     testRepoRoot(t),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeProject,
		CLITools:    tools,
		CLIPreset:   presetLevel,
	}
	if _, stderr := captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	}); stderr != "" {
		t.Logf("init stderr: %s", stderr)
	}
}

func scaffoldContextFromStoreData(t *testing.T, dir string, storeData *types.StoreData) (*scaffold.ScaffoldContext, types.PresetLevel) {
	t.Helper()
	ensureTestLibraryFS(t)
	presetLevel := preset.DefaultPresetForScope(storeData.Config.SetupScope)
	ctx := &scaffold.ScaffoldContext{
		TargetDir:      dir,
		LibraryDir:     getLibraryDir(),
		LibraryFS:      library.GetLibraryFS(),
		Tools:          storeData.Config.Tools,
		CLITools:       storeData.Config.CLITools,
		EnableServers:  storeData.Config.EnableServers,
		ProjectName:    storeData.Config.ProjectName,
		PlanningDir:    storeData.Config.PlanningDir,
		SetupScope:     storeData.Config.SetupScope,
		Features:       storeData.Selections.Features,
		GitConventions: storeData.Selections.GitConventions,
		Strategy:       types.ConflictStrategyAlign,
		Agents:         storeData.Selections.Agents,
		Skills:         storeData.Selections.Skills,
		Prompts:        storeData.Selections.Prompts,
		Templates:      storeData.Selections.Templates,
		Rules:          storeData.Selections.Rules,
		Infra:          storeData.Selections.Infra,
		SpecsDirs:      preset.SpecsDirsForPreset(presetLevel),
		Housekeeping:   storeData.Config.Housekeeping,
	}
	return ctx, presetLevel
}

func fileHashForTest(t *testing.T, path string) string {
	t.Helper()
	hash, err := files.FileHash(path)
	if err != nil {
		t.Fatalf("FileHash %q: %v", path, err)
	}
	return hash
}

func fileExistsInFS(t *testing.T, root, rel string) bool {
	t.Helper()
	_, err := fs.Stat(os.DirFS(root), rel)
	return err == nil
}

func writeManifestStoreData(t *testing.T, dir string, mutate func(*types.StoreData)) *types.StoreData {
	t.Helper()
	storeData := types.DefaultStoreData()
	storeData.Config.SetupScope = types.SetupScopeProject
	storeData.Config.ProjectName = filepath.Base(dir)
	storeData.Config.TargetDir = dir
	storeData.Config.PlanningDir = "specs"
	storeData.Meta.CLIVersion = Version
	if mutate != nil {
		mutate(&storeData)
	}
	if err := manifest.WriteManifest(dir, &storeData); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	return &storeData
}
