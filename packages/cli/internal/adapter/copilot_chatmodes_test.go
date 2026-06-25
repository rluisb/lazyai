package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestCopilotAdapter_InstallsCustomAgents verifies Copilot installs custom
// agent files at .github/agents/<name>.agent.md (project/workspace scope) and
// that the deprecated .github/chatmodes/ location is NOT used.
func TestCopilotAdapter_InstallsCustomAgents(t *testing.T) {
	cases := []struct {
		name  string
		scope types.SetupScope
	}{
		{"project", types.SetupScopeProject},
		{"workspace", types.SetupScopeWorkspace},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ctx, target, _ := newScopeTestContext(t, c.scope)
			ctx.Selections.ChatModes = []types.ChatModeId{types.ChatModeIdArchitect, types.ChatModeIdReviewer}

			adapter := &CopilotAdapter{}
			if _, err := adapter.Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			agentsDir := filepath.Join(target, ".github", "agents")

			// New location: .github/agents/<name>.agent.md
			for _, name := range []string{"architect", "reviewer"} {
				agentFile := filepath.Join(agentsDir, name+".agent.md")
				if !files.FileExists(agentFile) {
					t.Errorf("expected %s.agent.md at %q", name, agentsDir)
					continue
				}
				content, err := os.ReadFile(agentFile)
				if err != nil {
					t.Errorf("read %s.agent.md: %v", name, err)
					continue
				}
				s := string(content)
				if !strings.Contains(s, "description:") {
					t.Errorf("%s.agent.md missing 'description:' frontmatter key", name)
				}
				if !strings.Contains(s, "tools:") {
					t.Errorf("%s.agent.md missing 'tools:' frontmatter key", name)
				}
			}

			// Clean cutover: deprecated .chatmode.md files must NOT be emitted.
			chatmodesDir := filepath.Join(target, ".github", "chatmodes")
			for _, name := range []string{"architect", "reviewer"} {
				deprecated := filepath.Join(chatmodesDir, name+".chatmode.md")
				if files.FileExists(deprecated) {
					t.Errorf("deprecated %s.chatmode.md was emitted at %q — must not be", name, chatmodesDir)
				}
			}
		})
	}
}

// TestCopilotAdapter_CustomAgents_GlobalSkips verifies that global scope (not
// supported for Copilot) is a clean no-op — no custom agent files written.
func TestCopilotAdapter_CustomAgents_GlobalSkips(t *testing.T) {
	ctx, target, home := newScopeTestContext(t, types.SetupScopeGlobal)
	ctx.Selections.ChatModes = []types.ChatModeId{types.ChatModeIdArchitect}

	adapter := &CopilotAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install at global must not error: %v", err)
	}

	// No .agent.md files in project target agents dir.
	agentsDir := filepath.Join(target, ".github", "agents")
	for _, name := range []string{"architect", "reviewer"} {
		if files.FileExists(filepath.Join(agentsDir, name+".agent.md")) {
			t.Errorf("global scope must not create %s.agent.md under project", name)
		}
	}

	// No .agent.md files in home agents dir either.
	homeAgentsDir := filepath.Join(home, ".github", "agents")
	for _, name := range []string{"architect", "reviewer"} {
		if files.FileExists(filepath.Join(homeAgentsDir, name+".agent.md")) {
			t.Errorf("global scope must not create %s.agent.md under home", name)
		}
	}
}
