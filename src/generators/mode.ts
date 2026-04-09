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

export class ModeGenerator implements Generator {
  readonly type = 'mode' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'description',
        label: 'Mode skill description',
        type: 'text',
        required: false,
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const description = String(config.answers?.description ?? config.description ?? `Behavioral operating mode for ${title.toLowerCase()} execution.`)

    const content = `---
kind: mode-skill
name: ${slug || 'new-mode'}
description: ${description}
behavior:
  - keep work aligned to the active plan
  - surface risks before high-cost execution
  - prefer deterministic, auditable steps
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
approval_policy: normal
model_hint: opus
---

# ${title} Mode Skill

Use this skill to modify how the active agent behaves while preserving the base role.

- define when to ask for confirmation
- clarify autonomy expectations
- document trade-offs and stopping conditions
`

    return [
      {
        path: `.ai/orchestration/skills/modes/${slug || 'new-mode'}.md`,
        content,
      },
    ]
  }
}
