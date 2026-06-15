package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandRemovesRetiredRuntimeSurfaces(t *testing.T) {
	for _, name := range []string{"task", "workflow", "orchestration", "mcp-setup"} {
		if hasRootSubcommand(name) {
			t.Fatalf("root command should not expose %q", name)
		}
	}
}

func TestCreateRejectsRetiredOrchestrationArtifactTypes(t *testing.T) {
	for _, artifactType := range []string{"workflow", "domain", "mode"} {
		t.Run(artifactType, func(t *testing.T) {
			withTempDir(t)
			cmd := newCreateTestCommand()
			err := runCreate(cmd, []string{artifactType, "demo"})
			want := fmt.Sprintf("invalid artifact type: %s (valid: agent, skill, prompt, command, template)", artifactType)
			if err == nil || err.Error() != want {
				t.Fatalf("runCreate error = %v, want %q", err, want)
			}
		})
	}
}

func TestClassifyPathTreatsAIConfigAsGenericConfig(t *testing.T) {
	if got := classifyPath(".ai/runtime/state.json"); got != "config" {
		t.Fatalf("classifyPath() = %q, want config", got)
	}
}

func hasRootSubcommand(name string) bool {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name {
			return true
		}
	}
	return false
}

func newCreateTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("no-interactive", true, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().String("description", "", "")
	return cmd
}
