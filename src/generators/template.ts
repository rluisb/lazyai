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

function listFromCsv(raw: string, fallback: string[]): string[] {
  const list = raw
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
  return list.length > 0 ? list : fallback
}

export class TemplateGenerator implements Generator {
  readonly type = 'template' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'sections',
        label: 'Sections (comma-separated)',
        type: 'text',
        default: 'Objective,Subtasks,Files to Touch,Done When',
      },
      {
        key: 'fields',
        label: 'Fields (comma-separated)',
        type: 'text',
        default: 'Phase,Status,Depends on',
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const sections = listFromCsv(String(config.answers?.sections ?? ''), [
      'Objective',
      'Subtasks',
      'Files to Touch',
      'Done When',
    ])
    const fields = listFromCsv(String(config.answers?.fields ?? ''), ['Phase', 'Status', 'Depends on'])

    const fieldLines = fields.map((field) => `**${field}:** [value]`).join('\n')
    const sectionBlocks = sections.map((section) => `## ${section}\n\n[${section} details]`).join('\n\n')

    const content = `# ${title}

${fieldLines}

---

${sectionBlocks}
`

    return [
      {
        path: `library/templates/${slug || 'new-template'}.md`,
        content,
      },
    ]
  }
}
