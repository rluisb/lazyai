package library

import (
	"io/fs"
	"strings"
	"testing"
)

func TestBugfixRCATemplateIncludesCausalReasoningFields(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "templates/bugfix-rca-template.md")

	required := []string{
		"## Causal Method",
		"5-Whys",
		"causal chain",
		"Proximate cause",
		"Contributing factors",
		"Root cause",
		"Missing guardrail/test/standard",
		"Evidence",
		"Confidence",
		"Counterfactual check",
	}

	assertContainsAll(t, "templates/bugfix-rca-template.md", content, required)
}

func TestBugfixSkillRequiresCausalAnalysisBeforePlanning(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "skills/bugfix.md")

	required := []string{
		"For non-trivial bugs, complete causal analysis before fix planning",
		"proximate cause",
		"contributing factors",
		"root cause",
		"missing guardrail/test/standard",
		"evidence",
		"confidence",
		"counterfactual",
	}

	assertContainsAll(t, "skills/bugfix.md", content, required)
}

func TestBugfixRCAGuidanceDoesNotIntroduceDeferredInfrastructure(t *testing.T) {
	t.Parallel()

	paths := []string{
		"templates/bugfix-rca-template.md",
		"skills/bugfix.md",
	}
	for _, path := range paths {
		content := strings.ToLower(readLibraryFile(t, path))
		forbidden := []string{
			"external evaluation infrastructure",
			"learning loop",
			"runtime state tracking",
			"auto-recovery engine",
			"rag pipeline",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Errorf("%s should not require deferred infrastructure phrase %q", path, phrase)
			}
		}
	}
}

func readLibraryFile(t *testing.T, path string) string {
	t.Helper()

	content, err := fs.ReadFile(GetLibraryFS(), path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertContainsAll(t *testing.T, path, content string, required []string) {
	t.Helper()

	for _, phrase := range required {
		if !strings.Contains(content, phrase) {
			t.Errorf("%s missing required RCA guidance %q", path, phrase)
		}
	}
}
