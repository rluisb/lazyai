import type { Command } from 'commander'
import { join } from 'node:path'
import { readStore } from '../store/index.js'
import { Errors } from '../errors/index.js'
import { fileExists } from '../utils/files.js'

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

      const storeData = await readStore(targetDir)
      console.log(`Setup scope: ${storeData.config.setupScope}`)
      console.log(`Tools: ${storeData.config.tools.join(', ')}`)
      console.log(`Managed files: ${storeData.files.length}`)
    })
}
