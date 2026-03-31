import type { Generator, GeneratorConfig, GeneratedFile, PromptQuestion } from './types.js'

function toSlug(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function toTitleCase(value: string): string {
  return value
    .split(/[-_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

export class PromptGenerator implements Generator {
  readonly type = 'prompt' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'taskContext',
        label: 'Task context placeholder',
        type: 'text',
        default: '[Task Name]',
      },
      {
        key: 'outputFormat',
        label: 'Output format description',
        type: 'text',
        default: '[Describe expected output]',
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const taskContext = String(config.answers?.taskContext ?? '[Task Name]')
    const outputFormat = String(config.answers?.outputFormat ?? '[Describe expected output]')

    const content = `# ${title} Prompt

**Task:** ${taskContext}
**Spec:** ${config.description ?? '[Link to Task Spec]'}

---

## Instructions

1. Read existing context before making changes.
2. Keep scope aligned with task requirements.
3. Produce explicit, verifiable outputs.
4. Highlight risks, assumptions, and unknowns.

## Output Format

\`\`\`
${outputFormat}
\`\`\`
`

    return [
      {
        path: `library/prompts/${slug || 'new-prompt'}.md`,
        content,
      },
    ]
  }
}
