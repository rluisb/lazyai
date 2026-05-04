package generator

import (
	"fmt"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// SkillGenerator generates skill SKILL.md files.
// Ported from src/generators/skill.ts.
type SkillGenerator struct{}

func (g *SkillGenerator) Type() types.ArtifactType { return types.ArtifactTypeSkill }

func (g *SkillGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "command",
			Label:    "Command trigger (without leading slash)",
			Type:     "text",
			Required: true,
		},
		{
			Key:      "steps",
			Label:    "Workflow steps (newline or numbered list)",
			Type:     "text",
			Required: false,
		},
	}
}

func (g *SkillGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	title := ToTitleCase(slug)
	if title == "" {
		title = config.Name
	}

	command := getAnswer(config.Answers, "command", slug)
	if command == "" {
		command = "command"
	}

	goal := config.Description
	if goal == "" {
		goal = fmt.Sprintf("Execute %s effectively.", strings.ToLower(title))
	}

	steps := normalizeSteps(getAnswer(config.Answers, "steps", ""))

	var stepLines []string
	for i, step := range steps {
		stepLines = append(stepLines, fmt.Sprintf("%d. %s", i+1, step))
	}

	content := fmt.Sprintf(`# %s Skill

**Command:** /%s [args]
**Goal:** %s

---

## Workflow

%s

## Principles

- Keep actions scoped and reversible.
- Prefer explicit assumptions over hidden behavior.
- Verify outputs before completion.

## Trace Protocol

For complex tasks, provide concise traces:

1. Thought
2. Action
3. Observation
4. Decision

## Output Format

`+"```"+`markdown
## Skill Run: %s

### Steps Completed
- [step]: [status]

### Evidence
- [result or artifact]

### Follow-ups
- [if any]
`+"```"+`
`,
		title,
		command,
		goal,
		strings.Join(stepLines, "\n"),
		title,
	)

	fileName := slug
	if fileName == "" {
		fileName = "new-skill"
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("library/skills/%s.md", fileName),
			Content: content,
		},
	}, nil
}

// normalizeSteps parses a raw steps string into a list of steps.
func normalizeSteps(raw string) []string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove leading number+period or number+paren.
		if len(line) > 2 {
			if line[0] >= '0' && line[0] <= '9' {
				if line[1] == '.' || line[1] == ')' {
					line = strings.TrimSpace(line[2:])
				} else if len(line) > 3 && (line[1] >= '0' && line[1] <= '9') && (line[2] == '.' || line[2] == ')') {
					line = strings.TrimSpace(line[3:])
				}
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) > 0 {
		return lines
	}

	return []string{"Clarify scope", "Execute minimal safe actions", "Verify outcomes"}
}
