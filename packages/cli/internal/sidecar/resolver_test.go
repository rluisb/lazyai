package sidecar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupTestEnv creates temporary directories and workspace config for testing.
func setupTestEnv(t *testing.T) (homeDir, projectRoot, globalDir string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()
	homeDir = filepath.Join(tmpDir, "home")
	projectRoot = filepath.Join(tmpDir, "project")
	globalDir = filepath.Join(homeDir, ".lazyai")

	require.NoError(t, os.MkdirAll(globalDir, 0o755))
	require.NoError(t, os.MkdirAll(projectRoot, 0o755))

	// Override home directory. Go's os.UserHomeDir uses HOME on Unix and
	// USERPROFILE on Windows, so set both for cross-platform tests.
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", homeDir)
	os.Setenv("USERPROFILE", homeDir)

	cleanup = func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}

	return homeDir, projectRoot, globalDir, cleanup
}

func ensureDirs(t *testing.T, paths ...string) {
	for _, p := range paths {
		require.NoError(t, os.MkdirAll(p, 0o755))
	}
}

// writeWorkspaceConfig writes a workspace config to the global config dir.
func writeWorkspaceConfig(t *testing.T, cfg *WorkspaceConfig) {
	t.Helper()
	dir, err := getGlobalConfigDir()
	require.NoError(t, err)
	path := filepath.Join(dir, "workspaces.yaml")
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

// writeGlobalSidecar writes a global sidecar config.
func writeGlobalSidecar(t *testing.T, cfg *SidecarConfig) {
	t.Helper()
	dir, err := getGlobalConfigDir()
	require.NoError(t, err)
	path := filepath.Join(dir, "sidecar.yaml")
	data, err := yaml.Marshal(GlobalSidecarConfig{Sidecar: cfg})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

// writeProjectSidecar writes a project sidecar config.
func writeProjectSidecar(t *testing.T, projectRoot string, cfg *SidecarConfig) {
	t.Helper()
	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	data, err := yaml.Marshal(ProjectSidecarConfig{Sidecar: cfg})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

func TestResolve_NoSidecarAnyLevel(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// No workspace config, no sidecars.
	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "default", result.ConfigLevel)
	assert.Equal(t, filepath.Join(projectRoot, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectRoot, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectRoot, "plans"), result.PlansDir)
}

func TestResolve_GlobalOnly(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	sidecarPath := filepath.Join(globalDir, "kb")
	require.NoError(t, os.MkdirAll(sidecarPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     sidecarPath,
		SpecsDir: "my-specs",
		DocsDir:  "my-docs",
		PlansDir: "my-plans",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "global", result.ConfigLevel)
	assert.Equal(t, filepath.Join(sidecarPath, "my-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(sidecarPath, "my-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(sidecarPath, "my-plans"), result.PlansDir)
}

func TestResolve_ProjectOverridesGlobal(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	globalPath := filepath.Join(globalDir, "global-kb")
	projectPath := filepath.Join(projectRoot, "project-kb")
	require.NoError(t, os.MkdirAll(globalPath, 0o755))
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     globalPath,
		SpecsDir: "global-specs",
		DocsDir:  "global-docs",
		PlansDir: "global-plans",
	})

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectPath,
		SpecsDir: "project-specs",
		DocsDir:  "project-docs",
		PlansDir: "project-plans",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "project", result.ConfigLevel)
	assert.Equal(t, filepath.Join(projectPath, "project-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectPath, "project-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectPath, "project-plans"), result.PlansDir)
}

func TestResolve_WorkspaceOverridesProject(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	workspacePath := filepath.Join(globalDir, "workspace-kb")
	projectPath := filepath.Join(projectRoot, "project-kb")
	require.NoError(t, os.MkdirAll(workspacePath, 0o755))
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name: "my-project",
				Path: projectRoot,
				Sidecar: &SidecarConfig{
					Path:     workspacePath,
					SpecsDir: "workspace-specs",
					DocsDir:  "workspace-docs",
					PlansDir: "workspace-plans",
				},
			},
		},
		Active: "my-project",
	})

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectPath,
		SpecsDir: "project-specs",
		DocsDir:  "project-docs",
		PlansDir: "project-plans",
	})

	result, err := Resolve(ScopeWorkspace, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "workspace", result.ConfigLevel)
	assert.Equal(t, filepath.Join(workspacePath, "workspace-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(workspacePath, "workspace-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(workspacePath, "workspace-plans"), result.PlansDir)
}

