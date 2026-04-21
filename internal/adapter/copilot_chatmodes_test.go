package adapter

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestCopilotAdapter_InstallsChatModes verifies Copilot installs
// *.chatmode.md files at project/workspace scope.
func TestCopilotAdapter_InstallsChatModes(t *testing.T) {
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

			chatmodesDir := filepath.Join(target, ".github", "chatmodes")
			if !files.FileExists(filepath.Join(chatmodesDir, "architect.chatmode.md")) {
				t.Errorf("architect.chatmode.md not installed at %q", chatmodesDir)
			}
			if !files.FileExists(filepath.Join(chatmodesDir, "reviewer.chatmode.md")) {
				t.Errorf("reviewer.chatmode.md not installed at %q", chatmodesDir)
			}
		})
	}
}

// TestCopilotAdapter_ChatModes_GlobalSkips verifies that global scope (already
// unsupported for Copilot) is a clean no-op — no chatmodes written.
func TestCopilotAdapter_ChatModes_GlobalSkips(t *testing.T) {
	ctx, target, home := newScopeTestContext(t, types.SetupScopeGlobal)
	ctx.Selections.ChatModes = []types.ChatModeId{types.ChatModeIdArchitect}

	adapter := &CopilotAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install at global must not error: %v", err)
	}
	if files.DirExists(filepath.Join(target, ".github", "chatmodes")) {
		t.Error("global scope must not create .github/chatmodes under project")
	}
	if files.DirExists(filepath.Join(home, ".github", "chatmodes")) {
		t.Error("global scope must not create ~/.github/chatmodes")
	}
}
