package library

import (
	"strings"
	"testing"
)

func TestAgentStateTaxonomyDefinesLifecycleVocabulary(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "rules/agent-state.md")

	required := []string{
		"# Rule: Agent Lifecycle State Taxonomy",
		"loading_context",
		"planning",
		"awaiting_approval",
		"executing",
		"verifying",
		"blocked",
		"handoff",
		"done",
		"error",
		"## Lifecycle Vocabulary",
		"## Reporting Guidance",
	}

	assertContainsAll(t, "rules/agent-state.md", content, required)
}

func TestLifecycleGuidanceAppearsInStatusHandoffRecoverySurfaces(t *testing.T) {
	t.Parallel()

	paths := []string{
		"skills/orchestrate.md",
		"agents/orchestrator.md",
		"templates/task-harness-template.md",
	}
	required := []string{
		"Lifecycle label",
		"status reports",
		"handoff",
		"recovery summaries",
		"loading_context",
		"planning",
		"awaiting_approval",
		"executing",
		"verifying",
		"blocked",
		"handoff",
		"done",
		"error",
	}

	for _, path := range paths {
		content := readLibraryFile(t, path)
		assertContainsAll(t, path, content, required)
	}
}

func TestAgentStateGuidanceIsReportVocabularyOnly(t *testing.T) {
	t.Parallel()

	paths := []string{
		"rules/agent-state.md",
		"skills/orchestrate.md",
		"agents/orchestrator.md",
	}
	required := []string{
		"report vocabulary only",
		"does not add runtime per-agent state tracking",
		"does not imply runtime state-machine support",
		"ChainState",
		"StepState",
		"get_status",
	}

	for _, path := range paths {
		content := readLibraryFile(t, path)
		assertContainsAll(t, path, content, required)
	}
}

func TestAgentStateGuidanceDoesNotClaimRuntimeStateMachineSupport(t *testing.T) {
	t.Parallel()

	paths := []string{
		"rules/agent-state.md",
		"skills/orchestrate.md",
		"agents/orchestrator.md",
		"templates/task-harness-template.md",
	}
	for _, path := range paths {
		content := strings.ToLower(readLibraryFile(t, path))
		forbidden := []string{
			"runtime state tracking is enabled",
			"runtime lifecycle tracking is enabled",
			"persist lifecycle labels",
			"persisted lifecycle labels",
			"updates chainstate",
			"updates stepstate",
			"get_status includes lifecycle",
			"automatic state transitions",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Errorf("%s should not claim runtime agent state-machine support with phrase %q", path, phrase)
			}
		}
	}
}
