import type { Command } from 'commander'
import * as p from '@clack/prompts'
import { join } from 'node:path'
import type { ArtifactType } from '../types.js'
import { Errors } from '../errors/index.js'
import { fileExists, writeFile } from '../utils/files.js'
import { validateRequiredText } from '../utils/validation.js'
import { GeneratorRegistry } from '../generators/registry.js'
import { discoverLibraryArtifacts } from '../generators/workflow.js'

interface CreateOptions {
  type?: ArtifactType
  name?: string
  description?: string
  force?: boolean
  interactive: boolean
  model?: string
  mode?: string
  tools?: string
  command?: string
  steps?: string
  arguments?: string
  flagsDescription?: string
  taskContext?: string
  outputFormat?: string
  sections?: string
  fields?: string
  step?: string[]
}

function parseArtifactType(value: string): ArtifactType {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'agent' || normalized === 'skill' || normalized === 'command' || normalized === 'prompt' || normalized === 'template' || normalized === 'workflow') {
    return normalized
  }
  throw Errors.invalidInput(`invalid artifact type: ${value}`)
}

function ensurePromptText(value: string | symbol): string {
  if (typeof value !== 'string' || value.trim().length === 0) {
    throw Errors.invalidInput('prompt value cannot be empty')
  }
  return value
}

async function askText(message: string, placeholder?: string): Promise<string> {
  const textOptions: { message: string; placeholder?: string; defaultValue?: string; validate: (value: string) => string | undefined } = {
    message,
    validate: (value) => validateRequiredText(value, message),
  }
  if (placeholder !== undefined) {
    textOptions.placeholder = placeholder
    textOptions.defaultValue = placeholder
  }

  const result = await p.text(textOptions)
  if (p.isCancel(result)) {
    p.cancel('Create cancelled.')
    throw Errors.userCancelled()
  }
  return ensurePromptText(result)
}

async function buildWorkflowStepsInteractively(targetDir: string): Promise<string[]> {
  const discovered = discoverLibraryArtifacts(targetDir)
  const typeToItems: Array<{ type: 'agent' | 'skill' | 'prompt' | 'template'; items: string[] }> = [
    { type: 'agent', items: discovered.agents },
    { type: 'skill', items: discovered.skills },
    { type: 'prompt', items: discovered.prompts },
    { type: 'template', items: discovered.templates },
  ]

  const steps: string[] = []
  let keepAdding = true

  while (keepAdding) {
    const stepName = await askText('Step name?', 'Research')

    const refOptions = typeToItems.flatMap(({ type, items }) =>
      items.map((item) => ({
        value: `${type}=${item}`,
        label: `${type}: ${item}`,
      })),
    )

    const selected = refOptions.length
      ? await p.multiselect({
          message: 'Select artifact references for this step',
          options: refOptions,
          required: false,
        })
      : []

    if (p.isCancel(selected)) {
      p.cancel('Create cancelled.')
      throw Errors.userCancelled()
    }

    const refs = Array.isArray(selected) ? selected.join(',') : ''
    const stepSpec = refs ? `${stepName}:${refs}` : `${stepName}`
    steps.push(stepSpec)

    const addAnother = await p.confirm({
      message: 'Add another step?',
      initialValue: true,
    })

    if (p.isCancel(addAnother)) {
      p.cancel('Create cancelled.')
      throw Errors.userCancelled()
    }

    keepAdding = Boolean(addAnother)
  }

  return steps
}

function extractAnswersFromOptions(type: ArtifactType, opts: CreateOptions): Record<string, unknown> {
  const answers: Record<string, unknown> = {}

  if (type === 'agent') {
    if (opts.model) answers.model = opts.model
    if (opts.mode) answers.mode = opts.mode
    if (opts.tools) answers.tools = opts.tools
  }

  if (type === 'skill') {
    if (opts.command) answers.command = opts.command
    if (opts.steps) answers.steps = opts.steps
  }

  if (type === 'command') {
    if (opts.arguments) answers.arguments = opts.arguments
    if (opts.flagsDescription) answers.flagsDescription = opts.flagsDescription
  }

  if (type === 'prompt') {
    if (opts.taskContext) answers.taskContext = opts.taskContext
    if (opts.outputFormat) answers.outputFormat = opts.outputFormat
  }

  if (type === 'template') {
    if (opts.sections) answers.sections = opts.sections
    if (opts.fields) answers.fields = opts.fields
  }

  if (type === 'workflow') {
    if (opts.steps) answers.steps = opts.steps
    if (opts.step && opts.step.length > 0) answers.steps = opts.step
  }

  return answers
}

