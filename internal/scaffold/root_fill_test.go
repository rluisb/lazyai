package scaffold

import (
	"strings"
	"testing"
)

// TestFillClaudeMdPlaceholders covers the hybrid auto-infer + fill-in marker
// substitution helper.
func TestFillClaudeMdPlaceholders(t *testing.T) {
	tmpl := `Project: [YOUR_PROJECT_NAME]
Description: [YOUR_PROJECT_DESCRIPTION]
Stack: [YOUR_TECH_STACK]
Org: [YOUR_ORG]
Team: [YOUR_TEAM]
Dev: [YOUR_DEV_COMMAND]
Test: [YOUR_TEST_COMMAND]
Build: [YOUR_BUILD_COMMAND]
Rule1: [YOUR_RULE_1]
Arch: [YOUR_ARCHITECTURE_NOTES]
Session: [YOUR_SESSION_CHECK]`

	t.Run("with known language and org/team filled", func(t *testing.T) {
		out := fillClaudeMdPlaceholders(tmpl, ScaffoldCompiledRootOptions{
			ProjectName:        "demo-app",
			ProjectDescription: "demo description",
			PrimaryLanguage:    "Go",
			Framework:          "Cobra",
			Organization:       "Acme",
			Team:               "Platform",
		})

		mustContain(t, out, "Project: demo-app")
		mustContain(t, out, "Description: demo description")
		mustContain(t, out, "Stack: Go · Cobra")
		mustContain(t, out, "Org: Acme")
		mustContain(t, out, "Team: Platform")
		mustContain(t, out, "Dev: go run .")
		mustContain(t, out, "Test: go test ./...")
		mustContain(t, out, "Build: go build ./...")
		mustContain(t, out, "Rule1: <!-- fill-in: rule 1 -->")
		mustContain(t, out, "Arch: <!-- fill-in: architecture and key patterns -->")
		mustContain(t, out, "Session: <!-- fill-in: team-specific session check -->")
	})

	t.Run("no language and no org/team uses fill-in markers", func(t *testing.T) {
		out := fillClaudeMdPlaceholders(tmpl, ScaffoldCompiledRootOptions{
			ProjectName: "demo-app",
		})

		mustContain(t, out, "Project: demo-app")
		mustContain(t, out, "Stack: <!-- fill-in: tech stack -->")
		mustContain(t, out, "Org: <!-- fill-in: your org -->")
		mustContain(t, out, "Team: <!-- fill-in: your team -->")
		mustContain(t, out, "Dev: <!-- fill-in: dev command -->")
		mustContain(t, out, "Description: AI-assisted development project")
	})

	t.Run("no YOUR_ placeholder remains for known keys", func(t *testing.T) {
		out := fillClaudeMdPlaceholders(tmpl, ScaffoldCompiledRootOptions{
			ProjectName:     "demo",
			PrimaryLanguage: "Go",
		})
		if strings.Contains(out, "[YOUR_") {
			t.Errorf("output still contains raw [YOUR_*] placeholder:\n%s", out)
		}
	})
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("output missing %q\n---\n%s", needle, haystack)
	}
}
