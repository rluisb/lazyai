package sidecar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoadWorkspaceSidecar_MissingFile(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// No workspaces.yaml written.
	cfg, err := LoadWorkspaceSidecar()
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadWorkspaceSidecar_NoSidecarBlock(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name: "my-project",
				Path: projectRoot,
			},
		},
		Active: "my-project",
	})

	cfg, err := LoadWorkspaceSidecar()
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadWorkspaceSidecar_WithSidecar(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name: "my-project",
				Path: projectRoot,
				Sidecar: &SidecarConfig{
					Path: "/placeholder/kb",
				},
			},
		},
		Active: "my-project",
	})

	cfg, err := LoadWorkspaceSidecar()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "/placeholder/kb", cfg.Path)
}

func TestLoadWorkspaceSidecar_NoActiveWorkspace(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{},
		Active:     "",
	})

	cfg, err := LoadWorkspaceSidecar()
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadProjectSidecar_MissingFile(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg, err := LoadProjectSidecar(projectRoot)
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadProjectSidecar_Valid(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     "../kb",
		SpecsDir: "specs",
		DocsDir:  "docs",
		PlansDir: "plans",
	})

	cfg, err := LoadProjectSidecar(projectRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "../kb", cfg.Path)
	assert.Equal(t, "specs", cfg.SpecsDir)
	assert.Equal(t, "docs", cfg.DocsDir)
	assert.Equal(t, "plans", cfg.PlansDir)
}

func TestLoadProjectSidecar_MalformedYAML(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: valid: yaml: ["), 0o644))

	_, err := LoadProjectSidecar(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing project sidecar")
}

func TestLoadGlobalSidecar_MissingFile(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg, err := LoadGlobalSidecar()
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadGlobalSidecar_Valid(t *testing.T) {
	_, _, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     filepath.Join(globalDir, "kb"),
		SpecsDir: "specs",
		DocsDir:  "docs",
		PlansDir: "plans",
	})

	cfg, err := LoadGlobalSidecar()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, filepath.Join(globalDir, "kb"), cfg.Path)
}

func TestLoadGlobalSidecar_MalformedYAML(t *testing.T) {
	_, _, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	path := filepath.Join(globalDir, "sidecar.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: valid: yaml: ["), 0o644))

	_, err := LoadGlobalSidecar()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing global sidecar")
}

func TestLoadWorkspaceConfig_MalformedYAML(t *testing.T) {
	_, _, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	path := filepath.Join(globalDir, "workspaces.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: valid: yaml: ["), 0o644))

	_, err := LoadWorkspaceConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing workspace config")
}

func TestLoadWorkspaceConfig_BackwardCompat(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Write an old-style workspace config without sidecar field.
	path := filepath.Join(globalDir, "workspaces.yaml")
	oldConfig := map[string]any{
		"workspaces": []map[string]string{
			{"name": "my-project", "path": projectRoot},
		},
		"active": "my-project",
	}
	data, err := yaml.Marshal(oldConfig)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	cfg, err := LoadWorkspaceConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Len(t, cfg.Workspaces, 1)
	assert.Equal(t, "my-project", cfg.Workspaces[0].Name)
	assert.Nil(t, cfg.Workspaces[0].Sidecar)
}
