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

func TestTaskHarnessCarriesLifecycleVocabulary(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "templates/task-harness-template.md")
	required := []string{
		"Lifecycle label",
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
	assertContainsAll(t, "templates/task-harness-template.md", content, required)
}

func TestAgentStateGuidanceIsReportVocabularyOnly(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "rules/agent-state.md")
	required := []string{
		"report vocabulary only",
		"does not add runtime per-agent state tracking",
		"state-machine support",
		"host-tool status APIs",
	}
	assertContainsAll(t, "rules/agent-state.md", content, required)
}

func TestAgentStateGuidanceDoesNotClaimRuntimeStateMachineSupport(t *testing.T) {
	t.Parallel()

	paths := []string{
		"rules/agent-state.md",
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
