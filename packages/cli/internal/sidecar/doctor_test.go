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

	issues, err := Doctor(projectRoot)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestDoctor_ValidatesGlobalAndProjectWhenBothPresent(t *testing.T) {
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

	issues, err := Doctor(projectRoot)
	require.NoError(t, err)
	assert.Equal(t, []Issue{{
		Severity: IssueSeverityError,
		Message:  "sidecar global: path is required but empty",
		Path:     "",
		Level:    "global",
	}}, issues)
}

func TestDoctor_ValidatesAllThreeLayersWhenAllPresent(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// ancestorDir sits strictly between cwd and the real filesystem root and
	// is unrelated to $HOME, so it is discovered as the workspace layer.
	ancestorDir := filepath.Join(t.TempDir(), "workspace-root")
	cwd := filepath.Join(ancestorDir, "project")
	mkdirsAt(t, cwd)

	writeGlobalSidecar(t, &SidecarConfig{Path: ""})
	writeSidecarAtDir(t, ancestorDir, &SidecarConfig{Path: ""})
	writeSidecarAtDir(t, cwd, &SidecarConfig{Path: ""})

	issues, err := Doctor(cwd)
	require.NoError(t, err)

	require.Len(t, issues, 3)
	assert.Equal(t, "global", issues[0].Level)
	assert.Equal(t, "sidecar global: path is required but empty", issues[0].Message)
	assert.Equal(t, "workspace", issues[1].Level)
	assert.Equal(t, "sidecar workspace: path is required but empty", issues[1].Message)
	assert.Equal(t, "project", issues[2].Level)
	assert.Equal(t, "sidecar project: path is required but empty", issues[2].Message)
}

func TestDoctor_WorkspaceAnchoredOnOwnDiscoveredRoot(t *testing.T) {
	_, _, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// ancestorDir is the discovered workspace root; workspaceSidecarRel's
	// docs/specs/plans directories only exist relative to ancestorDir, NOT
	// anywhere near $HOME/.lazyai. Prior to the anchor-unification fix,
	// Doctor anchored the workspace layer's relative Path against
	// globalDir, which would report these directories missing.
	ancestorDir := filepath.Join(t.TempDir(), "workspace-root")
	cwd := filepath.Join(ancestorDir, "project")
	mkdirsAt(t, cwd)

	workspaceSidecarRel := "workspace-sidecar"
	require.NoError(t, os.MkdirAll(filepath.Join(ancestorDir, workspaceSidecarRel, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(ancestorDir, workspaceSidecarRel, "specs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(ancestorDir, workspaceSidecarRel, "plans"), 0o755))

	writeSidecarAtDir(t, ancestorDir, &SidecarConfig{
		Path:     workspaceSidecarRel,
		DocsDir:  "docs",
		SpecsDir: "specs",
		PlansDir: "plans",
	})

	issues, err := Doctor(cwd)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestDoctor_EmptyPathIsError(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path: "",
	})

	issues, err := Doctor(projectRoot)
	require.NoError(t, err)
	assert.Equal(t, []Issue{{
		Severity: IssueSeverityError,
		Message:  "sidecar project: path is required but empty",
		Path:     "",
		Level:    "project",
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

	issues, err := Doctor(projectRoot)
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

	issues, err := Doctor(projectRoot)
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

	issues, err := Doctor(projectRoot)
	require.NoError(t, err)

	assert.Empty(t, issues)
	assert.False(t, hasMessagePrefix(t, issues, "sidecar global: linked project"))
}

func TestDoctor_DetectsStaleWorkspacesYamlAndWarns(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(globalDir, "workspaces.yaml"), []byte("workspaces: []\n"), 0o644))

	issues, err := Doctor(projectRoot)
	require.NoError(t, err)

	assert.Contains(t, issues, Issue{
		Severity: IssueSeverityWarning,
		Level:    "",
		Message:  "found legacy ~/.lazyai/workspaces.yaml (workspace registry removed in #579) — re-create any workspace/project sidecars you still need with 'sidecar init' from the intended directory",
	})
}

func TestDoctor_NoStaleWorkspacesYamlNoWarning(t *testing.T) {
	_, projectRoot, globalDir, cleanup := setupTestEnv(t)
	defer cleanup()

	_, statErr := os.Stat(filepath.Join(globalDir, "workspaces.yaml"))
	require.True(t, os.IsNotExist(statErr))

	writeGlobalSidecar(t, &SidecarConfig{Path: ""})

	issues, err := Doctor(projectRoot)
	require.NoError(t, err)

	for _, issue := range issues {
		assert.NotContains(t, issue.Message, "workspaces.yaml")
	}
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
