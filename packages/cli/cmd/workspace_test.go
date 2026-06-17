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

func TestWorkspaceAdd_CreatesConfigAndAutoActivatesFirstWorkspace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withWorkingDir(t, t.TempDir())

	project := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")

	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{project}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadWorkspaceConfig()
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig: %v", err)
	}
	if len(cfg.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(cfg.Workspaces))
	}
	expectedName := defaultNameFromPath(project)
	if got := cfg.Workspaces[0].Name; got != expectedName {
		t.Fatalf("workspace name = %q, want %q", got, expectedName)
	}
	if cfg.Workspaces[0].Path != project {
		t.Fatalf("workspace path = %q, want %q", cfg.Workspaces[0].Path, project)
	}
	if cfg.Active != expectedName {
		t.Fatalf("active workspace = %q, want %q", cfg.Active, expectedName)
	}

	if _, err := os.Stat(filepath.Join(home, ".lazyai", "workspaces.yaml")); err != nil {
		t.Fatalf("expected workspaces.yaml to be created: %v", err)
	}
}

func TestWorkspaceAdd_RejectsDuplicateName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withWorkingDir(t, t.TempDir())

	project := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")

	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{project}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

	_, _ = captureOutput(t, func() {
		err := runWorkspaceAdd(addCmd, []string{project})
		if err == nil {
			t.Fatal("expected duplicate workspace error")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestWorkspaceSwitch_UpdatesActiveWorkspace(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withWorkingDir(t, t.TempDir())

	projectOne := t.TempDir()
	projectTwo := t.TempDir()

	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")
	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{projectOne}); err != nil {
			t.Fatalf("runWorkspaceAdd projectOne: %v", err)
		}
		if err := runWorkspaceAdd(addCmd, []string{projectTwo}); err != nil {
			t.Fatalf("runWorkspaceAdd projectTwo: %v", err)
		}
	})

	switchCmd := &cobra.Command{}
	secondName := defaultNameFromPath(projectTwo)
	_, _ = captureOutput(t, func() {
		if err := runWorkspaceSwitch(switchCmd, []string{secondName}); err != nil {
			t.Fatalf("runWorkspaceSwitch: %v", err)
		}
	})

	cfg, err := sidecarpkg.LoadWorkspaceConfig()
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig: %v", err)
	}
	if cfg.Active != secondName {
		t.Fatalf("active workspace = %q, want %q", cfg.Active, secondName)
	}
}

func TestWorkspaceStatus_ReportsMissingActivePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withWorkingDir(t, t.TempDir())

	project := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")

	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{project}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

	missing := filepath.Join(home, "missing-workspace")
	if err := sidecarpkg.UpdateWorkspaceConfig(func(cfg *sidecarpkg.WorkspaceConfig) error {
		if len(cfg.Workspaces) == 0 {
			return fmt.Errorf("expected workspace to exist")
		}
		cfg.Workspaces[0].Path = missing
		cfg.Active = cfg.Workspaces[0].Name
		return nil
	}); err != nil {
		t.Fatalf("UpdateWorkspaceConfig: %v", err)
	}

	statusCmd := &cobra.Command{}
	output, _ := captureOutput(t, func() {
		if err := runWorkspaceStatus(statusCmd, nil); err != nil {
			t.Fatalf("runWorkspaceStatus: %v", err)
		}
	})

	if !strings.Contains(output, "path missing") {
		t.Fatalf("unexpected status output: %s", output)
	}
}

func TestWorkspaceList_UsesSidecarWorkspaceConfigStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withWorkingDir(t, t.TempDir())

	project := t.TempDir()
	addCmd := &cobra.Command{}
	addCmd.Flags().String("name", "", "")
	_, _ = captureOutput(t, func() {
		if err := runWorkspaceAdd(addCmd, []string{project}); err != nil {
			t.Fatalf("runWorkspaceAdd: %v", err)
		}
	})

	listCmd := &cobra.Command{}
	output, _ := captureOutput(t, func() {
		if err := runWorkspaceList(listCmd, nil); err != nil {
			t.Fatalf("runWorkspaceList: %v", err)
		}
	})
	name := defaultNameFromPath(project)
	if !strings.Contains(output, name) {
		t.Fatalf("expected workspace name %q in output: %s", name, output)
	}
}