async function runCreate(type: ArtifactType, positionalName: string | undefined, opts: CreateOptions): Promise<void> {
  const registry = new GeneratorRegistry()
  const generator = registry.get(type)

  if (!generator) {
    throw Errors.missingDependency(`generator:${type}`)
  }

  const targetDir = process.cwd()
  const name = positionalName ?? opts.name ?? (opts.interactive ? await askText(`Name for ${type}?`) : undefined)

  if (!name) {
    throw Errors.invalidInput(`a name is required for create ${type} in non-interactive mode`)
  }

  const answers = extractAnswersFromOptions(type, opts)

  if (type === 'workflow' && opts.interactive && !answers.steps) {
    answers.steps = await buildWorkflowStepsInteractively(targetDir)
  }

  if (opts.interactive) {
    const questions = generator.getPromptQuestions()

    for (const question of questions) {
      if (answers[question.key] !== undefined) {
        continue
      }

      if (question.type === 'text') {
        const textOptions: { message: string; placeholder?: string; defaultValue?: string; validate: (value: string) => string | undefined } = {
          message: question.label,
          validate: (value) => validateRequiredText(value, question.label),
        }
        if (question.default !== undefined) {
          textOptions.placeholder = question.default
          textOptions.defaultValue = question.default
        }

        const result = await p.text(textOptions)

        if (p.isCancel(result)) {
          p.cancel('Create cancelled.')
          throw Errors.userCancelled()
        }

        if (result || question.required) {
          answers[question.key] = result
        }
      }

      if (question.type === 'select' && question.options) {
        const result = await p.select({
          message: question.label,
          options: question.options,
          initialValue: question.default,
        })

        if (p.isCancel(result)) {
          p.cancel('Create cancelled.')
          throw Errors.userCancelled()
        }

        answers[question.key] = result
      }

      if (question.type === 'multiselect' && question.options) {
        const result = await p.multiselect({
          message: question.label,
          options: question.options,
          required: Boolean(question.required),
        })

        if (p.isCancel(result)) {
          p.cancel('Create cancelled.')
          throw Errors.userCancelled()
        }

        answers[question.key] = result
      }
    }
  }

  p.intro(`Creating ${type}: ${name}`)

  const generatorConfig: {
    name: string
    description?: string
    force?: boolean
    targetDir: string
    answers: Record<string, unknown>
  } = {
    name,
    targetDir,
    answers,
  }

  if (opts.description !== undefined) {
    generatorConfig.description = opts.description
  }

  if (opts.force !== undefined) {
    generatorConfig.force = opts.force
  }

  const generated = await generator.generate(generatorConfig)

  for (const file of generated) {
    const outputPath = join(targetDir, file.path)
    if (fileExists(outputPath) && !opts.force) {
      throw Errors.invalidInput(`file already exists: ${file.path} (use --force to overwrite)`)
    }
    writeFile(outputPath, file.content)
  }

  p.outro(`✅ Created ${generated.length} file(s) for ${type}`)
}

function registerCreateSubcommand(createCmd: Command, type: ArtifactType): void {
  const sub = createCmd
    .command(`${type} [name]`)
    .description(`Create a new ${type}`)
    .option('--name <name>', 'Artifact name (alternative to positional [name])')
    .option('--description <description>', 'Artifact description')
    .option('--force', 'Overwrite files if they already exist')
    .option('--no-interactive', 'Disable interactive prompts')

  if (type === 'agent') {
    sub.option('--model <model>', 'Agent model')
    sub.option('--mode <mode>', 'Agent mode: autonomous | interactive | hybrid')
    sub.option('--tools <tools>', 'Comma-separated tools')
  }

  if (type === 'skill') {
    sub.option('--command <command>', 'Command trigger')
    sub.option('--steps <steps>', 'Workflow steps (newline-delimited)')
  }

  if (type === 'command') {
    sub.option('--arguments <arguments>', 'Command arguments signature')
    sub.option('--flags-description <flagsDescription>', 'Flags description')
  }

  if (type === 'prompt') {
    sub.option('--task-context <taskContext>', 'Prompt task context')
    sub.option('--output-format <outputFormat>', 'Prompt output format')
  }

  if (type === 'template') {
    sub.option('--sections <sections>', 'Comma-separated section names')
    sub.option('--fields <fields>', 'Comma-separated field names')
  }

  if (type === 'workflow') {
    sub.option('--steps <steps>', 'Workflow steps as newline-delimited step specs')
    sub.option(
      '--step <step>',
      'Add one step (format: "Research:agent=scout,skill=research")',
      (value: string, previous: string[] = []) => [...previous, value],
    )
  }

  sub.action(async (name: string | undefined, opts: CreateOptions) => {
    await runCreate(type, name, opts)
  })
}

export function registerCreate(program: Command): void {
  const create = program
    .command('create')
    .description('Create a new agent, skill, command, prompt, template, or workflow')
    .option('--type <type>', 'Artifact type for bare create command')
    .option('--name <name>', 'Artifact name')
    .option('--description <description>', 'Artifact description')
    .option('--force', 'Overwrite files if they already exist')
    .option('--no-interactive', 'Disable interactive prompts')
    .action(async (opts: CreateOptions) => {
      let selectedType: string | undefined = opts.type

      if (!selectedType && opts.interactive) {
        const selected = await p.select({
          message: 'What would you like to create?',
          options: [
            { value: 'agent', label: 'agent' },
            { value: 'skill', label: 'skill' },
            { value: 'command', label: 'command' },
            { value: 'prompt', label: 'prompt' },
            { value: 'template', label: 'template' },
            { value: 'workflow', label: 'workflow' },
          ],
        })

        if (p.isCancel(selected)) {
          p.cancel('Create cancelled.')
          throw Errors.userCancelled()
        }

        selectedType = String(selected)
      }

      if (!selectedType) {
        throw Errors.invalidInput('type is required in non-interactive mode (use --type)')
      }

      await runCreate(parseArtifactType(selectedType), undefined, opts)
    })

  registerCreateSubcommand(create, 'agent')
  registerCreateSubcommand(create, 'skill')
  registerCreateSubcommand(create, 'command')
  registerCreateSubcommand(create, 'prompt')
  registerCreateSubcommand(create, 'template')
  registerCreateSubcommand(create, 'workflow')
}
