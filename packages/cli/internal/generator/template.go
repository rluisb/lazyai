package generator

import (
	"fmt"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TemplateGenerator generates template files.
// Ported from src/generators/template.ts.
type TemplateGenerator struct{}

func (g *TemplateGenerator) Type() types.ArtifactType { return types.ArtifactTypeTemplate }

func (g *TemplateGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:     "sections",
			Label:   "Sections (comma-separated)",
			Type:    "text",
			Default: "Objective,Subtasks,Files to Touch,Done When",
		},
		{
			Key:     "fields",
			Label:   "Fields (comma-separated)",
			Type:    "text",
			Default: "Phase,Status,Depends on",
		},
	}
}

func (g *TemplateGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	title := ToTitleCase(slug)
	if title == "" {
		title = config.Name
	}

	sections := listFromCSV(getAnswer(config.Answers, "sections", ""), []string{
		"Objective",
		"Subtasks",
		"Files to Touch",
		"Done When",
	})

	fields := listFromCSV(getAnswer(config.Answers, "fields", ""), []string{
		"Phase",
		"Status",
		"Depends on",
	})

	var fieldLines []string
	for _, field := range fields {
		fieldLines = append(fieldLines, fmt.Sprintf("**%s:** [value]", field))
	}

	var sectionBlocks []string
	for _, section := range sections {
		sectionBlocks = append(sectionBlocks, fmt.Sprintf("## %s\n\n[%s details]", section, section))
	}

	content := fmt.Sprintf(`# %s

%s

---

%s
`,
		title,
		strings.Join(fieldLines, "\n"),
		strings.Join(sectionBlocks, "\n\n"),
	)

	fileName := slug
	if fileName == "" {
		fileName = "new-template"
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("library/templates/%s.md", fileName),
			Content: content,
		},
	}, nil
}

// listFromCSV parses a comma-separated list, falling back to the provided list.
func listFromCSV(raw string, fallback []string) []string {
	var items []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}
	if len(items) > 0 {
		return items
	}
	return fallback
}
