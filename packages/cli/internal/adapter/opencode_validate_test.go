package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func makeValidateCtx(t *testing.T) *AdapterContext {
	t.Helper()
	return &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
		LibraryFS:  createTestFS(),
		Strategy:   types.ConflictStrategyAlign,
	}
}

func makeFakeOpenCodeBin(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	fakeBin := filepath.Join(binDir, "opencode")
	content := []byte("#!/bin/sh\nexit 0\n")
	if runtime.GOOS == "windows" {
		fakeBin = filepath.Join(binDir, "opencode.bat")
		content = []byte("@echo off\r\nexit /b 0\r\n")
	}
	if err := os.WriteFile(fakeBin, content, 0o755); err != nil {
		t.Fatalf("writing fake binary: %v", err)
	}
	return binDir
}

func TestValidateOpenCodeInstall_BinaryAbsent(t *testing.T) {
	ctx := makeValidateCtx(t)

	// Stub runner that always fails — but since LookPath will fail for a
	// nonexistent binary we never reach the runner.
	stubRunner := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("unexpected call")
	}

	// Temporarily override PATH to guarantee the binary is absent.
	t.Setenv("PATH", "")

	warnings, err := validateOpenCodeInstall(ctx, stubRunner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings when binary absent, got %v", warnings)
	}
}

func TestValidateOpenCodeInstall_ConfigFails(t *testing.T) {
	ctx := makeValidateCtx(t)

	t.Setenv("PATH", makeFakeOpenCodeBin(t))

	callCount := 0
	stubRunner := func(name string, args ...string) ([]byte, error) {
		callCount++
		if len(args) >= 2 && args[0] == "debug" && args[1] == "config" {
			return nil, fmt.Errorf("config parse error")
		}
		return []byte("ok"), nil
	}

	warnings, err := validateOpenCodeInstall(ctx, stubRunner)
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}

	found := false
	for _, w := range warnings {
		if w.Scope == "config" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a config warning; got %v", warnings)
	}
}

func TestValidateOpenCodeInstall_AgentFails(t *testing.T) {
	ctx := makeValidateCtx(t)

	// Write a fake agent file so the glob finds it.
	ocDir := filepath.Join(ctx.TargetDir, ".opencode")
	agentsDir := filepath.Join(ocDir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "implementer.md"), []byte("# Implementer"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	t.Setenv("PATH", makeFakeOpenCodeBin(t))

	stubRunner := func(name string, args ...string) ([]byte, error) {
		if len(args) >= 3 && args[0] == "debug" && args[1] == "agent" {
			return nil, fmt.Errorf("parse error in frontmatter")
		}
		// config passes but returns empty (no "mcp" substring)
		return []byte(`{"mcp": {}}`), nil
	}

	warnings, err := validateOpenCodeInstall(ctx, stubRunner)
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}

	found := false
	for _, w := range warnings {
		if w.Scope == "agent" && w.Item == "implementer" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected agent warning for implementer; got %v", warnings)
	}
}

func TestValidateOpenCodeInstall_AllPass(t *testing.T) {
	ctx := makeValidateCtx(t)

	ocDir := filepath.Join(ctx.TargetDir, ".opencode")
	agentsDir := filepath.Join(ocDir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "implementer.md"), []byte("# Implementer"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	t.Setenv("PATH", makeFakeOpenCodeBin(t))

	stubRunner := func(name string, args ...string) ([]byte, error) {
		return []byte(`{"mcp": {"servers": {}}}`), nil
	}

	warnings, err := validateOpenCodeInstall(ctx, stubRunner)
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings on all-pass, got %v", warnings)
	}
}
