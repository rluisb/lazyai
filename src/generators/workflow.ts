import { join } from 'node:path'
import { ensureDir, fileExists, listDir } from '../utils/files.js'
import type { GeneratedFile, Generator, GeneratorConfig, PromptQuestion } from './types.js'

type WorkflowRefType = 'agent' | 'skill' | 'prompt' | 'template'

interface WorkflowRef {
  type: WorkflowRefType
  name: string
}

interface WorkflowStep {
  name: string
  refs: WorkflowRef[]
}

export interface DiscoveredArtifacts {
  agents: string[]
  skills: string[]
  prompts: string[]
  templates: string[]
}

const REF_TYPE_TO_DIR: Record<WorkflowRefType, keyof DiscoveredArtifacts> = {
  agent: 'agents',
  skill: 'skills',
  prompt: 'prompts',
  template: 'templates',
}

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

function parseStepSpec(spec: string): WorkflowStep | undefined {
  const [rawName, rawRefs] = spec.split(':', 2)
  const stepName = rawName?.trim()
  if (!stepName) {
    return undefined
  }

  const refs: WorkflowRef[] = []
  for (const refChunk of (rawRefs ?? '').split(',').map((item) => item.trim()).filter(Boolean)) {
    const [rawType, rawValue] = refChunk.split('=', 2)
    const type = rawType?.trim() as WorkflowRefType | undefined
    const name = rawValue?.trim()
    if (!type || !name) {
      continue
    }
    if (!['agent', 'skill', 'prompt', 'template'].includes(type)) {
      continue
    }
    refs.push({ type, name })
  }

  return { name: stepName, refs }
}

function getStepsFromAnswers(answers: Record<string, unknown> | undefined): WorkflowStep[] {
  const rawSteps = answers?.steps

  if (Array.isArray(rawSteps)) {
    return rawSteps
      .map((value) => {
        if (typeof value === 'string') {
          return parseStepSpec(value)
        }
        return undefined
      })
      .filter((value): value is WorkflowStep => Boolean(value))
  }

  if (typeof rawSteps === 'string' && rawSteps.trim()) {
    return rawSteps
      .split(/\n+/)
      .map((line) => parseStepSpec(line))
      .filter((value): value is WorkflowStep => Boolean(value))
  }

  return []
}

function toRelativeRef(ref: WorkflowRef): string {
  const dir = REF_TYPE_TO_DIR[ref.type]
  return `../${dir}/${toSlug(ref.name)}.md`
}

export function discoverLibraryArtifacts(targetDir: string): DiscoveredArtifacts {
  const libraryDir = join(targetDir, 'library')

  const readNames = (subdir: keyof DiscoveredArtifacts): string[] => {
    const dir = join(libraryDir, subdir)
    if (!fileExists(dir)) {
      return []
    }

    return listDir(dir)
      .filter((entry) => entry.endsWith('.md'))
      .map((entry) => entry.replace(/\.md$/, ''))
      .sort((a, b) => a.localeCompare(b))
  }

  return {
    agents: readNames('agents'),
    skills: readNames('skills'),
    prompts: readNames('prompts'),
    templates: readNames('templates'),
  }
}

export class WorkflowGenerator implements Generator {
  readonly type = 'workflow' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'steps',
        label: 'Workflow steps (one per line: Step Name:agent=foo,skill=bar)',
        type: 'text',
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    ensureDir(join(config.targetDir, 'library/workflows'))

    const slug = toSlug(config.name)
    const title = toTitleCase(slug || config.name)
    const discovered = discoverLibraryArtifacts(config.targetDir)

    const steps = getStepsFromAnswers(config.answers)
    const workflowSteps =
      steps.length > 0
        ? steps
        : [
            {
              name: 'Execute workflow',
              refs: [],
            },
          ]

    const warnings: string[] = []
    for (const step of workflowSteps) {
      for (const ref of step.refs) {
        const key = REF_TYPE_TO_DIR[ref.type]
        const exists = discovered[key].includes(toSlug(ref.name))
        if (!exists) {
          warnings.push(`${ref.type} "${ref.name}" not found in library/${key}`)
        }
      }
    }

    if (warnings.length > 0) {
      for (const warning of warnings) {
        console.warn(`⚠️  Workflow warning: ${warning}`)
      }
    }

    const stepsBlock = workflowSteps
      .map((step, index) => {
        const refs = step.refs.length
          ? step.refs.map((ref) => `   - ${ref.type}: [${toSlug(ref.name)}](${toRelativeRef(ref)})`).join('\n')
          : '   - references: none'

        return `${index + 1}. **${step.name}**\n${refs}`
      })
      .join('\n\n')

    const warningBlock = warnings.length
      ? `\n## Warnings\n\n${warnings.map((warning) => `- ${warning}`).join('\n')}\n`
      : ''

    const content = `# ${title} Workflow

**Goal:** ${config.description ?? `Run the ${title.toLowerCase()} workflow.`}

## Steps

${stepsBlock}${warningBlock}
`

    return [
      {
        path: `library/workflows/${slug || 'new-workflow'}.md`,
        content,
      },
    ]
  }
}
