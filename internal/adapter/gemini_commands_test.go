package adapter

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestGeminiAdapter_InstallsCommands verifies Gemini installs *.toml custom
// commands from library/commands/ at every supported scope.
func TestGeminiAdapter_InstallsCommands(t *testing.T) {
	cases := []struct {
		name  string
		scope types.SetupScope
	}{
		{"project", types.SetupScopeProject},
		{"workspace", types.SetupScopeWorkspace},
		{"global", types.SetupScopeGlobal},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ctx, _, _ := newScopeTestContext(t, c.scope)
			ctx.Selections.Commands = []types.CommandId{types.CommandIdRpi, types.CommandIdReview}

			adapter := &GeminiAdapter{}
			if _, err := adapter.Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			root, err := ResolveToolRoot(types.ToolIdGemini, c.scope, ctx)
			if err != nil {
				t.Fatalf("ResolveToolRoot: %v", err)
			}
			commandsDir := filepath.Join(root, "commands")

			// Both selected commands must be present.
			if !files.FileExists(filepath.Join(commandsDir, "rpi.toml")) {
				t.Errorf("rpi.toml not installed at %q", commandsDir)
			}
			if !files.FileExists(filepath.Join(commandsDir, "review.toml")) {
				t.Errorf("review.toml not installed at %q", commandsDir)
			}
			// Unselected plan.toml must NOT be installed.
			if files.FileExists(filepath.Join(commandsDir, "plan.toml")) {
				t.Errorf("plan.toml should not be installed (not in Selections.Commands)")
			}
		})
	}
}
