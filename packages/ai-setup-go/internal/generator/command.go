package generator

import (
	"fmt"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// CommandGenerator generates command files.
// Ported from src/generators/command.ts.
type CommandGenerator struct{}

func (g *CommandGenerator) Type() types.ArtifactType { return types.ArtifactTypeCommand }

func (g *CommandGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:   "arguments",
			Label: "Arguments signature (example: [name] or <tool>)",
			Type:  "text",
		},
		{
			Key:   "flagsDescription",
			Label: "Flags description (human readable)",
			Type:  "text",
		},
	}
}

func (g *CommandGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	fnSuffix := toFunctionName(slug)
	if fnSuffix == "" {
		fnSuffix = "NewCommand"
	}

	argsSignature := strings.TrimSpace(getAnswer(config.Answers, "arguments", ""))
	commandNameWithArgs := slug
	if commandNameWithArgs == "" {
		commandNameWithArgs = "new-command"
	}
	if argsSignature != "" {
		commandNameWithArgs += " " + argsSignature
	}

	flagsDescription := getAnswer(config.Answers, "flagsDescription", "No additional flags yet.")
	escapedFlags := strings.ReplaceAll(flagsDescription, "'", "\\'")

	description := config.Description
	if description == "" {
		description = fmt.Sprintf("Run %s command", slug)
	}

	content := fmt.Sprintf(`import type { Command } from 'commander'

interface %sOptions {
  interactive: boolean
}

export function register%s(program: Command): void {
  program
    .command('%s')
    .description('%s')
    .option('--no-interactive', 'Disable interactive mode')
    .action(async (_name: string | undefined, opts: %sOptions) => {
      if (!opts.interactive) {
        console.log('%s executed in non-interactive mode')
        return
      }

      console.log('%s executed')
      console.log('Flags: %s')
    })
}
`,
		fnSuffix,
		fnSuffix,
		commandNameWithArgs,
		description,
		fnSuffix,
		slug,
		slug,
		escapedFlags,
	)

	fileName := slug
	if fileName == "" {
		fileName = "new-command"
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("src/commands/%s.ts", fileName),
			Content: content,
		},
	}, nil
}

// toFunctionName converts a slug to a PascalCase function name.
func toFunctionName(slug string) string {
	if slug == "" {
		return ""
	}

	parts := strings.Split(slug, "-")
	var result []byte
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(result) == 0 {
			result = append(result, []byte(part)...)
		} else {
			result = append(result, byte(strings.ToUpper(part[:1])[0]))
			result = append(result, []byte(part[1:])...)
		}
	}

	// Capitalize first letter.
	if len(result) > 0 && result[0] >= 'a' && result[0] <= 'z' {
		result[0] = result[0] - 32
	}

	return string(result)
}
