package generator

import (
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ModeGenerator generates mode skill files.
// Ported from src/generators/mode.ts.
type ModeGenerator struct{}

func (g *ModeGenerator) Type() types.ArtifactType { return types.ArtifactTypeMode }

func (g *ModeGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "description",
			Label:    "Mode skill description",
			Type:     "text",
			Required: false,
		},
	}
}

func (g *ModeGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	title := ToTitleCase(slug)
	if title == "" {
		title = config.Name
	}

	description := getAnswer(config.Answers, "description", "")
	if description == "" {
		description = config.Description
	}
	if description == "" {
		description = fmt.Sprintf("Behavioral operating mode for %s execution.", stringsToLower(title))
	}

	modeSlug := slug
	if modeSlug == "" {
		modeSlug = "new-mode"
	}

	content := fmt.Sprintf(`---
kind: mode-skill
name: %s
description: %s
behavior:
  - keep work aligned to the active plan
  - surface risks before high-cost execution
  - prefer deterministic, auditable steps
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
approval_policy: normal
model_hint: opus
---

# %s Mode Skill

Use this skill to modify how the active agent behaves while preserving the base role.

- define when to ask for confirmation
- clarify autonomy expectations
- document trade-offs and stopping conditions
`,
		modeSlug,
		description,
		title,
	)

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf(".ai/orchestration/skills/modes/%s.md", modeSlug),
			Content: content,
		},
	}, nil
}