func TestResolve_RelativePathResolution(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Global sidecar with relative path.
	kbPath := filepath.Join(globalDir, "kb")
	require.NoError(t, os.MkdirAll(kbPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     "kb", // relative to globalDir
		SpecsDir: "specs",
		DocsDir:  "docs",
		PlansDir: "plans",
	})

	result, err := Resolve(ScopeGlobal, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "global", result.ConfigLevel)
	assert.Equal(t, filepath.Join(kbPath, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(kbPath, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(kbPath, "plans"), result.PlansDir)
}

func TestResolve_AbsolutePathPassthrough(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	absPath := filepath.Join(os.TempDir(), "lazyai-test-abs-kb")
	require.NoError(t, os.MkdirAll(absPath, 0o755))
	defer os.RemoveAll(absPath)

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     absPath,
		SpecsDir: "specs",
		DocsDir:  "docs",
		PlansDir: "plans",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "project", result.ConfigLevel)
	assert.Equal(t, filepath.Join(absPath, "docs"), result.DocsDir)
}

func TestResolve_DefaultDirValues(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	kbPath := filepath.Join(globalDir, "kb")
	require.NoError(t, os.MkdirAll(kbPath, 0o755))

	// Sidecar with empty *_dir fields.
	writeGlobalSidecar(t, &SidecarConfig{
		Path: kbPath,
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(kbPath, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(kbPath, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(kbPath, "plans"), result.PlansDir)
}

func TestResolve_AbsoluteDirPaths(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	absDocs := filepath.Join(os.TempDir(), "lazyai-test-abs-docs")
	absSpecs := filepath.Join(os.TempDir(), "lazyai-test-abs-specs")
	require.NoError(t, os.MkdirAll(absDocs, 0o755))
	require.NoError(t, os.MkdirAll(absSpecs, 0o755))
	defer os.RemoveAll(absDocs)
	defer os.RemoveAll(absSpecs)

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectRoot,
		DocsDir:  absDocs,
		SpecsDir: absSpecs,
		PlansDir: "plans",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, absDocs, result.DocsDir)
	assert.Equal(t, absSpecs, result.SpecsDir)
	assert.Equal(t, filepath.Join(projectRoot, "plans"), result.PlansDir)
}

func TestResolve_WorkscopeFallback(t *testing.T) {
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

	result, err := Resolve(ScopeWorkspace, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "default", result.ConfigLevel)
	assert.Equal(t, filepath.Join(projectRoot, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectRoot, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectRoot, "plans"), result.PlansDir)
}

func TestResolve_GlobalScopeDefault(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	result, err := Resolve(ScopeGlobal, projectRoot)
	require.NoError(t, err)

	assert.Equal(t, "default", result.ConfigLevel)
	assert.Equal(t, filepath.Join(globalDir, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(globalDir, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(globalDir, "plans"), result.PlansDir)
}

func TestResolve_GlobalScopeIgnoresProjectAndWorkspace(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	globalSidecarPath := filepath.Join(globalDir, "global-sidecar")
	projectSidecarPath := filepath.Join(projectRoot, "project-sidecar")
	workspaceSidecarPath := filepath.Join(globalDir, "workspace-sidecar")
	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name:    "my-project",
				Path:    projectRoot,
				Sidecar: &SidecarConfig{Path: workspaceSidecarPath},
			},
		},
		Active: "my-project",
	})
	ensureDirs(t,
		filepath.Join(globalSidecarPath, "docs"),
		filepath.Join(globalSidecarPath, "specs"),
		filepath.Join(globalSidecarPath, "plans"),
		filepath.Join(projectSidecarPath, "docs"),
		filepath.Join(projectSidecarPath, "specs"),
		filepath.Join(projectSidecarPath, "plans"),
		filepath.Join(workspaceSidecarPath, "docs"),
		filepath.Join(workspaceSidecarPath, "specs"),
		filepath.Join(workspaceSidecarPath, "plans"),
	)

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     globalSidecarPath,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectSidecarPath,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})

	result, err := Resolve(ScopeGlobal, projectRoot)
	require.NoError(t, err)
	assert.Equal(t, "global", result.ConfigLevel)
	assert.Equal(t, filepath.Join(globalSidecarPath, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(globalSidecarPath, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(globalSidecarPath, "plans"), result.PlansDir)
}

func TestResolve_ProjectScopeIgnoresWorkspaceSidecar(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	globalSidecarPath := filepath.Join(globalDir, "global-sidecar")
	projectSidecarPath := filepath.Join(projectRoot, "project-sidecar")
	workspaceSidecarPath := filepath.Join(globalDir, "workspace-sidecar")
	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name:    "my-project",
				Path:    projectRoot,
				Sidecar: &SidecarConfig{Path: workspaceSidecarPath},
			},
		},
		Active: "my-project",
	})
	ensureDirs(t,
		filepath.Join(globalSidecarPath, "docs"),
		filepath.Join(globalSidecarPath, "specs"),
		filepath.Join(globalSidecarPath, "plans"),
		filepath.Join(projectSidecarPath, "docs"),
		filepath.Join(projectSidecarPath, "specs"),
		filepath.Join(projectSidecarPath, "plans"),
		filepath.Join(workspaceSidecarPath, "docs"),
		filepath.Join(workspaceSidecarPath, "specs"),
		filepath.Join(workspaceSidecarPath, "plans"),
	)

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     globalSidecarPath,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectSidecarPath,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)
	assert.Equal(t, "project", result.ConfigLevel)
	assert.Equal(t, filepath.Join(projectSidecarPath, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectSidecarPath, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectSidecarPath, "plans"), result.PlansDir)
}

