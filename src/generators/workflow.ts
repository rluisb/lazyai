import type { GeneratedFile, Generator, GeneratorConfig, PromptQuestion } from './types.js'

function toSlug(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/\.[a-z0-9]+$/i, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function pickText(value: unknown): string | undefined {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

export class WorkflowGenerator implements Generator {
  readonly type = 'workflow' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'chain',
        label: 'Primary chain reference',
        type: 'text',
        required: false,
        default: 'feature',
      },
      {
        key: 'team',
        label: 'Optional review/synthesis team reference',
        type: 'text',
        required: false,
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name) || 'new-workflow'
    const chainRef = pickText(config.answers?.chain) ?? 'feature'
    const teamRef = pickText(config.answers?.team)
    const chainPhaseId = `run-${toSlug(chainRef) || 'chain'}`
    const teamPhaseId = teamRef ? `run-${toSlug(teamRef) || 'team'}` : undefined

    const phases: Array<Record<string, unknown>> = [
      {
        id: chainPhaseId,
        kind: 'chain',
        ref: chainRef,
        on: {
          success: teamPhaseId ?? 'complete',
          failure: 'handoff',
        },
      },
    ]

    if (teamRef && teamPhaseId) {
      phases.push({
        id: teamPhaseId,
        kind: 'team',
        ref: teamRef,
        on: {
          success: 'complete',
          failure: 'handoff',
        },
      })
    }

    phases.push(
      { id: 'handoff', kind: 'terminal' },
      { id: 'complete', kind: 'terminal' },
    )

    const workflow = {
      kind: 'workflow',
      name: slug,
      description: config.description ?? `Workflow scaffold for ${slug}.`,
      version: '1.0.0',
      entry: chainPhaseId,
      phases,
    }

    return [
      {
        path: `.ai/orchestration/workflows/${slug}.json`,
        content: `${JSON.stringify(workflow, null, 2)}\n`,
      },
    ]
  }
}
