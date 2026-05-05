package library

import (
	"strings"
	"testing"
)

func TestBugfixSkillRequiresFeedbackLoopBeforeHypothesis(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "skills/bugfix.md")

	assertContainsAll(t, "skills/bugfix.md", content, []string{
		"Step 0",
		"automated pass/fail signal",
		"feedback loop",
	})

	if !strings.Contains(content, "before hypothesizing") && !strings.Contains(content, "before hypothesis") {
		t.Errorf("skills/bugfix.md missing required feedback-loop timing %q or %q", "before hypothesizing", "before hypothesis")
	}
}

func TestBugfixRCATemplateIncludesFeedbackLoopFields(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "templates/bugfix-rca-template.md")

	assertContainsAll(t, "templates/bugfix-rca-template.md", content, []string{
		"Feedback Loop",
		"Signal command/source",
		"Expected failing behavior",
		"Expected passing behavior",
		"Current result",
	})
}

func TestBugfixFeedbackLoopGuidanceDoesNotRequireExternalRuntimeInfrastructure(t *testing.T) {
	t.Parallel()

	paths := []string{
		"skills/bugfix.md",
		"templates/bugfix-rca-template.md",
	}
	for _, path := range paths {
		content := strings.ToLower(readLibraryFile(t, path))
		forbidden := []string{
			"external evaluation infrastructure",
			"runtime engine",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Errorf("%s should not require feedback-loop infrastructure phrase %q", path, phrase)
			}
		}
	}
}