func TestResolve_NoActiveWorkspace(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{},
		Active:     "",
	})

	_, err := Resolve(ScopeWorkspace, projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active workspace")
}

func TestResolve_RejectsEmptySidecarPath(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path: "",
	})

	_, err := Resolve(ScopeProject, projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sidecar path is required but empty")
}

func TestResolve_GlobalRelativePathUsesGlobalConfigDirAtProjectScope(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     "kb",
		SpecsDir: "specs",
		DocsDir:  "docs",
		PlansDir: "plans",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	globalSidecarDir := filepath.Join(globalDir, "kb")
	assert.Equal(t, "global", result.ConfigLevel)
	assert.Equal(t, filepath.Join(globalSidecarDir, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(globalSidecarDir, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(globalSidecarDir, "plans"), result.PlansDir)
	assert.NotEqual(t, filepath.Join(projectRoot, "kb", "docs"), result.DocsDir)
}

func TestResolve_ProjectRelativePathUsesProjectRootAtWorkspaceScope(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
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

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path: "project-kb",
	})

	result, err := Resolve(ScopeWorkspace, projectRoot)
	require.NoError(t, err)

	projectSidecarDir := filepath.Join(projectRoot, "project-kb")
	assert.Equal(t, "project", result.ConfigLevel)
	assert.Equal(t, filepath.Join(projectSidecarDir, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectSidecarDir, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectSidecarDir, "plans"), result.PlansDir)
	assert.NotEqual(t, filepath.Join(globalDir, "project-kb", "docs"), result.DocsDir)
}

func TestResolve_WorkspaceRelativePathUsesGlobalConfigDirAtWorkspaceScope(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name: "my-project",
				Path: projectRoot,
				Sidecar: &SidecarConfig{
					Path:     "workspace-kb",
					SpecsDir: "workspace-specs",
					DocsDir:  "workspace-docs",
					PlansDir: "workspace-plans",
				},
			},
		},
		Active: "my-project",
	})

	result, err := Resolve(ScopeWorkspace, projectRoot)
	require.NoError(t, err)

	workspaceSidecarDir := filepath.Join(globalDir, "workspace-kb")
	assert.Equal(t, "workspace", result.ConfigLevel)
	assert.Equal(t, filepath.Join(workspaceSidecarDir, "workspace-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(workspaceSidecarDir, "workspace-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(workspaceSidecarDir, "workspace-plans"), result.PlansDir)
	assert.NotEqual(t, filepath.Join(projectRoot, "workspace-kb", "workspace-docs"), result.DocsDir)
}

func TestResolve_HigherPrioritySidecarReplacesWholeTuple(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     filepath.Join(string(os.PathSeparator), "global-kb"),
		SpecsDir: "global-specs",
		DocsDir:  "global-docs",
		PlansDir: "global-plans",
	})

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     "project-kb",
		DocsDir:  "custom-docs",
		SpecsDir: "",
		PlansDir: "",
	})

	result, err := Resolve(ScopeProject, projectRoot)
	require.NoError(t, err)

	projectSidecarDir := filepath.Join(projectRoot, "project-kb")
	assert.Equal(t, "project", result.ConfigLevel)
	assert.Equal(t, filepath.Join(projectSidecarDir, "custom-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectSidecarDir, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectSidecarDir, "plans"), result.PlansDir)
	assert.NotEqual(t, filepath.Join(string(os.PathSeparator), "global-kb", "custom-docs"), result.DocsDir)
}
