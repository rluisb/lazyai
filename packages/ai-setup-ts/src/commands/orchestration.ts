import path from 'node:path'
import { fileURLToPath } from 'node:url'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import { Errors } from '../errors/index.js'
import { GeneratorRegistry } from '../generators/registry.js'
import {
  getOrchestrationCounts,
  listOrchestrationItems,
  type OrchestrationListCategory,
} from '../orchestration/catalog.js'
import type { ArtifactType } from '../types.js'
import { fileExists, resolveLibraryDir, writeFile } from '../utils/files.js'

interface OrchestrationCreateOptions {
  description?: string
  force?: boolean
  interactive?: boolean
  chain?: string
  team?: string
}

const CREATE_TYPES: ArtifactType[] = ['workflow', 'domain', 'mode']
const LIST_CATEGORIES: OrchestrationListCategory[] = ['workflows', 'chains', 'teams', 'domains', 'modes']

function toCreateType(value: string): ArtifactType {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'workflow' || normalized === 'domain' || normalized === 'mode') {
    return normalized
  }
  throw Errors.invalidInput(`unsupported orchestration create type: ${value}`)
}

function toListCategory(value?: string): OrchestrationListCategory | undefined {
  if (!value) {
    return undefined
  }

  const normalized = value.trim().toLowerCase()
  if (
    normalized === 'workflows' ||
    normalized === 'chains' ||
    normalized === 'teams' ||
    normalized === 'domains' ||
    normalized === 'modes'
  ) {
    return normalized
  }

  throw Errors.invalidInput(`unsupported orchestration list kind: ${value}`)
}

async function createOrchestrationArtifact(type: ArtifactType, name: string, opts: OrchestrationCreateOptions): Promise<void> {
  const registry = new GeneratorRegistry()
  const generator = registry.get(type)
  if (!generator) {
    throw Errors.missingDependency(`generator:${type}`)
  }

  const answers: Record<string, unknown> = {}
  if (type === 'workflow') {
    if (opts.chain) answers.chain = opts.chain
    if (opts.team) answers.team = opts.team
  }

  const config = {
    name,
    targetDir: process.cwd(),
    answers,
  } as {
    name: string
    targetDir: string
    answers: Record<string, unknown>
    description?: string
    force?: boolean
  }

  if (opts.description !== undefined) {
    config.description = opts.description
  }

  if (opts.force !== undefined) {
    config.force = opts.force
  }

  const generated = await generator.generate(config)

  for (const file of generated) {
    const outputPath = path.join(process.cwd(), file.path)
    if (fileExists(outputPath) && !opts.force) {
      throw Errors.invalidInput(`file already exists: ${file.path} (use --force to overwrite)`)
    }
    writeFile(outputPath, file.content)
  }
}

export function registerOrchestration(program: Command): void {
  const orchestration = program
    .command('orchestration')
    .description('Manage orchestration workflows, chains, teams, domains, and modes')

  orchestration
    .command('list [kind]')
    .description('List orchestration artifacts')
    .option('--json', 'Output as JSON')
    .action((kind: string | undefined, opts: { json?: boolean }) => {
      const libraryDir = resolveLibraryDir(path.dirname(fileURLToPath(import.meta.url)))
      const category = toListCategory(kind)

      if (category) {
        const result = {
          [category]: listOrchestrationItems(process.cwd(), libraryDir, category),
        }

        if (opts.json) {
          console.log(JSON.stringify(result, null, 2))
          return
        }

        p.log.message(JSON.stringify(result, null, 2))
        return
      }

      const result = Object.fromEntries(
        LIST_CATEGORIES.map((entry) => [entry, listOrchestrationItems(process.cwd(), libraryDir, entry)]),
      )

      if (opts.json) {
        console.log(JSON.stringify(result, null, 2))
        return
      }

      p.log.message(JSON.stringify(result, null, 2))
    })

  orchestration
    .command('create <type> <name>')
    .description('Create an orchestration workflow, domain, or mode')
    .option('--description <description>', 'Artifact description')
    .option('--force', 'Overwrite files if they already exist')
    .option('--no-interactive', 'Disable interactive prompts')
    .option('--chain <chain>', 'Primary chain reference for workflow creation')
    .option('--team <team>', 'Optional review/synthesis team reference for workflow creation')
    .action(async (type: string, name: string, opts: OrchestrationCreateOptions) => {
      await createOrchestrationArtifact(toCreateType(type), name, opts)
    })

  orchestration
    .command('status')
    .description('Show orchestration scaffold status')
    .option('--json', 'Output as JSON')
    .action((opts: { json?: boolean }) => {
      const libraryDir = resolveLibraryDir(path.dirname(fileURLToPath(import.meta.url)))
      const status = getOrchestrationCounts(process.cwd(), libraryDir)

      if (opts.json) {
        console.log(JSON.stringify(status, null, 2))
        return
      }

      p.log.message(JSON.stringify(status, null, 2))
    })
}
