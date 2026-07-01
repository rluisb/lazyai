package sidecar

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupDiscoveryTestEnv creates a temporary home directory and overrides
// HOME/USERPROFILE for the duration of the test, mirroring resolver_test.go's
// setupTestEnv but scoped to discovery tests, which need to construct
// arbitrary multi-level ancestor directories rather than a fixed
// home/project pair.
func setupDiscoveryTestEnv(t *testing.T) (homeDir string) {
	t.Helper()

	tmpDir := t.TempDir()
	homeDir = filepath.Join(tmpDir, "home")
	require.NoError(t, os.MkdirAll(homeDir, 0o755))

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", homeDir)
	os.Setenv("USERPROFILE", homeDir)

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	})

	return homeDir
}

// mkdirsAt creates dir (and all parents) under 0o755 permissions.
func mkdirsAt(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
}

// writeSidecarAtDir writes a minimal <dir>/.lazyai/sidecar.yaml layer hit.
func writeSidecarAtDir(t *testing.T, dir string, cfg *SidecarConfig) {
	t.Helper()
	lazyaiDir := filepath.Join(dir, ".lazyai")
	require.NoError(t, os.MkdirAll(lazyaiDir, 0o755))
	data, err := yaml.Marshal(SidecarFile{Sidecar: cfg})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(lazyaiDir, "sidecar.yaml"), data, 0o644))
}

func TestDiscoverLayers_ProjectOnly(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	cwd := filepath.Join(homeDir, "workroot", "project")
	mkdirsAt(t, cwd)
	writeSidecarAtDir(t, cwd, &SidecarConfig{Path: "."})

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	require.NotNil(t, layers.Project)
	assert.Equal(t, "project", layers.Project.Level)
	assert.Equal(t, cwd, layers.Project.Root)
	assert.Nil(t, layers.Workspace)
}

func TestDiscoverLayers_WorkspaceOnly(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	ancestor := filepath.Join(homeDir, "workroot")
	cwd := filepath.Join(ancestor, "project")
	mkdirsAt(t, cwd)
	writeSidecarAtDir(t, ancestor, &SidecarConfig{Path: "."})

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	assert.Nil(t, layers.Project)
	require.NotNil(t, layers.Workspace)
	assert.Equal(t, "workspace", layers.Workspace.Level)
	assert.Equal(t, ancestor, layers.Workspace.Root)
}

func TestDiscoverLayers_BothProjectAndWorkspace(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	ancestor := filepath.Join(homeDir, "workroot")
	cwd := filepath.Join(ancestor, "project")
	mkdirsAt(t, cwd)
	writeSidecarAtDir(t, ancestor, &SidecarConfig{Path: "."})
	writeSidecarAtDir(t, cwd, &SidecarConfig{Path: "."})

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	require.NotNil(t, layers.Project)
	require.NotNil(t, layers.Workspace)
	assert.Equal(t, cwd, layers.Project.Root)
	assert.Equal(t, ancestor, layers.Workspace.Root)
	assert.NotEqual(t, layers.Project.Root, layers.Workspace.Root)
}

func TestDiscoverLayers_StopsAtNearestWorkspaceHit(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	farAncestor := filepath.Join(homeDir, "a")
	nearAncestor := filepath.Join(homeDir, "a", "b")
	cwd := filepath.Join(homeDir, "a", "b", "c")
	mkdirsAt(t, cwd)
	writeSidecarAtDir(t, farAncestor, &SidecarConfig{Path: "far"})
	writeSidecarAtDir(t, nearAncestor, &SidecarConfig{Path: "near"})

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	require.NotNil(t, layers.Workspace)
	assert.Equal(t, nearAncestor, layers.Workspace.Root)
	assert.Equal(t, "near", layers.Workspace.Config.Path)
}

func TestDiscoverLayers_StopsBeforeHome(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	cwd := filepath.Join(homeDir, "a", "b", "c")
	mkdirsAt(t, cwd)
	// Layer hit at $HOME itself — must never be treated as workspace.
	writeSidecarAtDir(t, homeDir, &SidecarConfig{Path: "."})

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	assert.Nil(t, layers.Workspace)
}

func TestDiscoverLayers_CwdEqualsHome(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	writeSidecarAtDir(t, homeDir, &SidecarConfig{Path: "."})

	layers, err := DiscoverLayers(homeDir)
	require.NoError(t, err)

	require.NotNil(t, layers.Project)
	assert.Equal(t, homeDir, layers.Project.Root)
	assert.Nil(t, layers.Workspace)
}

