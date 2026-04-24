package generator

import (
	"fmt"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// DomainGenerator generates domain skill files.
// Ported from src/generators/domain.ts.
type DomainGenerator struct{}

func (g *DomainGenerator) Type() types.ArtifactType { return types.ArtifactTypeDomain }

func (g *DomainGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "description",
			Label:    "Domain skill description",
			Type:     "text",
			Required: false,
		},
	}
}

func (g *DomainGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
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
		description = fmt.Sprintf("Domain knowledge for %s work.", stringsToLower(title))
	}

	domainSlug := slug
	if domainSlug == "" {
		domainSlug = "new-domain"
	}

	content := fmt.Sprintf(`---
kind: domain-skill
name: %s
description: %s
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - %s
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
model_hint: sonnet
---

# %s Domain Skill

Use this skill to inject project-specific domain knowledge for %s tasks.

- surface domain constraints early
- prefer explicit contracts and invariants
- document risks, assumptions, and edge cases
`,
		domainSlug,
		description,
		domainSlug,
		title,
		stringsToLower(title),
	)

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf(".ai/orchestration/skills/domains/%s.md", domainSlug),
			Content: content,
		},
	}, nil
}

// stringsToLower is a simple ASCII lowercase helper.
func stringsToLower(s string) string {
	result := make([]byte, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = byte(r + 32)
		} else {
			result[i] = byte(r)
		}
	}
	return string(result)
}
