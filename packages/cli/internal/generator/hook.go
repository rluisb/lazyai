package generator

import (
	"fmt"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// HookGenerator generates hook policy files.
type HookGenerator struct{}

func (g *HookGenerator) Type() types.ArtifactType { return types.ArtifactTypeHook }

func (g *HookGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "purpose",
			Label:    "What does this hook prevent or automate?",
			Type:     "text",
			Required: true,
		},
		{
			Key:     "events",
			Label:   "Hook events (comma-separated: PreToolUse, PostToolUse, Stop, etc.)",
			Type:    "text",
			Default: "PreToolUse",
		},
		{
			Key:   "denied",
			Label: "Denied behaviors (comma-separated list)",
			Type:  "text",
		},
		{
			Key:   "allowed",
			Label: "Allowed behaviors (comma-separated list)",
			Type:  "text",
		},
	}
}

func (g *HookGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	title := ToTitleCase(slug)
	if title == "" {
		title = config.Name
	}

	purpose := getAnswer(config.Answers, "purpose", "")
	if purpose == "" {
		purpose = config.Description
	}
	if purpose == "" {
		purpose = fmt.Sprintf("Policy for %s.", strings.ToLower(title))
	}

	events := listFromCSV(getAnswer(config.Answers, "events", ""), []string{"PreToolUse"})
	denied := listFromCSV(getAnswer(config.Answers, "denied", ""), []string{"[denied behavior]"})
	allowed := listFromCSV(getAnswer(config.Answers, "allowed", ""), []string{"[allowed behavior]"})

	var eventLines []string
	for _, e := range events {
		eventLines = append(eventLines, fmt.Sprintf("- `%s`", e))
	}

	var deniedLines []string
	for _, d := range denied {
		deniedLines = append(deniedLines, fmt.Sprintf("- %s", d))
	}

	var allowedLines []string
	for _, a := range allowed {
		allowedLines = append(allowedLines, fmt.Sprintf("- %s", a))
	}

	content := fmt.Sprintf(`# %s Policy

## Purpose

%s

## Events

%s

## Decision

- Allow when: input matches allowed patterns.
- Deny when: input matches denied patterns.

## Allowed

%s

## Denied

%s

## Fail-Closed Semantics

If the hook adapter cannot be loaded or fails to execute, deny the action.
`,
		title,
		purpose,
		strings.Join(eventLines, "\n"),
		strings.Join(allowedLines, "\n"),
		strings.Join(deniedLines, "\n"),
	)

	fileName := slug
	if fileName == "" {
		fileName = "new-hook"
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("library/hooks/%s.md", fileName),
			Content: content,
		},
	}, nil
}
