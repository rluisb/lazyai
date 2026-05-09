package adapter

import (
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/models"
)

// TestSkillSpecOrDefault_NoFrontmatter verifies that a skill without any
// frontmatter falls through to the Balanced default. This is the
// "zero-migration" path: existing skills (which all lack tier metadata)
// continue to work without source changes.
func TestSkillSpecOrDefault_NoFrontmatter(t *testing.T) {
	body := []byte("# Just a body\n\nNo frontmatter at all.")
	spec := skillSpecOrDefault(body, "review")

	if spec.Tier != models.TierBalanced {
		t.Errorf("default tier: want %q, got %q", models.TierBalanced, spec.Tier)
	}
	if spec.Name != "review" {
		t.Errorf("name: want %q, got %q", "review", spec.Name)
	}
	if spec.Risk != 3 {
		t.Errorf("default risk: want %d, got %d", 3, spec.Risk)
	}
	if spec.Thinking != models.ThinkingLow {
		t.Errorf("default thinking: want %q, got %q", models.ThinkingLow, spec.Thinking)
	}
}

// TestSkillSpecOrDefault_NoTierAnnotation verifies that a skill with
// frontmatter but without `tier:` (the common case for current skills)
// also falls through to the Balanced default. `frontmatter.ParseAgentSpec`
// errors when tier is empty; the helper swallows that error.
func TestSkillSpecOrDefault_NoTierAnnotation(t *testing.T) {
	src := []byte(`---
name: bugfix
description: "Fix a bug"
trigger: /bugfix
phase: implementation
---

Body.`)
	spec := skillSpecOrDefault(src, "bugfix")
	if spec.Tier != models.TierBalanced {
		t.Errorf("tier: want %q (default), got %q", models.TierBalanced, spec.Tier)
	}
}

// TestSkillSpecOrDefault_WithTierFrontier verifies that a skill that
// declares `tier: frontier` is parsed correctly — this is the opt-in
// override path for skills genuinely needing Opus-class reasoning.
func TestSkillSpecOrDefault_WithTierFrontier(t *testing.T) {
	src := []byte(`---
name: red-team-plan
tier: frontier
risk: 5
temperature: 0.7
thinking: high
description: Adversarial review.
---

Body.`)
	spec := skillSpecOrDefault(src, "red-team-plan")
	if spec.Tier != models.TierFrontier {
		t.Errorf("tier: want %q, got %q", models.TierFrontier, spec.Tier)
	}
	if spec.Risk != 5 {
		t.Errorf("risk: want %d, got %d", 5, spec.Risk)
	}
	if spec.Thinking != models.ThinkingHigh {
		t.Errorf("thinking: want %q, got %q", models.ThinkingHigh, spec.Thinking)
	}
}

// TestStripModelSection verifies that the `## Model\n<paragraph>\n\n`
// editorial section is stripped from a markdown document. This is the
// fix for #199 Bug 1: OpenCode/Copilot deployments shouldn't carry
// Claude-centric model commentary that contradicts the resolved
// provider/model pair.
func TestStripModelSection(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantOut  string
		wantNoTo string // if non-empty, asserts the substring is absent
	}{
		{
			name: "strips Model section between other H2 sections",
			input: `# Builder

## Identity
You build things.

## Model
Sonnet or equivalent fast model. Building is coordination.

## Tone
Verbose.`,
			wantNoTo: "## Model",
		},
		{
			name: "strips Model section at end of doc",
			input: `# Builder

Body content.

## Model
Sonnet or equivalent fast model.
`,
			wantNoTo: "## Model",
		},
		{
			name: "no-op when section absent",
			input: `# Builder

Body without a Model section.
`,
			wantOut: `# Builder

Body without a Model section.
`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := string(stripModelSection([]byte(c.input)))
			if c.wantOut != "" && out != c.wantOut {
				t.Errorf("strip output:\n--- want ---\n%s\n--- got ---\n%s", c.wantOut, out)
			}
			if c.wantNoTo != "" && strings.Contains(out, c.wantNoTo) {
				t.Errorf("output still contains %q:\n%s", c.wantNoTo, out)
			}
		})
	}
}

// TestOpencodeReasoningEffortMapping verifies the source `thinking:` field
// is translated to OpenCode's `reasoningEffort` enum.
func TestOpencodeReasoningEffortMapping(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"minimal", "minimal"},
		{"none", ""},
		{"", ""},
		{"BoGuS", ""},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			if got := opencodeReasoningEffort(c.input); got != c.want {
				t.Errorf("input %q: want %q, got %q", c.input, c.want, got)
			}
		})
	}
}

// TestOpencodeTextVerbosityFromRisk verifies the risk → textVerbosity
// derivation: high-risk roles (planning, review) prefer terse output;
// lower-risk roles get medium.
func TestOpencodeTextVerbosityFromRisk(t *testing.T) {
	cases := []struct {
		risk int
		want string
	}{
		{1, "medium"},
		{2, "medium"},
		{3, "medium"},
		{4, "low"},
		{5, "low"},
	}
	for _, c := range cases {
		if got := opencodeTextVerbosity(c.risk); got != c.want {
			t.Errorf("risk %d: want %q, got %q", c.risk, c.want, got)
		}
	}
}

// TestOpencodeStepsForTier verifies per-tier max-iteration caps mirror the
// canonical configs at `~/.config/opencode/agents/`.
func TestOpencodeStepsForTier(t *testing.T) {
	cases := map[string]int{
		"frontier": 16,
		"balanced": 20,
		"speed":    10,
		"":         0,
		"invalid":  0,
	}
	for tier, want := range cases {
		if got := opencodeStepsFor(tier); got != want {
			t.Errorf("tier %q: want %d, got %d", tier, want, got)
		}
	}
}
