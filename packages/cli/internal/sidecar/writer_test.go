package sidecar

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestWriteSidecarAt_RejectsMissingScopeRoot asserts that WriteSidecarAt
// still fails when scopeRoot is genuinely unreachable. WriteSidecarAt does
// MkdirAll(<scopeRoot>/.lazyai) internally, so a merely-missing intermediate
// directory now succeeds via recursive creation (unlike the old
// WriteProjectSidecar, which required scopeRoot to pre-exist) — the only
// remaining failure mode is a scopeRoot whose parent chain cannot be
// created at all, e.g. a permission-denied ancestor.
func TestWriteSidecarAt_RejectsMissingScopeRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits required")
	}
	if os.Geteuid() == 0 {
		t.Skip("root ignores permission bits")
	}

	tmpDir := t.TempDir()
	lockedParent := filepath.Join(tmpDir, "locked")
	require.NoError(t, os.Mkdir(lockedParent, 0o555))
	t.Cleanup(func() { _ = os.Chmod(lockedParent, 0o755) })

	scopeRoot := filepath.Join(lockedParent, "missing", "deeper")
	err := WriteSidecarAt(scopeRoot, &SidecarConfig{})
	require.Error(t, err)
}

// TestWriteSidecarAt_RejectsFileScopeRoot asserts scopeRoot being a regular
// file (not a directory) is rejected — MkdirAll cannot create
// <scopeRoot>/.lazyai underneath a file.
func TestWriteSidecarAt_RejectsFileScopeRoot(t *testing.T) {
	tmpDir := t.TempDir()
	fileRoot := filepath.Join(tmpDir, "not-a-dir")
	require.NoError(t, os.WriteFile(fileRoot, []byte("x"), 0o644))

	err := WriteSidecarAt(fileRoot, &SidecarConfig{})
	require.Error(t, err)
}

// TestWriteSidecarAt_WritesAtomicallyAndBacksUpExistingFile replaces
// TestWriteProjectSidecar_WritesAtomicallyAndBacksUpExistingFile and
// TestWriteGlobalSidecar_WritesAtomicallyAndBacksUpExistingFile: WriteSidecarAt
// is a single function regardless of which scope root it is called with, so
// both scenarios are now subtests of the same table. Asserts the written
// path is exactly <scopeRoot>/.lazyai/sidecar.yaml (standard doc Enforcement,
// spec.md §8).
func TestWriteSidecarAt_WritesAtomicallyAndBacksUpExistingFile(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "project-style root"},
		{name: "global-style root"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scopeRoot := t.TempDir()

			oldConfig := &SidecarConfig{
				Path:     "old-kb",
				DocsDir:  "old-docs",
				SpecsDir: "old-specs",
				PlansDir: "old-plans",
			}
			newConfig := &SidecarConfig{
				Path:     "new-kb",
				DocsDir:  "new-docs",
				SpecsDir: "new-specs",
				PlansDir: "new-plans",
			}

			require.NoError(t, WriteSidecarAt(scopeRoot, oldConfig))
			require.NoError(t, WriteSidecarAt(scopeRoot, newConfig))

			path := filepath.Join(scopeRoot, ".lazyai", "sidecar.yaml")
			newData, err := os.ReadFile(path)
			require.NoError(t, err)
			backupData, err := os.ReadFile(path + ".bak")
			require.NoError(t, err)

			var newFile SidecarFile
			var backup SidecarFile
			require.NoError(t, yaml.Unmarshal(newData, &newFile))
			require.NoError(t, yaml.Unmarshal(backupData, &backup))

			assert.Equal(t, newConfig, newFile.Sidecar)
			assert.Equal(t, oldConfig, backup.Sidecar)
		})
	}
}

// TestWriteSidecarAt_NeverProducesFlatLazyaiSidecarYamlFile is one of the
// standard doc's required Enforcement tests (spec.md §8): asserts the old
// flat <scopeRoot>/.lazyai-sidecar.yaml file is never produced by the new
// unified writer, for both root styles.
func TestWriteSidecarAt_NeverProducesFlatLazyaiSidecarYamlFile(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "project-style root"},
		{name: "global-style root"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scopeRoot := t.TempDir()
			require.NoError(t, WriteSidecarAt(scopeRoot, &SidecarConfig{Path: "kb"}))

			flatPath := filepath.Join(scopeRoot, ".lazyai-sidecar.yaml")
			_, err := os.Stat(flatPath)
			require.True(t, os.IsNotExist(err), "flat sidecar file must never be produced")
		})
	}
}

func TestSaveWorkspaceConfig_WritesSingleSlotBackup(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	firstConfig := &WorkspaceConfig{
		Active: "first",
		Workspaces: []WorkspaceEntry{
			{Name: "workspace-a", Path: "/placeholder/a"},
		},
	}
	secondConfig := &WorkspaceConfig{
		Active: "second",
		Workspaces: []WorkspaceEntry{
			{Name: "workspace-b", Path: "/placeholder/b"},
		},
	}

	require.NoError(t, SaveWorkspaceConfig(firstConfig))
	require.NoError(t, SaveWorkspaceConfig(secondConfig))

	path, err := getWorkspacesConfigPath()
	require.NoError(t, err)

	newData, err := os.ReadFile(path)
	require.NoError(t, err)
	backupData, err := os.ReadFile(path + ".bak")
	require.NoError(t, err)

	var newFile WorkspaceConfig
	var backup WorkspaceConfig
	require.NoError(t, yaml.Unmarshal(newData, &newFile))
	require.NoError(t, yaml.Unmarshal(backupData, &backup))

	assert.Equal(t, secondConfig, &newFile)
	assert.Equal(t, firstConfig, &backup)
}

func TestUpdateWorkspaceConfig_PreservesConcurrentMutations(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	require.NoError(t, SaveWorkspaceConfig(&WorkspaceConfig{}))

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		errCh <- UpdateWorkspaceConfig(func(cfg *WorkspaceConfig) error {
			cfg.Workspaces = append(cfg.Workspaces, WorkspaceEntry{Name: "a", Path: "/placeholder/a"})
			return nil
		})
	}()
	go func() {
		defer wg.Done()
		errCh <- UpdateWorkspaceConfig(func(cfg *WorkspaceConfig) error {
			cfg.Workspaces = append(cfg.Workspaces, WorkspaceEntry{Name: "b", Path: "/placeholder/b"})
			return nil
		})
	}()
	wg.Wait()
	close(errCh)

	errCount := 0
	for err := range errCh {
		if err != nil {
			errCount++
		}
	}
	require.Zero(t, errCount)

	cfg, err := LoadWorkspaceConfig()
	require.NoError(t, err)

	got := map[string]bool{}
	for _, workspace := range cfg.Workspaces {
		got[workspace.Name] = true
	}

	assert.Len(t, cfg.Workspaces, 2)
	assert.True(t, got["a"])
	assert.True(t, got["b"])
}
