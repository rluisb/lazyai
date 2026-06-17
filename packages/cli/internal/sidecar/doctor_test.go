package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDoctor_NoSidecarsReturnsNoIssues(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	for _, scope := range []Scope{ScopeGlobal, ScopeProject, ScopeWorkspace} {
		issues, err := Doctor(scope, projectRoot)
		require.NoError(t, err)
		assert.Empty(t, issues)
	}
}

func TestDoctor_ValidatesGlobalAndProjectForProjectScope(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeGlobalSidecar(t, &SidecarConfig{
		Path: "",
	})

	projectSidecarPath := filepath.Join(projectRoot, "project-sidecar")
	require.NoError(t, os.MkdirAll(filepath.Join(projectSidecarPath, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectSidecarPath, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectSidecarPath, "plans"), 0o755))
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectSidecarPath,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})

	issues, err := Doctor(ScopeProject, projectRoot)
	require.NoError(t, err)
	assert.Equal(t, []Issue{{
		Severity: IssueSeverityError,
		Message:  "sidecar global: path is required but empty",
		Path:     "",
	}}, issues)
}

func TestDoctor_ValidatesGlobalProjectWorkspaceForWorkspaceScope(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	writeGlobalSidecar(t, &SidecarConfig{
		Path: "",
	})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path: "",
	})

	workspaceSidecarPath := filepath.Join(globalDir, "workspace-sidecar")
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceSidecarPath, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceSidecarPath, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceSidecarPath, "plans"), 0o755))
	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name: "my-project",
				Path: projectRoot,
				Sidecar: &SidecarConfig{
					Path:     workspaceSidecarPath,
					DocsDir:  "docs",
					SpecsDir: "specs",
					PlansDir: "plans",
				},
			},
		},
		Active: "my-project",
	})

	issues, err := Doctor(ScopeWorkspace, projectRoot)
	require.NoError(t, err)

	assert.Len(t, issues, 2)
	assert.Equal(t, "sidecar global: path is required but empty", issues[0].Message)
	assert.Equal(t, IssueSeverityError, issues[0].Severity)
	assert.Equal(t, "sidecar project: path is required but empty", issues[1].Message)
	assert.Equal(t, IssueSeverityError, issues[1].Severity)
}

func TestDoctor_UsesLevelSpecificAnchors(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	globalSidecarRel := "global/sidecar"
	projectSidecarRel := "project/sidecar"
	workspaceSidecarRel := "workspace/sidecar"

	require.NoError(t, os.MkdirAll(filepath.Join(globalDir, globalSidecarRel, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(globalDir, globalSidecarRel, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(globalDir, globalSidecarRel, "plans"), 0o755))

	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, projectSidecarRel, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, projectSidecarRel, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, projectSidecarRel, "plans"), 0o755))

	require.NoError(t, os.MkdirAll(filepath.Join(globalDir, workspaceSidecarRel, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(globalDir, workspaceSidecarRel, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(globalDir, workspaceSidecarRel, "plans"), 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     globalSidecarRel,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     projectSidecarRel,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})
	writeWorkspaceConfig(t, &WorkspaceConfig{
		Workspaces: []WorkspaceEntry{
			{
				Name: "my-project",
				Path: projectRoot,
				Sidecar: &SidecarConfig{
					Path:     workspaceSidecarRel,
					DocsDir:  "docs",
					SpecsDir: "specs",
					PlansDir: "plans",
				},
			},
		},
		Active: "my-project",
	})

	projectIssues, err := Doctor(ScopeProject, projectRoot)
	require.NoError(t, err)
	assert.Empty(t, projectIssues)

	workspaceIssues, err := Doctor(ScopeWorkspace, projectRoot)
	require.NoError(t, err)
	assert.Empty(t, workspaceIssues)
}

func TestDoctor_EmptyPathIsError(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path: "",
	})

	issues, err := Doctor(ScopeProject, projectRoot)
	require.NoError(t, err)
	assert.Equal(t, []Issue{{
		Severity: IssueSeverityError,
		Message:  "sidecar project: path is required but empty",
		Path:     "",
	}}, issues)
}

func TestDoctor_MissingSidecarRootIsWarning(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	missingRoot := filepath.Join(projectRoot, "missing-sidecar-root")
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     missingRoot,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})

	issues, err := Doctor(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.NotEmpty(t, issues)
	for _, issue := range issues {
		assert.Equal(t, IssueSeverityWarning, issue.Severity)
	}
	assert.True(t, containsMessage(t, issues, fmt.Sprintf("sidecar project: path does not exist: %s", missingRoot)))
}

func TestDoctor_PathFileIsError(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sidecarFile := filepath.Join(projectRoot, "not-a-dir")
	require.NoError(t, os.WriteFile(sidecarFile, []byte("x"), 0o644))
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:     sidecarFile,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})

	issues, err := Doctor(ScopeProject, projectRoot)
	require.NoError(t, err)

	assert.True(t, containsMessage(t, issues, fmt.Sprintf("sidecar project: path is a file, not a directory: %s", sidecarFile)))
}

func TestDoctor_LinkedProjectsIgnored(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	sidecarPath := filepath.Join(globalDir, "linked-sidecar")
	require.NoError(t, os.MkdirAll(filepath.Join(sidecarPath, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sidecarPath, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sidecarPath, "plans"), 0o755))

	yamlPath := filepath.Join(globalDir, "sidecar.yaml")
	linkedConfig := map[string]any{
		"sidecar": map[string]any{
			"path":      sidecarPath,
			"docs_dir":  "docs",
			"specs_dir": "specs",
			"plans_dir": "plans",
			"linked_projects": []map[string]string{
				{
					"name": "other",
					"path": filepath.Join(globalDir, "missing-linked-project"),
				},
			},
		},
	}
	data, err := yaml.Marshal(linkedConfig)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(yamlPath, data, 0o644))

	issues, err := Doctor(ScopeGlobal, projectRoot)
	require.NoError(t, err)

	assert.Empty(t, issues)
	assert.False(t, hasMessagePrefix(t, issues, "sidecar global: linked project"))
}

func containsMessage(t *testing.T, issues []Issue, expected string) bool {
	t.Helper()
	for _, issue := range issues {
		if issue.Message == expected {
			return true
		}
	}
	return false
}

func hasMessagePrefix(t *testing.T, issues []Issue, prefix string) bool {
	t.Helper()
	for _, issue := range issues {
		if len(issue.Message) >= len(prefix) && issue.Message[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
