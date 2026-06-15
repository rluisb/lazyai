package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestAdapterNeutralContract_PrimaryAgent(t *testing.T) {
	cases := []struct {
		name             string
		adapter          ToolAdapter
		primaryAgentPath string
		orchestratorPath string
		retiredSkillPath string
		assertConfig     func(t *testing.T, targetDir string)
	}{
		{
			name:             "opencode",
			adapter:          &OpenCodeAdapter{},
			primaryAgentPath: filepath.Join(".opencode", "agents", "primary-agent.md"),
			orchestratorPath: filepath.Join(".opencode", "agents", "orchestrator.md"),
			retiredSkillPath: filepath.Join(".opencode", "skills", "orchestrate", "SKILL.md"),
			assertConfig: func(t *testing.T, targetDir string) {
				t.Helper()
				cfg, err := jsonc.ReadJSONCFile(filepath.Join(targetDir, ".opencode", OpenCodeConfigFilename))
				if err != nil {
					t.Fatalf("read opencode config: %v", err)
				}
				if got, _ := cfg["default_agent"].(string); got != "primary-agent" {
					t.Fatalf("default_agent = %q, want primary-agent", got)
				}
				if instructions, ok := cfg["instructions"].([]any); ok {
					for _, raw := range instructions {
						if s, _ := raw.(string); s == "STARTUP.md" {
							t.Fatalf("neutral OpenCode config must not include STARTUP.md: %v", instructions)
						}
					}
				}
				assertMissing(t, filepath.Join(targetDir, ".opencode", "STARTUP.md"))
			},
		},
		{
			name:             "claude-code",
			adapter:          &ClaudeCodeAdapter{},
			primaryAgentPath: filepath.Join(".claude", "agents", "primary-agent.md"),
			orchestratorPath: filepath.Join(".claude", "agents", "orchestrator.md"),
			retiredSkillPath: filepath.Join(".claude", "skills", "orchestrate", "SKILL.md"),
		},
		{
			name:             "copilot",
			adapter:          &CopilotAdapter{},
			primaryAgentPath: filepath.Join(".github", "agents", "primary-agent.agent.yaml"),
			orchestratorPath: filepath.Join(".github", "agents", "orchestrator.agent.yaml"),
			retiredSkillPath: filepath.Join(".github", "agents", "orchestrate.agent.yaml"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, targetDir := newNeutralContractContext(t)
			if _, err := tc.adapter.Install(ctx); err != nil {
				t.Fatalf("install: %v", err)
			}
			assertExists(t, filepath.Join(targetDir, tc.primaryAgentPath))
			assertMissing(t, filepath.Join(targetDir, tc.orchestratorPath))
			if tc.assertConfig != nil {
				tc.assertConfig(t, targetDir)
			}
			assertMissing(t, filepath.Join(targetDir, tc.retiredSkillPath))
		})
	}
}

func newNeutralContractContext(t *testing.T) (*AdapterContext, string) {
	t.Helper()
	targetDir := t.TempDir()
	return &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  neutralContractFS(),
		Strategy:   types.ConflictStrategyAlign,
	}, targetDir
}

func neutralContractFS() fstest.MapFS {
	return fstest.MapFS{
		"canonical/agents/primary-agent.md":               &fstest.MapFile{Data: []byte("# Primary Agent\n\nDefault LazyAI runtime entry point.\n")},
		"canonical/agents/builder.md":                     &fstest.MapFile{Data: canonicalAgentFixture("builder", "Builder agent.")},
		"canonical/agents/planner.md":                     &fstest.MapFile{Data: canonicalAgentFixture("planner", "Planner agent.")},
		"canonical/agents/reviewer.md":                    &fstest.MapFile{Data: canonicalAgentFixture("reviewer", "Reviewer agent.")},
		"canonical/agents/scout.md":                       &fstest.MapFile{Data: canonicalAgentFixture("scout", "Scout agent.")},
		"canonical/agents/evidence-verifier.md":           &fstest.MapFile{Data: canonicalAgentFixture("evidence-verifier", "Evidence verifier agent.")},
		"skills/codebase-exploration.md":                  &fstest.MapFile{Data: canonicalSkillFixture("codebase-exploration", "Codebase exploration skill.")},
		"skills/test-first-change.md":                     &fstest.MapFile{Data: canonicalSkillFixture("test-first-change", "Test first change skill.")},
		"skills/diagnose.md":                              &fstest.MapFile{Data: canonicalSkillFixture("diagnose", "Diagnose skill.")},
		"skills/issue-triage.md":                          &fstest.MapFile{Data: canonicalSkillFixture("issue-triage", "Issue triage skill.")},
		"rules/typescript.md":                             &fstest.MapFile{Data: []byte("---\npaths:\n  - \"src/**/*.ts\"\n---\n\n# TypeScript Rules\n")},
		"copilot/instructions/repository.instructions.md": &fstest.MapFile{Data: []byte("# Copilot Instructions\n")},
	}
}

func canonicalAgentFixture(name, description string) []byte {
	return []byte("---\nname: " + name + "\ndescription: " + description + "\ntier: balanced\ntemperature: 0.1\nthinking: low\nrisk: 3\n---\n\n# " + name + "\n\nAgent body.\n")
}

func canonicalSkillFixture(name, description string) []byte {
	return []byte("---\nname: " + name + "\ndescription: " + description + "\ntier: balanced\nthinking: low\nrisk: 3\n---\n\n# " + name + "\n\nSkill body.\n")
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		t.Fatalf("expected %s to be absent, found contents prefix %q", path, strings.TrimSpace(string(data[:min(len(data), 120)])))
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}
