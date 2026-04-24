package scaffold

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestScaffoldCompiledRoot_GlobalRequiresHomeDir verifies that passing an empty
// HomeDir at global scope returns an error rather than falling through to the
// real os.UserHomeDir() (R-3 mitigation from spec 008 risks).
func TestScaffoldCompiledRoot_GlobalRequiresHomeDir(t *testing.T) {
	err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:  t.TempDir(),
		HomeDir:    "", // intentionally empty
		SetupScope: types.SetupScopeGlobal,
		Tools:      []types.ToolId{types.ToolIdOpenCode},
	})
	if err == nil {
		t.Fatal("expected error for empty HomeDir at global scope, got nil")
	}
	if !strings.Contains(err.Error(), "HomeDir must be set") {
		t.Errorf("expected 'HomeDir must be set' message, got: %v", err)
	}
}

func TestMemoryDocDestPath(t *testing.T) {
	target := "/tmp/target"
	home := "/tmp/home"

	cases := []struct {
		name  string
		tool  types.ToolId
		scope types.SetupScope
		want  string
		unsup bool
	}{
		// opencode
		{"opencode_project", types.ToolIdOpenCode, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"opencode_workspace", types.ToolIdOpenCode, types.SetupScopeWorkspace, filepath.Join(target, "AGENTS.md"), false},
		{"opencode_global", types.ToolIdOpenCode, types.SetupScopeGlobal, filepath.Join(home, ".config", "opencode", "AGENTS.md"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			outputFile := RootFileByTool[c.tool]
			got, err := memoryDocDestPath(c.tool, c.scope, target, home, outputFile)
			if c.unsup {
				if !errors.Is(err, errMemoryDocScopeUnsupported) {
					t.Fatalf("want errMemoryDocScopeUnsupported, got err=%v path=%q", err, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
