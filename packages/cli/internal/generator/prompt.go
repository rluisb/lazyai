package generator

import (
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// PromptGenerator generates prompt template files.
// Ported from src/generators/prompt.ts.
type PromptGenerator struct{}

func (g *PromptGenerator) Type() types.ArtifactType { return types.ArtifactTypePrompt }

func (g *PromptGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:     "taskContext",
			Label:   "Task context placeholder",
			Type:    "text",
			Default: "[Task Name]",
		},
		{
			Key:     "outputFormat",
			Label:   "Output format description",
			Type:    "text",
			Default: "[Describe expected output]",
		},
	}
}

func (g *PromptGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	title := ToTitleCase(slug)
	if title == "" {
		title = config.Name
	}

	taskContext := getAnswer(config.Answers, "taskContext", "[Task Name]")
	outputFormat := getAnswer(config.Answers, "outputFormat", "[Describe expected output]")
	spec := config.Description
	if spec == "" {
		spec = "[Link to Task Spec]"
	}

	content := fmt.Sprintf(`# %s Prompt

**Task:** %s
**Spec:** %s

---

## Instructions

1. Read existing context before making changes.
2. Keep scope aligned with task requirements.
3. Produce explicit, verifiable outputs.
4. Highlight risks, assumptions, and unknowns.

## Output Format

`+"```"+`
%s
`+"```"+`
`,
		title,
		taskContext,
		spec,
		outputFormat,
	)

	fileName := slug
	if fileName == "" {
		fileName = "new-prompt"
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("library/prompts/%s.md", fileName),
			Content: content,
		},
	}, nil
}
