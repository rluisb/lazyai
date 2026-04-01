import type { Command } from 'commander'
import { join } from 'node:path'
import * as p from '@clack/prompts'
import { fileExists, fileHash } from '../utils/files.js'
import { readStore, writeStore } from '../store/index.js'
import { Errors } from '../errors/index.js'

export function registerDoctor(program: Command): void {
  program
    .command('doctor')
    .description('Verify setup integrity against .ai-setup.json')
    .action(async () => {
      const targetDir = process.cwd()
      const configPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(configPath)) {
        throw Errors.manifestNotFound(targetDir)
      }

      const storeData = await readStore(targetDir)

      const missing: string[] = []
      const modified: string[] = []
      let healthy = 0

      p.intro('Running ai-setup doctor...')

      const checkedAt = new Date().toISOString()

      for (const record of storeData.files) {
        const absPath = join(targetDir, record.path)

        if (!fileExists(absPath)) {
          missing.push(record.path)
          record.status = 'missing'
          record.lastCheckedAt = checkedAt
          continue
        }

        const currentHash = fileHash(absPath)
        if (currentHash !== record.hash) {
          modified.push(record.path)
          record.status = 'modified'
          record.lastCheckedAt = checkedAt
          continue
        }

        healthy += 1
        record.status = 'installed'
        record.lastCheckedAt = checkedAt
      }

      await writeStore(targetDir, storeData)

      p.log.info(`Healthy: ${healthy}`)
      p.log.info(`Missing: ${missing.length}`)
      p.log.info(`Modified: ${modified.length}`)

      if (missing.length > 0) {
        p.log.warn('Missing files:')
        for (const file of missing) {
          p.log.message(`  - ${file}`)
        }
      }

      if (modified.length > 0) {
        p.log.warn('Modified files:')
        for (const file of modified) {
          p.log.message(`  - ${file}`)
        }
      }

      if (missing.length > 0 || modified.length > 0) {
        p.outro('❌ Setup integrity issues found')
        const issueCount = missing.length + modified.length
        throw Errors.unknown(`Doctor found ${issueCount} integrity issue(s)`) 
      }

      p.outro('✅ Setup integrity is healthy')
    })
}
