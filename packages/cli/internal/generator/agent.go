package generator

import (
	"fmt"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// AgentGenerator generates agent markdown files.
// Ported from src/generators/agent.ts.
type AgentGenerator struct{}

func (g *AgentGenerator) Type() types.ArtifactType { return types.ArtifactTypeAgent }

func (g *AgentGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "model",
			Label:    "Model",
			Type:     "select",
			Required: true,
			Default:  "claude-sonnet-4-5",
			Options: []PromptOption{
				{Value: "claude-sonnet-4-5", Label: "claude-sonnet-4-5 (recommended for most agents)"},
				{Value: "claude-opus-4-5", Label: "claude-opus-4-5 (deep reasoning: planner, reviewer, red-team)"},
				{Value: "gpt-4o", Label: "gpt-4o"},
				{Value: "gemini-pro", Label: "gemini-pro"},
				{Value: "other", Label: "other"},
			},
		},
		{
			Key:      "mode",
			Label:    "Mode",
			Type:     "select",
			Required: true,
			Default:  "interactive",
			Options: []PromptOption{
				{Value: "autonomous", Label: "autonomous"},
				{Value: "interactive", Label: "interactive"},
				{Value: "hybrid", Label: "hybrid"},
			},
		},
		{
			Key:     "tools",
			Label:   "Tools (comma-separated)",
			Type:    "text",
			Default: "",
		},
	}
}

func (g *AgentGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	title := ToTitleCase(slug)
	if title == "" {
		title = config.Name
	}

	model := getAnswer(config.Answers, "model", "claude-sonnet-4-5")
	mode := getAnswer(config.Answers, "mode", "interactive")
	toolsRaw := getAnswer(config.Answers, "tools", "")

	var tools []string
	for _, item := range strings.Split(toolsRaw, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			tools = append(tools, item)
		}
	}

	var capabilityLines []string
	if len(tools) > 0 {
		for _, tool := range tools {
			capabilityLines = append(capabilityLines, fmt.Sprintf("- Use %s effectively when needed", tool))
		}
	} else {
		capabilityLines = []string{"- Execute tasks aligned to this role"}
	}

	description := config.Description
	if description == "" {
		description = "high-quality task execution"
	}

	content := fmt.Sprintf(`---
name: %s
model: %s
mode: %s
---

# %s Agent

## Model
Recommended: %s

## Identity

You are %s — an AI specialist focused on %s.

## Capability

%s

## Rules

1. Understand scope and constraints before acting.
2. Prefer minimal, verifiable changes.
3. Preserve established project patterns.
4. Communicate assumptions and risks clearly.
5. Validate outputs before handoff.

## Reasoning Protocol

Before execution:
1. Identify objective and acceptance criteria.
2. Determine the smallest safe action.
3. Execute with evidence-driven checks.
4. Confirm outcomes and side effects.

## Trace Protocol

For complex tasks, capture concise traces:
1. Thought
2. Action
3. Observation
4. Decision

## Confidence Gate

- High: proceed and verify.
- Medium: proceed with explicit assumptions and extra checks.
- Low: ask for clarification before irreversible changes.

## Verification Protocol

1. Validate against stated requirements.
2. Confirm no unintended scope expansion.
3. Re-check critical paths impacted by the task.

## Self-Improvement

- Record what worked well.
- Note what should be improved next iteration.
- Capture reusable patterns discovered.
`,
		title, model, mode,
		title,
		model,
		title, description,
		strings.Join(capabilityLines, "\n"),
	)

	return []GeneratedFile{
		{
			Path: fmt.Sprintf("library/agents/%s.md", func() string {
				if slug != "" {
					return slug
				}
				return "new-agent"
			}()),
			Content: content,
		},
	}, nil
}
