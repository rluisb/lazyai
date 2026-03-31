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

function normalizeSteps(raw: string): string[] {
  const lines = raw
    .split(/\n+/)
    .map((line) => line.replace(/^\d+[.)]\s*/, '').trim())
    .filter(Boolean)

  if (lines.length > 0) {
    return lines
  }

  return ['Clarify scope', 'Execute minimal safe actions', 'Verify outcomes']
}

export class SkillGenerator implements Generator {
  readonly type = 'skill' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'command',
        label: 'Command trigger (without leading slash)',
        type: 'text',
        required: true,
      },
      {
        key: 'steps',
        label: 'Workflow steps (newline or numbered list)',
        type: 'text',
        required: false,
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const command = String(config.answers?.command ?? (slug || 'command'))
    const goal = config.description ?? `Execute ${title.toLowerCase()} effectively.`
    const steps = normalizeSteps(String(config.answers?.steps ?? ''))

    const content = `# ${title} Skill

**Command:** /${command} [args]
**Goal:** ${goal}

---

## Workflow

${steps.map((step, index) => `${index + 1}. ${step}`).join('\n')}

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

\`\`\`markdown
## Skill Run: ${title}

### Steps Completed
- [step]: [status]

### Evidence
- [result or artifact]

### Follow-ups
- [if any]
\`\`\`
`

    return [
      {
        path: `library/skills/${slug || 'new-skill'}.md`,
        content,
      },
    ]
  }
}
