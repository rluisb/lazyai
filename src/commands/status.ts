import type { Command } from 'commander'
import { join } from 'node:path'
import * as p from '@clack/prompts'
import { readStoreReadonly } from '../store/index.js'
import { Errors } from '../errors/index.js'
import { fileExists, fileHash } from '../utils/files.js'

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
      p.log.info(`Files: total=${total}, healthy=${healthy}, missing=${missing}, modified=${modified}`)
      p.log.info(`Last init date: ${store.meta.installedAt}`)
      p.log.info(`CLI version: ${store.meta.cliVersion}`)
      p.outro('Done')
    })
}
