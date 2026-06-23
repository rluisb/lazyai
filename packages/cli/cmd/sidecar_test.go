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
	cmd.Flags().String("scope", "", "")
	if err := cmd.Flags().Set("scope", "project"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}

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
	cmd.Flags().String("scope", "", "")
	if err := cmd.Flags().Set("scope", "project"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}

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
func TestSidecarDetachProject_DoesNotCallGetProjectRootWithPositionalArg(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	missingConfigProject := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(missingConfigProject, 0o755); err != nil {
		t.Fatalf("prepare project path: %v", err)
	}

	origGetProjectRoot := getProjectRoot
	getProjectRoot = func() (string, error) {
		t.Fatal("getProjectRoot should not be called when project positional arg is provided")
		return "", fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { getProjectRoot = origGetProjectRoot })

	cmd := &cobra.Command{}
	cmd.Flags().String("scope", "", "")
	cmd.Flags().Bool("force", false, "")
	if err := cmd.Flags().Set("scope", "project"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	if err := cmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("set force flag: %v", err)
	}

	output, _ := captureOutput(t, func() {
		if err := runSidecarDetach(cmd, []string{missingConfigProject}); err != nil {
			t.Fatalf("runSidecarDetach: %v", err)
		}
	})
	if !strings.Contains(output, "No sidecar configured for project.") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestSidecarAttachProject_RejectsMissingProjectPathWithoutCreatingIt(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	missingProject := filepath.Join(t.TempDir(), "missing-project")
	projectSidecarRoot := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().String("scope", "", "")
	cmd.Flags().String("path", "", "")
	cmd.Flags().String("specs-dir", "", "")
	cmd.Flags().String("docs-dir", "", "")
	cmd.Flags().String("plans-dir", "", "")
	if err := cmd.Flags().Set("scope", "project"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	if err := cmd.Flags().Set("path", projectSidecarRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		err := runSidecarAttach(cmd, []string{missingProject})
		if err == nil {
			t.Fatal("expected missing project path error")
		}
		if !strings.Contains(err.Error(), "project path not accessible") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if _, err := os.Stat(filepath.Join(missingProject, ".lazyai-sidecar.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected missing sidecar file, got: %v", err)
	}
}

func TestDetermineScope_ReturnsWorkspaceConfigLoadError(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)

	withWorkingDir(t, t.TempDir())
	configPath := filepath.Join(home, ".lazyai", "workspaces.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("workspaces: [\n"), 0o644); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("scope", "", "")

	if _, err := determineScope(cmd, false); err == nil {
		t.Fatal("expected workspace config load error")
	} else if !strings.Contains(err.Error(), "parsing workspace config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSidecarInitWorkspace_UsesLockedWorkspaceUpdate(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	workspaceRoot := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")
	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{workspaceRoot}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

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

	cfg, err := sidecarpkg.LoadWorkspaceConfig()
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig: %v", err)
	}
	active := findWorkspace(cfg, cfg.Active)
	if active == nil {
		t.Fatal("expected active workspace")
	}
	if active.Sidecar == nil || active.Sidecar.Path != sidecarRoot {
		t.Fatalf("expected workspace sidecar %q, got %#v", sidecarRoot, active.Sidecar)
	}
}

func TestSidecarAttachWorkspace_UsesLockedWorkspaceUpdate(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	workspaceRoot := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")
	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{workspaceRoot}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

	attachCmd := &cobra.Command{}
	attachCmd.Flags().String("scope", "", "")
	attachCmd.Flags().String("path", "", "")
	attachCmd.Flags().String("specs-dir", "", "")
	attachCmd.Flags().String("docs-dir", "", "")
	attachCmd.Flags().String("plans-dir", "", "")
	if err := attachCmd.Flags().Set("scope", "workspace"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	sidecarRoot := t.TempDir()
	if err := attachCmd.Flags().Set("path", sidecarRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		if err := runSidecarAttach(attachCmd, nil); err != nil {
			t.Fatalf("runSidecarAttach: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadWorkspaceConfig()
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig: %v", err)
	}
	active := findWorkspace(cfg, cfg.Active)
	if active == nil {
		t.Fatal("expected active workspace")
	}
	if active.Sidecar == nil || active.Sidecar.Path != sidecarRoot {
		t.Fatalf("expected workspace sidecar %q, got %#v", sidecarRoot, active.Sidecar)
	}
}

func TestSidecarDetachWorkspace_UsesLockedWorkspaceUpdate(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	withWorkingDir(t, t.TempDir())

	workspaceRoot := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")
	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{workspaceRoot}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

	attachRoot := t.TempDir()
	attachCmd := &cobra.Command{}
	attachCmd.Flags().String("scope", "", "")
	attachCmd.Flags().String("path", "", "")
	attachCmd.Flags().String("specs-dir", "", "")
	attachCmd.Flags().String("docs-dir", "", "")
	attachCmd.Flags().String("plans-dir", "", "")
	if err := attachCmd.Flags().Set("scope", "workspace"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	if err := attachCmd.Flags().Set("path", attachRoot); err != nil {
		t.Fatalf("set path flag: %v", err)
	}
	_, _ = captureOutput(t, func() {
		if err := runSidecarAttach(attachCmd, nil); err != nil {
			t.Fatalf("runSidecarAttach: %v", err)
		}
	})

	detachCmd := &cobra.Command{}
	detachCmd.Flags().String("scope", "", "")
	detachCmd.Flags().Bool("force", false, "")
	if err := detachCmd.Flags().Set("scope", "workspace"); err != nil {
		t.Fatalf("set scope flag: %v", err)
	}
	if err := detachCmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("set force flag: %v", err)
	}

	_, _ = captureOutput(t, func() {
		if err := runSidecarDetach(detachCmd, nil); err != nil {
			t.Fatalf("runSidecarDetach: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadWorkspaceConfig()
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig: %v", err)
	}
	active := findWorkspace(cfg, cfg.Active)
	if active == nil {
		t.Fatal("expected active workspace")
	}
	if active.Sidecar != nil {
		t.Fatalf("expected workspace sidecar removed, got %#v", active.Sidecar)
	}
}
