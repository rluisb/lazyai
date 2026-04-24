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

export class DomainGenerator implements Generator {
  readonly type = 'domain' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'description',
        label: 'Domain skill description',
        type: 'text',
        required: false,
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const description = String(config.answers?.description ?? config.description ?? `Domain knowledge for ${title.toLowerCase()} work.`)

    const content = `---
kind: domain-skill
name: ${slug || 'new-domain'}
description: ${description}
applies_to:
  - scout
  - planner
  - builder
  - reviewer
knowledge_areas:
  - ${slug || 'domain-area'}
allowed_tools:
  - Read
  - Grep
  - Glob
  - Edit
  - Write
model_hint: sonnet
---

# ${title} Domain Skill

Use this skill to inject project-specific domain knowledge for ${title.toLowerCase()} tasks.

- surface domain constraints early
- prefer explicit contracts and invariants
- document risks, assumptions, and edge cases
`

    return [
      {
        path: `.ai/orchestration/skills/domains/${slug || 'new-domain'}.md`,
        content,
      },
    ]
  }
}
