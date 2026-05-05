package library

import (
	"strings"
	"testing"
)

func TestSpeckitTasksDefinesAFKHITLMarkers(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "skills/speckit-tasks.md")

	required := []string{
		"[AFK]",
		"[HITL]",
		"AFK",
		"HITL",
		"autonomously",
		"human interaction",
	}
	assertContainsAll(t, "skills/speckit-tasks.md", content, required)
}

func TestOrchestrateDefinesCupcakeSignalMappingForDispatch(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "skills/orchestrate.md")

	required := []string{
		"plan_attested",
		"AFK",
		"HITL",
		"src/ writes",
		"plan_attested = true",
	}
	assertContainsAll(t, "skills/orchestrate.md", content, required)
}

func TestParallelExecutionDefinesAFKHITLWaveGating(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "skills/parallel-execution.md")

	required := []string{
		"AFK",
		"HITL",
		"AFK tasks",
		"HITL tasks",
		"wave",
	}
	assertContainsAll(t, "skills/parallel-execution.md", content, required)
}

func TestAFKHITLGuidanceDoesNotClaimNewEnforcement(t *testing.T) {
	t.Parallel()

	paths := []string{
		"skills/speckit-tasks.md",
		"skills/orchestrate.md",
		"skills/parallel-execution.md",
	}
	for _, path := range paths {
		content := strings.ToLower(readLibraryFile(t, path))
		forbidden := []string{
			"new enforcement layer",
			"replaces cupcake",
			"replaces pre-commit",
			"afk/hitl runtime enforcement",
			"afk classification engine",
			"hitl classification engine",
			"bypass cupcake",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Errorf("%s should not claim new enforcement with phrase %q", path, phrase)
			}
		}
	}
}