func TestDiscoverLayers_CwdOutsideHomeWalksToRoot(t *testing.T) {
	setupDiscoveryTestEnv(t)
	// cwd lives under a tempdir unrelated to $HOME, so the $HOME boundary
	// check never trips and the walk must run all the way to the fs root
	// without erroring or early-terminating.
	outsideRoot := t.TempDir()
	cwd := filepath.Join(outsideRoot, "x", "y", "z")
	mkdirsAt(t, cwd)

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)
	// No layer placed anywhere in this chain — walk must complete cleanly
	// having checked every ancestor up to (and including) the fs root.
	assert.Nil(t, layers.Workspace)
	assert.Nil(t, layers.Project)
}

func TestDiscoverLayers_BareLazyaiDirWithoutSidecarYamlIsNotALayer(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	ancestor := filepath.Join(homeDir, "workroot")
	cwd := filepath.Join(ancestor, "project")
	mkdirsAt(t, cwd)
	// Bare .lazyai/ dir with an unrelated file inside, no sidecar.yaml.
	bareLazyaiDir := filepath.Join(ancestor, ".lazyai")
	mkdirsAt(t, bareLazyaiDir)
	require.NoError(t, os.WriteFile(filepath.Join(bareLazyaiDir, "notes.txt"), []byte("hi"), 0o644))

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	assert.Nil(t, layers.Project)
	assert.Nil(t, layers.Workspace)
}

func TestDiscoverLayers_SymlinkedCwdResolvesForHomeComparison(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}
	tmpDir := t.TempDir()
	actualHome := filepath.Join(tmpDir, "actualHome")
	mkdirsAt(t, actualHome)

	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", actualHome)
	os.Setenv("USERPROFILE", actualHome)
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	})

	// A symlink elsewhere in the tree whose real target is inside $HOME.
	// The lexical ancestor of the symlink itself ("tmpDir") is NOT inside
	// $HOME, but the resolved target is — so the walk must use the resolved
	// path when comparing against $HOME, not raw lexical ancestors.
	realNested := filepath.Join(actualHome, "nested", "deep")
	mkdirsAt(t, realNested)
	symlinkPath := filepath.Join(tmpDir, "shortcut")
	require.NoError(t, os.Symlink(realNested, symlinkPath))

	// Layer hit placed at the real $HOME — must not be picked up as
	// workspace once the walk (via resolved-path comparison) reaches it.
	writeSidecarAtDir(t, actualHome, &SidecarConfig{Path: "."})

	layers, err := DiscoverLayers(symlinkPath)
	require.NoError(t, err)

	assert.Nil(t, layers.Workspace)
}

func TestDiscoverLayers_BrokenSymlinkFallsBackToRawPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}
	homeDir := setupDiscoveryTestEnv(t)
	outer := filepath.Join(homeDir, "outer")
	mkdirsAt(t, outer)
	writeSidecarAtDir(t, outer, &SidecarConfig{Path: "outer-hit"})

	// A dangling symlink directory between "outer" and cwd — EvalSymlinks
	// will fail to resolve it, and discovery must fall back to the raw path
	// rather than aborting.
	danglingLink := filepath.Join(outer, "broken")
	require.NoError(t, os.Symlink(filepath.Join(outer, "does-not-exist"), danglingLink))
	cwd := filepath.Join(danglingLink, "leaf")

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	require.NotNil(t, layers.Workspace)
	assert.Equal(t, outer, layers.Workspace.Root)
	assert.Equal(t, "outer-hit", layers.Workspace.Config.Path)
}

func TestDiscoverLayers_GlobalAlwaysPresentEvenWhenSidecarYamlAbsent(t *testing.T) {
	homeDir := setupDiscoveryTestEnv(t)
	cwd := filepath.Join(homeDir, "workroot", "project")
	mkdirsAt(t, cwd)
	// Deliberately no ~/.lazyai/sidecar.yaml written.

	layers, err := DiscoverLayers(cwd)
	require.NoError(t, err)

	assert.Equal(t, "global", layers.Global.Level)
	assert.Equal(t, homeDir, layers.Global.Root)
	assert.Nil(t, layers.Global.Config)
}
