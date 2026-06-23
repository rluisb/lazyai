package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/manifest"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/tui/wizard"
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

func setTestHome(t *testing.T, home string) {
	t.Helper()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}
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

	// Read with timeout to avoid deadlock when background goroutines
	// (e.g. SQLite connection opener) hold references to stdout.
	var stdout, stderr string
	select {
	case stdout = <-stdoutDone:
	case <-time.After(2 * time.Second):
		stdout = "(timed out)"
	}
	select {
	case stderr = <-stderrDone:
	case <-time.After(2 * time.Second):
		stderr = "(timed out)"
	}
	return stdout, stderr
}

func newUpdateCommand(force, nonInteractive, dryRun bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("no-interactive", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	_ = cmd.Flags().Set("force", boolString(force))
	_ = cmd.Flags().Set("no-interactive", boolString(nonInteractive))
	_ = cmd.Flags().Set("dry-run", boolString(dryRun))
	return cmd
}

func newCompileCommand(dir, tool string, dryRun bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("dir", "", "")
	cmd.Flags().String("tool", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("local-secrets", false, "")
	cmd.Flags().Bool("validate-contracts", true, "")
	cmd.Flags().Bool("strict-contracts", false, "")
	_ = cmd.Flags().Set("dir", dir)
	if tool != "" {
		_ = cmd.Flags().Set("tool", tool)
	}
	_ = cmd.Flags().Set("dry-run", boolString(dryRun))
	return cmd
}

func newImportCommand(tool string, nonInteractive, preview bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("tool", "", "")
	cmd.Flags().Bool("no-interactive", false, "")
	cmd.Flags().Bool("preview", false, "")
	cmd.Flags().String("strategy", "smart", "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("skip-backup", false, "")
	if tool != "" {
		_ = cmd.Flags().Set("tool", tool)
	}
	_ = cmd.Flags().Set("no-interactive", boolString(nonInteractive))
	_ = cmd.Flags().Set("preview", boolString(preview))
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

func fileHashForTest(t *testing.T, path string) string {
	t.Helper()
	hash, err := files.FileHash(path)
	if err != nil {
		t.Fatalf("FileHash %q: %v", path, err)
	}
	return hash
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
