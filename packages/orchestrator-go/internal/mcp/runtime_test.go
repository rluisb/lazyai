package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

func TestLoadRuntimeConfigDefaultsToNativeWithoutConfig(t *testing.T) {
	t.Setenv("AI_SETUP_ORCHESTRATOR_CONFIG", "")
	runtime, err := LoadRuntimeConfig(RuntimeConfigOptions{ProjectRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("LoadRuntimeConfig: %v", err)
	}
	if runtime.ExecutionMode != types.ExecutionNative || runtime.ConfigPath != "" || runtime.A2AConfig != nil {
		t.Fatalf("expected default native without config, got %+v", runtime)
	}
}

func TestLoadRuntimeConfigUsesProjectConfig(t *testing.T) {
	t.Setenv("AI_SETUP_ORCHESTRATOR_CONFIG", "")
	root := t.TempDir()
	writeRuntimeConfig(t, filepath.Join(root, ".ai", "orchestrator.json"), "a2a")

	runtime, err := LoadRuntimeConfig(RuntimeConfigOptions{ProjectRoot: root})
	if err != nil {
		t.Fatalf("LoadRuntimeConfig: %v", err)
	}
	if runtime.ExecutionMode != types.ExecutionA2A || runtime.A2AConfig == nil {
		t.Fatalf("expected project config a2a runtime, got %+v", runtime)
	}
	if runtime.ConfigPath != filepath.Join(root, ".ai", "orchestrator.json") {
		t.Fatalf("unexpected config path %q", runtime.ConfigPath)
	}
}

func TestLoadRuntimeConfigEnvPathOverridesProjectConfig(t *testing.T) {
	root := t.TempDir()
	writeRuntimeConfig(t, filepath.Join(root, ".ai", "orchestrator.json"), "native")
	envPath := filepath.Join(t.TempDir(), "orchestrator.json")
	writeRuntimeConfig(t, envPath, "hybrid")
	t.Setenv("AI_SETUP_ORCHESTRATOR_CONFIG", envPath)

	runtime, err := LoadRuntimeConfig(RuntimeConfigOptions{ProjectRoot: root})
	if err != nil {
		t.Fatalf("LoadRuntimeConfig: %v", err)
	}
	if runtime.ExecutionMode != types.ExecutionHybrid || runtime.ConfigPath != envPath {
		t.Fatalf("expected env config hybrid runtime, got %+v", runtime)
	}
}

func TestLoadRuntimeConfigExplicitConfigAndModeOverride(t *testing.T) {
	t.Setenv("AI_SETUP_ORCHESTRATOR_CONFIG", "")
	root := t.TempDir()
	writeRuntimeConfig(t, filepath.Join(root, ".ai", "orchestrator.json"), "native")
	explicitPath := filepath.Join(t.TempDir(), "orchestrator.json")
	writeRuntimeConfig(t, explicitPath, "a2a")

	runtime, err := LoadRuntimeConfig(RuntimeConfigOptions{
		ProjectRoot:           root,
		ConfigPath:            explicitPath,
		ConfigPathExplicit:    true,
		ExecutionMode:         "hybrid",
		ExecutionModeExplicit: true,
	})
	if err != nil {
		t.Fatalf("LoadRuntimeConfig: %v", err)
	}
	if runtime.ConfigPath != explicitPath || runtime.ExecutionMode != types.ExecutionHybrid {
		t.Fatalf("expected explicit path with CLI mode override, got %+v", runtime)
	}
}

func TestLoadRuntimeConfigDefaultExecutionFlagDoesNotOverrideConfig(t *testing.T) {
	t.Setenv("AI_SETUP_ORCHESTRATOR_CONFIG", "")
	root := t.TempDir()
	writeRuntimeConfig(t, filepath.Join(root, ".ai", "orchestrator.json"), "a2a")

	runtime, err := LoadRuntimeConfig(RuntimeConfigOptions{ProjectRoot: root, ExecutionMode: "native"})
	if err != nil {
		t.Fatalf("LoadRuntimeConfig: %v", err)
	}
	if runtime.ExecutionMode != types.ExecutionA2A {
		t.Fatalf("default CLI value should not override config mode, got %q", runtime.ExecutionMode)
	}
}

func writeRuntimeConfig(t *testing.T, path, mode string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	content := `{
  "version":1,
  "execution":{"mode":"` + mode + `"},
  "providers":{"local":{"endpoint":"https://a2a.example.test/rpc","auth":{"type":"none"}}},
  "agents":{"builder":{"provider":"local","enabled":true,"tools":["opencode"]}}
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}
