import type { Command } from 'commander'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import * as p from '@clack/prompts'
import type { AiSetupConfig } from '../types.js'
import { fileExists, fileHash } from '../utils/files.js'

export function registerDoctor(program: Command): void {
  program
    .command('doctor')
    .description('Verify setup integrity against .ai-setup.json')
    .action(() => {
      const targetDir = process.cwd()
      const configPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(configPath)) {
        p.log.error('No .ai-setup.json found. Please run init first.')
        process.exit(1)
      }

      const config = JSON.parse(readFileSync(configPath, 'utf-8')) as AiSetupConfig

      const missing: string[] = []
      const modified: string[] = []
      let healthy = 0

      p.intro('Running ai-setup doctor...')

      for (const record of config.files) {
        const absPath = join(targetDir, record.path)
        if (!fileExists(absPath)) {
          missing.push(record.path)
          continue
        }

        const currentHash = fileHash(absPath)
        if (currentHash !== record.hash) {
          modified.push(record.path)
          continue
        }

        healthy += 1
      }

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
        process.exit(1)
      }

      p.outro('✅ Setup integrity is healthy')
    })
}
