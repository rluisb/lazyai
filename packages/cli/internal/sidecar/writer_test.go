package sidecar

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWriteProjectSidecar_RejectsMissingProjectRoot(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	err := WriteProjectSidecar(filepath.Join(projectRoot, "missing"), &SidecarConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project root not accessible")
}

func TestWriteProjectSidecar_RejectsFileProjectRoot(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	fileRoot := filepath.Join(projectRoot, "not-a-dir")
	require.NoError(t, os.WriteFile(fileRoot, []byte("x"), 0o644))

	err := WriteProjectSidecar(fileRoot, &SidecarConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project root is not a directory")
}

func TestWriteProjectSidecar_WritesAtomicallyAndBacksUpExistingFile(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

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

	writeProjectSidecar(t, projectRoot, oldConfig)
	require.NoError(t, WriteProjectSidecar(projectRoot, newConfig))

	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	newData, err := os.ReadFile(path)
	require.NoError(t, err)
	backupData, err := os.ReadFile(path + ".bak")
	require.NoError(t, err)

	var newFile ProjectSidecarConfig
	var backup ProjectSidecarConfig
	require.NoError(t, yaml.Unmarshal(newData, &newFile))
	require.NoError(t, yaml.Unmarshal(backupData, &backup))

	assert.Equal(t, newConfig, newFile.Sidecar)
	assert.Equal(t, oldConfig, backup.Sidecar)
}

func TestWriteGlobalSidecar_WritesAtomicallyAndBacksUpExistingFile(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	oldConfig := &SidecarConfig{
		Path:     filepath.Join(projectRoot, "old-global"),
		DocsDir:  "old-docs",
		SpecsDir: "old-specs",
		PlansDir: "old-plans",
	}
	newConfig := &SidecarConfig{
		Path:     filepath.Join(projectRoot, "new-global"),
		DocsDir:  "new-docs",
		SpecsDir: "new-specs",
		PlansDir: "new-plans",
	}

	writeGlobalSidecar(t, oldConfig)
	require.NoError(t, WriteGlobalSidecar(newConfig))

	globalDir, err := getGlobalConfigDir()
	require.NoError(t, err)
	path := filepath.Join(globalDir, "sidecar.yaml")

	newData, err := os.ReadFile(path)
	require.NoError(t, err)
	backupData, err := os.ReadFile(path + ".bak")
	require.NoError(t, err)

	var newFile GlobalSidecarConfig
	var backup GlobalSidecarConfig
	require.NoError(t, yaml.Unmarshal(newData, &newFile))
	require.NoError(t, yaml.Unmarshal(backupData, &backup))

	assert.Equal(t, newConfig, newFile.Sidecar)
	assert.Equal(t, oldConfig, backup.Sidecar)
}

func TestSaveWorkspaceConfig_WritesSingleSlotBackup(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	firstConfig := &WorkspaceConfig{
		Active: "first",
		Workspaces: []WorkspaceEntry{
			{Name: "workspace-a", Path: "/tmp/a"},
		},
	}
	secondConfig := &WorkspaceConfig{
		Active: "second",
		Workspaces: []WorkspaceEntry{
			{Name: "workspace-b", Path: "/tmp/b"},
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
			cfg.Workspaces = append(cfg.Workspaces, WorkspaceEntry{Name: "a", Path: "/tmp/a"})
			return nil
		})
	}()
	go func() {
		defer wg.Done()
		errCh <- UpdateWorkspaceConfig(func(cfg *WorkspaceConfig) error {
			cfg.Workspaces = append(cfg.Workspaces, WorkspaceEntry{Name: "b", Path: "/tmp/b"})
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
