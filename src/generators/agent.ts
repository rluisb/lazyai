import type { GeneratedFile, Generator, GeneratorConfig, PromptQuestion } from './types.js'

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

export class AgentGenerator implements Generator {
  readonly type = 'agent' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'model',
        label: 'Model',
        type: 'select',
        required: true,
        default: 'gpt-4o',
        options: [
          { value: 'gpt-4o', label: 'gpt-4o' },
          { value: 'claude-sonnet', label: 'claude-sonnet' },
          { value: 'gemini-pro', label: 'gemini-pro' },
          { value: 'other', label: 'other' },
        ],
      },
      {
        key: 'mode',
        label: 'Mode',
        type: 'select',
        required: true,
        default: 'interactive',
        options: [
          { value: 'autonomous', label: 'autonomous' },
          { value: 'interactive', label: 'interactive' },
          { value: 'hybrid', label: 'hybrid' },
        ],
      },
      {
        key: 'tools',
        label: 'Tools (comma-separated)',
        type: 'text',
        default: '',
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const model = String(config.answers?.model ?? 'gpt-4o')
    const mode = String(config.answers?.mode ?? 'interactive')
    const tools = String(config.answers?.tools ?? '')
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)

    const capabilityLines = tools.length
      ? tools.map((tool) => `- Use ${tool} effectively when needed`)
      : ['- Execute tasks aligned to this role']

    const content = `---
name: ${title}
model: ${model}
mode: ${mode}
---

# ${title} Agent

## Identity

You are ${title} — an AI specialist focused on ${config.description ?? 'high-quality task execution'}.

## Capability

${capabilityLines.join('\n')}

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
`

    return [
      {
        path: `library/agents/${slug || 'new-agent'}.md`,
        content,
      },
    ]
  }
}
