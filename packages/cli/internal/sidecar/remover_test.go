package sidecar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRemoveWorkspaceSidecar_RemovesActiveSidecarAndPreservesOtherEntries(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	initial := &WorkspaceConfig{
		Active: "workspace-1",
		Workspaces: []WorkspaceEntry{
			{
				Name: "workspace-1",
				Path: projectRoot,
				Sidecar: &SidecarConfig{
					Path: "workspace-1-sidecar",
				},
			},
			{
				Name: "workspace-2",
				Path: filepath.Join(projectRoot, "other"),
			},
		},
	}
	writeWorkspaceConfig(t, initial)

	require.NoError(t, RemoveWorkspaceSidecar())

	cfg, err := LoadWorkspaceConfig()
	require.NoError(t, err)

	assert.Equal(t, 2, len(cfg.Workspaces))

	var active, other WorkspaceEntry
	for _, workspace := range cfg.Workspaces {
		if workspace.Name == "workspace-1" {
			active = workspace
		}
		if workspace.Name == "workspace-2" {
			other = workspace
		}
	}

	assert.Nil(t, active.Sidecar)
	assert.Nil(t, other.Sidecar)
}

func TestRemoveWorkspaceSidecar_NoActiveWorkspaceIsNoop(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	initial := &WorkspaceConfig{
		Active: "",
		Workspaces: []WorkspaceEntry{
			{
				Name:    "workspace-1",
				Path:    projectRoot,
				Sidecar: &SidecarConfig{Path: "workspace-1-sidecar"},
			},
		},
	}
	writeWorkspaceConfig(t, initial)

	require.NoError(t, RemoveWorkspaceSidecar())

	cfg, err := LoadWorkspaceConfig()
	require.NoError(t, err)

	assert.Equal(t, initial, cfg)
}

func TestRemoveWorkspaceSidecar_NoSidecarIsNoop(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	initial := &WorkspaceConfig{
		Active: "workspace-1",
		Workspaces: []WorkspaceEntry{
			{
				Name: "workspace-1",
				Path: projectRoot,
			},
		},
	}
	writeWorkspaceConfig(t, initial)

	require.NoError(t, RemoveWorkspaceSidecar())

	cfg, err := LoadWorkspaceConfig()
	require.NoError(t, err)

	assert.Equal(t, initial, cfg)
}

func TestRemoveWorkspaceSidecar_ActiveWorkspaceMissingReturnsError(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	initial := &WorkspaceConfig{
		Active: "missing",
		Workspaces: []WorkspaceEntry{
			{
				Name:    "workspace-1",
				Path:    projectRoot,
				Sidecar: &SidecarConfig{Path: "workspace-1-sidecar"},
			},
		},
	}
	writeWorkspaceConfig(t, initial)

	err := RemoveWorkspaceSidecar()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active workspace \"missing\" not found")
}

func TestRemoveProjectSidecar_RemovesFile(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeProjectSidecar(t, projectRoot, &SidecarConfig{Path: projectRoot})

	require.NoError(t, RemoveProjectSidecar(projectRoot))

	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveProjectSidecar_MissingFileIsNoop(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	require.NoError(t, RemoveProjectSidecar(projectRoot))
}

func TestRemoveWorkspaceSidecar_WritesBackupThroughSaveWorkspaceConfig(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	initial := &WorkspaceConfig{
		Active: "workspace-1",
		Workspaces: []WorkspaceEntry{
			{
				Name:    "workspace-1",
				Path:    "some-path",
				Sidecar: &SidecarConfig{Path: "workspace-1-sidecar"},
			},
		},
	}
	writeWorkspaceConfig(t, initial)

	require.NoError(t, RemoveWorkspaceSidecar())

	path, err := getWorkspacesConfigPath()
	require.NoError(t, err)

	backupPath := path + ".bak"
	backupData, err := os.ReadFile(backupPath)
	require.NoError(t, err)

	var backupConfig WorkspaceConfig
	require.NoError(t, yaml.Unmarshal(backupData, &backupConfig))

	assert.Equal(t, initial, &backupConfig)
}
