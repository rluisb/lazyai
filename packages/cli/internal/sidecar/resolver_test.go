package sidecar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// writeGlobalSidecar writes the current $HOME's .lazyai/sidecar.yaml.
// Rewritten to target the unified SidecarFile on-disk shape (via
// writeSidecarAtDir, discovery_test.go) instead of the deleted
// GlobalSidecarConfig wrapper / flat-file layout.
func writeGlobalSidecar(t *testing.T, cfg *SidecarConfig) {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	writeSidecarAtDir(t, home, cfg)
}

// writeProjectSidecar writes <projectRoot>/.lazyai/sidecar.yaml. Rewritten
// to target the unified SidecarFile shape instead of the deleted flat
// ".lazyai-sidecar.yaml" / ProjectSidecarConfig wrapper.
func writeProjectSidecar(t *testing.T, projectRoot string, cfg *SidecarConfig) {
	t.Helper()
	writeSidecarAtDir(t, projectRoot, cfg)
}

func TestResolve_NoSidecarAnyLevel(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// No layer anywhere: no global sidecar, no workspace ancestor, no
	// project sidecar.
	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.True(t, result.IsAllDefault())
	assert.Equal(t, filepath.Join(projectRoot, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectRoot, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectRoot, "plans"), result.PlansDir)
}

func TestResolve_GlobalOnly(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sidecarPath := filepath.Join(homeDir, "kb")
	require.NoError(t, os.MkdirAll(sidecarPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     sidecarPath,
		SpecsDir: "my-specs",
		DocsDir:  "my-docs",
		PlansDir: "my-plans",
	})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(sidecarPath, "my-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(sidecarPath, "my-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(sidecarPath, "my-plans"), result.PlansDir)
	assert.Equal(t, map[string]string{
		"docs_dir":  "global",
		"specs_dir": "global",
		"plans_dir": "global",
	}, result.Provenance)
}

func TestResolve_ProjectOverridesGlobal(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	globalPath := filepath.Join(homeDir, "global-kb")
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

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(projectPath, "project-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(projectPath, "project-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(projectPath, "project-plans"), result.PlansDir)
	assert.Equal(t, map[string]string{
		"docs_dir":  "project",
		"specs_dir": "project",
		"plans_dir": "project",
	}, result.Provenance)
}

func TestResolve_RelativePathResolution(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Global sidecar with a relative Path anchors on the global layer's
	// own Root (home), NOT home/.lazyai — spec.md §3.3 Refinement.
	kbPath := filepath.Join(homeDir, "kb")
	require.NoError(t, os.MkdirAll(kbPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     "kb", // relative to homeDir
		SpecsDir: "specs",
		DocsDir:  "docs",
		PlansDir: "plans",
	})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(kbPath, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(kbPath, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(kbPath, "plans"), result.PlansDir)
	assert.Equal(t, "global", result.Provenance["docs_dir"])
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

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(absPath, "docs"), result.DocsDir)
}

func TestResolve_DefaultDirValues(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	kbPath := filepath.Join(homeDir, "kb")
	require.NoError(t, os.MkdirAll(kbPath, 0o755))

	// Global sidecar sets Path but leaves every *_dir field blank. Under
	// field-level merge, a blank RAW field means "not set by this layer" —
	// docs/specs/plans all fall through to the true no-layer-sets-this-
	// field default, anchored on the deepest present layer's Root (global's
	// Root == homeDir), NOT on kbPath.
	writeGlobalSidecar(t, &SidecarConfig{
		Path: kbPath,
	})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(homeDir, "docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(homeDir, "specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(homeDir, "plans"), result.PlansDir)
	assert.Equal(t, map[string]string{
		"docs_dir":  "default",
		"specs_dir": "default",
		"plans_dir": "default",
	}, result.Provenance)
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

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, absDocs, result.DocsDir)
	assert.Equal(t, absSpecs, result.SpecsDir)
	assert.Equal(t, filepath.Join(projectRoot, "plans"), result.PlansDir)
}

func TestResolve_RejectsEmptySidecarPath(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path: "",
	})

	_, err := Resolve(projectRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sidecar path is required but empty")
}

// TestMergeLayers_PartialProjectConfigInheritsRemainingFieldsFromWorkspace
// directly supersedes the deleted TestResolve_HigherPrioritySidecarReplacesWholeTuple:
// a project layer setting only docs_dir must NOT wipe out workspace's
// specs_dir/plans_dir, and global (which also sets all three) must lose to
// workspace for those two fields since workspace outranks it.
func TestMergeLayers_PartialProjectConfigInheritsRemainingFieldsFromWorkspace(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// filepath.Dir(projectRoot) sits strictly above projectRoot and is not
	// homeDir (a sibling, not an ancestor) — a valid workspace-layer root.
	workspaceRoot := filepath.Dir(projectRoot)

	globalPath := filepath.Join(homeDir, "global-kb")
	workspacePath := filepath.Join(workspaceRoot, "workspace-kb")
	projectPath := filepath.Join(projectRoot, "project-kb")
	require.NoError(t, os.MkdirAll(globalPath, 0o755))
	require.NoError(t, os.MkdirAll(workspacePath, 0o755))
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     globalPath,
		DocsDir:  "global-docs",
		SpecsDir: "global-specs",
		PlansDir: "global-plans",
	})
	writeSidecarAtDir(t, workspaceRoot, &SidecarConfig{
		Path:     workspacePath,
		SpecsDir: "workspace-specs",
		PlansDir: "workspace-plans",
	})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:    projectPath,
		DocsDir: "project-docs",
	})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(projectPath, "project-docs"), result.DocsDir)
	assert.Equal(t, "project", result.Provenance["docs_dir"])
	assert.Equal(t, filepath.Join(workspacePath, "workspace-specs"), result.SpecsDir)
	assert.Equal(t, "workspace", result.Provenance["specs_dir"])
	assert.Equal(t, filepath.Join(workspacePath, "workspace-plans"), result.PlansDir)
	assert.Equal(t, "workspace", result.Provenance["plans_dir"])
}

