package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func makeFakeBin(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	fakeBin := filepath.Join(binDir, "opencode")
	if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("writing fake binary: %v", err)
	}
	return binDir
}

func TestInstallOpenCodePlugins_BinaryAbsent(t *testing.T) {
	t.Setenv("PATH", "")

	called := false
	stub := func(name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
		Selections: AdapterSelections{
			OpenCodePlugins: []string{"@opencode/desktop-commander"},
		},
	}

	if err := installOpenCodePlugins(ctx, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("runner should not be called when binary is absent")
	}
}

func TestInstallOpenCodePlugins_EmptyList(t *testing.T) {
	called := false
	stub := func(name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
		Selections: AdapterSelections{
			OpenCodePlugins: nil,
		},
	}

	if err := installOpenCodePlugins(ctx, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("runner should not be called when plugin list is empty")
	}
}

func TestInstallOpenCodePlugins_GlobalScopePassesFlag(t *testing.T) {
	t.Setenv("PATH", makeFakeBin(t))

	var capturedArgs []string
	stub := func(name string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte("ok"), nil
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeGlobal,
		Selections: AdapterSelections{
			OpenCodePlugins: []string{"@opencode/git-tools"},
		},
	}

	if err := installOpenCodePlugins(ctx, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedArgs) == 0 {
		t.Fatal("runner was not called")
	}
	found := false
	for _, a := range capturedArgs {
		if a == "-g" {
			found = true
		}
	}
	if !found {
		t.Errorf("global scope must pass -g; got args %v", capturedArgs)
	}
}

func TestInstallOpenCodePlugins_ProjectScopeNoFlag(t *testing.T) {
	t.Setenv("PATH", makeFakeBin(t))

	var capturedArgs []string
	stub := func(name string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte("ok"), nil
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
		Selections: AdapterSelections{
			OpenCodePlugins: []string{"@opencode/git-tools"},
		},
	}

	if err := installOpenCodePlugins(ctx, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, a := range capturedArgs {
		if a == "-g" {
			t.Errorf("project scope must NOT pass -g; got args %v", capturedArgs)
		}
	}
}

func TestInstallOpenCodePlugins_FailureIsNonFatal(t *testing.T) {
	t.Setenv("PATH", makeFakeBin(t))

	stub := func(name string, args ...string) ([]byte, error) {
		return []byte("install failed"), fmt.Errorf("exit status 1")
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
		Selections: AdapterSelections{
			OpenCodePlugins: []string{"@opencode/desktop-commander"},
		},
	}

	// installOpenCodePlugins itself returns nil — failures are logged only.
	if err := installOpenCodePlugins(ctx, stub); err != nil {
		t.Errorf("plugin failure must not propagate as error, got: %v", err)
	}
}

func TestInstallOpenCodePlugins_MultiplePlugins(t *testing.T) {
	t.Setenv("PATH", makeFakeBin(t))

	var invocations []string
	stub := func(name string, args ...string) ([]byte, error) {
		invocations = append(invocations, strings.Join(args, " "))
		return []byte("ok"), nil
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
		Selections: AdapterSelections{
			OpenCodePlugins: []string{"@opencode/git-tools", "@opencode/context-files"},
		},
	}

	if err := installOpenCodePlugins(ctx, stub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(invocations) != 2 {
		t.Errorf("expected 2 invocations, got %d: %v", len(invocations), invocations)
	}
}
