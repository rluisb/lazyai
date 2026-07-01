package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sidecarpkg "github.com/rluisb/lazyai/packages/cli/internal/sidecar"
	"github.com/spf13/cobra"
)

func TestSidecarStatus_PropagatesGetProjectRootError(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	origGetProjectRoot := getProjectRoot
	getProjectRoot = func() (string, error) {
		return "", fmt.Errorf("getting working directory: sentinel error")
	}
	t.Cleanup(func() { getProjectRoot = origGetProjectRoot })

	cmd := &cobra.Command{}

	_, _ = captureOutput(t, func() {
		err := runSidecarStatus(cmd, nil)
		if err == nil {
			t.Fatal("expected project-root propagation error")
		}
		if !strings.Contains(err.Error(), "getting project root") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSidecarDoctor_PropagatesGetProjectRootError(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	origGetProjectRoot := getProjectRoot
	getProjectRoot = func() (string, error) {
		return "", fmt.Errorf("getting working directory: sentinel error")
	}
	t.Cleanup(func() { getProjectRoot = origGetProjectRoot })

	cmd := &cobra.Command{}

	_, _ = captureOutput(t, func() {
		err := runSidecarDoctor(cmd, nil)
		if err == nil {
			t.Fatal("expected project-root propagation error")
		}
		if !strings.Contains(err.Error(), "getting project root") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSidecarInitWorkspaceScope_WritesToCwdLazyaiDir(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	cwd := t.TempDir()
	withWorkingDir(t, cwd)
	realCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	initCmd := &cobra.Command{}
	initCmd.Flags().String("scope", "", "")
	initCmd.Flags().String("path", "", "")
	initCmd.Flags().String("specs-dir", "", "")
	initCmd.Flags().String("docs-dir", "", "")
	initCmd.Flags().String("plans-dir", "", "")
	if err := initCmd.Flags().Set("scope", "workspace"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	sidecarRoot := t.TempDir()
	if err := initCmd.Flags().Set("path", sidecarRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		if err := runSidecarInit(initCmd, nil); err != nil {
			t.Fatalf("runSidecarInit: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadSidecarAt(realCwd)
	if err != nil {
		t.Fatalf("LoadSidecarAt: %v", err)
	}
	if cfg == nil || cfg.Path != sidecarRoot {
		t.Fatalf("expected sidecar at cwd/.lazyai with path %q, got %#v", sidecarRoot, cfg)
	}
	if _, err := os.Stat(filepath.Join(realCwd, ".lazyai", "sidecar.yaml")); err != nil {
		t.Fatalf("expected sidecar.yaml at cwd/.lazyai: %v", err)
	}
}

func TestSidecarInitProjectScope_WritesToCwdLazyaiDir(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	cwd := t.TempDir()
	withWorkingDir(t, cwd)
	realCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	initCmd := &cobra.Command{}
	initCmd.Flags().String("scope", "", "")
	initCmd.Flags().String("path", "", "")
	initCmd.Flags().String("specs-dir", "", "")
	initCmd.Flags().String("docs-dir", "", "")
	initCmd.Flags().String("plans-dir", "", "")
	if err := initCmd.Flags().Set("scope", "project"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	sidecarRoot := t.TempDir()
	if err := initCmd.Flags().Set("path", sidecarRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		if err := runSidecarInit(initCmd, nil); err != nil {
			t.Fatalf("runSidecarInit: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadSidecarAt(realCwd)
	if err != nil {
		t.Fatalf("LoadSidecarAt: %v", err)
	}
	if cfg == nil || cfg.Path != sidecarRoot {
		t.Fatalf("expected sidecar at cwd/.lazyai with path %q, got %#v", sidecarRoot, cfg)
	}
	if _, err := os.Stat(filepath.Join(realCwd, ".lazyai", "sidecar.yaml")); err != nil {
		t.Fatalf("expected sidecar.yaml at cwd/.lazyai: %v", err)
	}
}

func TestSidecarInitGlobalScope_WritesToHomeLazyaiDir(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	initCmd := &cobra.Command{}
	initCmd.Flags().String("scope", "", "")
	initCmd.Flags().String("path", "", "")
	initCmd.Flags().String("specs-dir", "", "")
	initCmd.Flags().String("docs-dir", "", "")
	initCmd.Flags().String("plans-dir", "", "")
	if err := initCmd.Flags().Set("scope", "global"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	sidecarRoot := t.TempDir()
	if err := initCmd.Flags().Set("path", sidecarRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		if err := runSidecarInit(initCmd, nil); err != nil {
			t.Fatalf("runSidecarInit: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadSidecarAt(home)
	if err != nil {
		t.Fatalf("LoadSidecarAt: %v", err)
	}
	if cfg == nil || cfg.Path != sidecarRoot {
		t.Fatalf("expected global sidecar with path %q, got %#v", sidecarRoot, cfg)
	}
	if _, err := os.Stat(filepath.Join(home, ".lazyai", "sidecar.yaml")); err != nil {
		t.Fatalf("expected sidecar.yaml at home/.lazyai: %v", err)
	}
}

func TestSidecarStatus_ShowsAllDiscoveredLayersAndProvenance(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	cwd := t.TempDir()
	withWorkingDir(t, cwd)
	realCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	globalSidecarRoot := t.TempDir()
	if err := sidecarpkg.WriteSidecarAt(home, &sidecarpkg.SidecarConfig{
		Path:     globalSidecarRoot,
		DocsDir:  "gdocs",
		SpecsDir: "gspecs",
		PlansDir: "gplans",
	}); err != nil {
		t.Fatalf("WriteSidecarAt (global): %v", err)
	}

	projectSidecarRoot := t.TempDir()
	if err := sidecarpkg.WriteSidecarAt(realCwd, &sidecarpkg.SidecarConfig{
		Path:    projectSidecarRoot,
		DocsDir: "pdocs",
	}); err != nil {
		t.Fatalf("WriteSidecarAt (project): %v", err)
	}

	cmd := &cobra.Command{}
	output, _ := captureOutput(t, func() {
		if err := runSidecarStatus(cmd, nil); err != nil {
			t.Fatalf("runSidecarStatus: %v", err)
		}
	})

	if !strings.Contains(output, "global") {
		t.Fatalf("expected global layer line, got: %s", output)
	}
	if !strings.Contains(output, "workspace") {
		t.Fatalf("expected workspace layer line, got: %s", output)
	}
	if !strings.Contains(output, "project") {
		t.Fatalf("expected project layer line, got: %s", output)
	}
	if !strings.Contains(output, "(from: project)") {
		t.Fatalf("expected docs_dir provenance from project, got: %s", output)
	}
	if !strings.Contains(output, "(from: global)") {
		t.Fatalf("expected specs_dir/plans_dir provenance from global, got: %s", output)
	}
}

func TestSidecarStatus_PrintsGuidanceWhenNoLayerFound(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	cmd := &cobra.Command{}
	output, _ := captureOutput(t, func() {
		if err := runSidecarStatus(cmd, nil); err != nil {
			t.Fatalf("runSidecarStatus: %v", err)
		}
	})

	if !strings.Contains(output, "No .lazyai/ configuration found — using built-in defaults. Run 'sidecar init' to configure.") {
		t.Fatalf("expected guidance line, got: %s", output)
	}
}

func TestSidecarDoctor_PrefixesIssuesWithLevel(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	cwd := t.TempDir()
	withWorkingDir(t, cwd)
	realCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	missingRoot := filepath.Join(cwd, "does-not-exist")
	if err := sidecarpkg.WriteSidecarAt(realCwd, &sidecarpkg.SidecarConfig{
		Path: missingRoot,
	}); err != nil {
		t.Fatalf("WriteSidecarAt (project): %v", err)
	}

	cmd := &cobra.Command{}
	output, _ := captureOutput(t, func() {
		if err := runSidecarDoctor(cmd, nil); err != nil {
			t.Fatalf("runSidecarDoctor: %v", err)
		}
	})

	if !strings.Contains(output, "[project] WARN:") {
		t.Fatalf("expected issue prefixed with [project] WARN, got: %s", output)
	}
}

// TestSidecarInit_DefaultScopeIsProjectAndWritesOnlyToCwd is the #578
// regression test: running `sidecar init` with no --scope flag at all must
// default to ScopeProject, resolve via ScopeRoot to exactly cwd, and write
// cwd/.lazyai/sidecar.yaml -- never any other path, and never touch
// ~/.lazyai/workspaces.yaml (that whole registry concept is gone).
func TestSidecarInit_DefaultScopeIsProjectAndWritesOnlyToCwd(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	cwd := t.TempDir()
	withWorkingDir(t, cwd)
	realCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	initCmd := &cobra.Command{}
	initCmd.Flags().String("scope", "", "")
	initCmd.Flags().String("path", "", "")
	initCmd.Flags().String("specs-dir", "", "")
	initCmd.Flags().String("docs-dir", "", "")
	initCmd.Flags().String("plans-dir", "", "")
	sidecarRoot := t.TempDir()
	if err := initCmd.Flags().Set("path", sidecarRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		if err := runSidecarInit(initCmd, nil); err != nil {
			t.Fatalf("runSidecarInit: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadSidecarAt(realCwd)
	if err != nil {
		t.Fatalf("LoadSidecarAt: %v", err)
	}
	if cfg == nil || cfg.Path != sidecarRoot {
		t.Fatalf("expected sidecar at cwd/.lazyai with path %q, got %#v", sidecarRoot, cfg)
	}
	if _, err := os.Stat(filepath.Join(realCwd, ".lazyai", "sidecar.yaml")); err != nil {
		t.Fatalf("expected sidecar.yaml at cwd/.lazyai: %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".lazyai", "workspaces.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected no workspaces.yaml to be created under $HOME, stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".lazyai")); !os.IsNotExist(err) {
		t.Fatalf("expected no files written under $HOME at all for default (project) scope init, stat err: %v", err)
	}
}