func TestMergeLayers_ProjectFieldWinsOverWorkspaceAndGlobal(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	workspaceRoot := filepath.Dir(projectRoot)

	globalPath := filepath.Join(homeDir, "global-kb")
	workspacePath := filepath.Join(workspaceRoot, "workspace-kb")
	projectPath := filepath.Join(projectRoot, "project-kb")
	require.NoError(t, os.MkdirAll(globalPath, 0o755))
	require.NoError(t, os.MkdirAll(workspacePath, 0o755))
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{Path: globalPath, DocsDir: "global-docs"})
	writeSidecarAtDir(t, workspaceRoot, &SidecarConfig{Path: workspacePath, DocsDir: "workspace-docs"})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{Path: projectPath, DocsDir: "project-docs"})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(projectPath, "project-docs"), result.DocsDir)
	assert.Equal(t, "project", result.Provenance["docs_dir"])
}

func TestMergeLayers_ProvenanceReflectsActualSourcePerField(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	workspaceRoot := filepath.Dir(projectRoot)

	globalPath := filepath.Join(homeDir, "global-kb")
	workspacePath := filepath.Join(workspaceRoot, "workspace-kb")
	projectPath := filepath.Join(projectRoot, "project-kb")
	require.NoError(t, os.MkdirAll(globalPath, 0o755))
	require.NoError(t, os.MkdirAll(workspacePath, 0o755))
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	// project sets only docs_dir; workspace sets only specs_dir; global
	// sets neither, and no layer sets plans_dir at all -> "default".
	writeGlobalSidecar(t, &SidecarConfig{Path: globalPath})
	writeSidecarAtDir(t, workspaceRoot, &SidecarConfig{Path: workspacePath, SpecsDir: "workspace-specs"})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{Path: projectPath, DocsDir: "project-docs"})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"docs_dir":  "project",
		"specs_dir": "workspace",
		"plans_dir": "default",
	}, result.Provenance)
	assert.Equal(t, filepath.Join(projectPath, "project-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(workspacePath, "workspace-specs"), result.SpecsDir)
	// No layer sets plans_dir; deepest present root is the project layer's
	// own Root (cwd), matching today's single-tuple fallback intent.
	assert.Equal(t, filepath.Join(projectRoot, "plans"), result.PlansDir)
}

// TestMergeLayers_GlobalAlwaysAppliesAsBaseLayer is the third required case
// for spec.md §8's enforcement row 3: global must still surface a field
// neither project nor workspace touches, even though a "closer" layer
// (project) is present for other fields.
func TestMergeLayers_GlobalAlwaysAppliesAsBaseLayer(t *testing.T) {
	homeDir, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	globalPath := filepath.Join(homeDir, "global-kb")
	projectPath := filepath.Join(projectRoot, "project-kb")
	require.NoError(t, os.MkdirAll(globalPath, 0o755))
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	writeGlobalSidecar(t, &SidecarConfig{
		Path:     globalPath,
		SpecsDir: "global-specs",
	})
	writeProjectSidecar(t, projectRoot, &SidecarConfig{
		Path:    projectPath,
		DocsDir: "project-docs",
	})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(globalPath, "global-specs"), result.SpecsDir)
	assert.Equal(t, "global", result.Provenance["specs_dir"])
	assert.Equal(t, filepath.Join(projectPath, "project-docs"), result.DocsDir)
	assert.Equal(t, "project", result.Provenance["docs_dir"])
}

// TestResolve_WorkspaceRelativePathUsesDiscoveredWorkspaceRoot supersedes
// the deleted TestResolve_WorkspaceRelativePathUsesGlobalConfigDirAtWorkspaceScope,
// which locked in the bug this refactor fixes (workspace anchored on
// globalDir instead of its own discovered Root).
func TestResolve_WorkspaceRelativePathUsesDiscoveredWorkspaceRoot(t *testing.T) {
	_, projectRoot, _, cleanup := setupTestEnv(t)
	defer cleanup()

	workspaceRoot := filepath.Dir(projectRoot)

	workspaceKB := filepath.Join(workspaceRoot, "workspace-kb")
	require.NoError(t, os.MkdirAll(workspaceKB, 0o755))

	writeSidecarAtDir(t, workspaceRoot, &SidecarConfig{
		Path:     "workspace-kb", // relative -> must anchor on workspaceRoot
		SpecsDir: "workspace-specs",
		DocsDir:  "workspace-docs",
		PlansDir: "workspace-plans",
	})

	result, err := Resolve(projectRoot)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(workspaceKB, "workspace-docs"), result.DocsDir)
	assert.Equal(t, filepath.Join(workspaceKB, "workspace-specs"), result.SpecsDir)
	assert.Equal(t, filepath.Join(workspaceKB, "workspace-plans"), result.PlansDir)
	assert.Equal(t, "workspace", result.Provenance["docs_dir"])
}
