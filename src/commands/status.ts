import { join } from 'node:path'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import { Errors } from '../errors/index.js'
import { readStoreReadonly } from '../store/index.js'
import { fileExists, fileHash } from '../utils/files.js'

function summarizeFeatures(features?: Record<string, boolean>): string {
  if (!features) return '(defaults: all enabled)'

  const enabled = Object.entries(features)
    .filter(([, value]) => value)
    .map(([key]) => key)

  if (enabled.length === 0) return '(none enabled)'
  return enabled.join(', ')
}

function summarizeGitConventions(gitConventions?: {
  branchPattern?: string
  commitPattern?: string
  requireTicket?: boolean
  types?: string[]
}): string {
  if (!gitConventions) return '(defaults)'

  const branch = gitConventions.branchPattern ?? '{type}/{ticket}-{description}'
  const commit = gitConventions.commitPattern ?? '{type}({scope}): {description}'
  const requireTicket = gitConventions.requireTicket ?? false
  const typeCount = gitConventions.types?.length ?? 0

  return `branch=${branch} | commit=${commit} | requireTicket=${requireTicket} | types=${typeCount}`
}

export function registerStatus(program: Command): void {
  program
    .command('status')
    .description('Show current setup status')
    .action(async () => {
      const targetDir = process.cwd()
      const manifestPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(manifestPath)) {
        throw Errors.manifestNotFound(targetDir)
      }

      const store = await readStoreReadonly(targetDir)

      let healthy = 0
      let missing = 0
      let modified = 0

      for (const record of store.files) {
        const absPath = join(targetDir, record.path)

        if (!fileExists(absPath)) {
          missing += 1
          continue
        }

        if (fileHash(absPath) !== record.hash) {
          modified += 1
          continue
        }

        healthy += 1
      }

      const total = store.files.length
      const projectName = store.config.projectName || store.config.workspaceName || '(unnamed)'
      const tools = store.config.tools.length > 0 ? store.config.tools.join(', ') : '(none)'

      p.intro('ai-setup status')
      p.log.info(`Scope: ${store.config.setupScope}`)
      p.log.info(`Project name: ${projectName}`)
      p.log.info(`Tools: ${tools}`)
      p.log.info(`Planning dir: ${store.config.planningDir ?? '.planning'}`)
      p.log.info(`Use compiled root: ${store.config.useCompiledRoot ?? true}`)
      p.log.info(`Active features: ${summarizeFeatures(store.selections.features as Record<string, boolean> | undefined)}`)
      p.log.info(`Git conventions: ${summarizeGitConventions(store.selections.gitConventions)}`)
      p.log.info(`Files: total=${total}, healthy=${healthy}, missing=${missing}, modified=${modified}`)
      p.log.info(`Last init date: ${store.meta.installedAt}`)
      p.log.info(`CLI version: ${store.meta.cliVersion}`)
      p.outro('Done')
    })
}
