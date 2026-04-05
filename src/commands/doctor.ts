
import { join } from 'node:path'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'
import { Errors } from '../errors/index.js'
import { migrationCheck } from '../migration/index.js'
import type { DriftCheckResult } from '../migration/types.js'
import { readStore, writeStore } from '../store/index.js'
import { fileExists, fileHash } from '../utils/files.js'

export function registerDoctor(program: Command): void {
  program
    .command('doctor')
    .description('Verify setup integrity against .ai-setup.json')
    .option('--migration-check', 'Check for drift between current setup and clean ai-setup state')
    .option('--verbose', 'Show detailed output')
    .action(async (opts: { migrationCheck?: boolean; verbose?: boolean }) => {
      if (opts.migrationCheck) {
        await runMigrationCheck(opts.verbose)
      } else {
        await runIntegrityCheck()
      }
    })
}

async function runIntegrityCheck(): Promise<void> {
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
}

async function runMigrationCheck(verbose?: boolean): Promise<void> {
  const targetDir = process.cwd()

  p.intro('Running migration check...')

  const spinner = p.spinner()
  spinner.start('Analyzing current setup...')

  try {
    const result: DriftCheckResult = await migrationCheck({
      path: targetDir,
      ...(verbose !== undefined ? { verbose } : {}),
    })

    spinner.stop('Analysis complete')

    if (result.clean) {
      console.log(pc.green('\n✅ No drift detected'))
      console.log(pc.gray('Current setup matches clean ai-setup state'))
      p.outro('Done')
      return
    }

    // Show drift information
    console.log(pc.yellow('\n⚠️  Drift detected between current setup and clean ai-setup'))
    console.log('')

    if (result.missingFiles.length > 0) {
      console.log(pc.red(`Missing files (${result.missingFiles.length}):`))
      for (const file of result.missingFiles) {
        console.log(`  - ${file}`)
      }
      console.log('')
    }

    if (result.extraFiles.length > 0) {
      console.log(pc.blue(`Extra files not in ai-setup (${result.extraFiles.length}):`))
      if (verbose) {
        for (const file of result.extraFiles) {
          console.log(`  + ${file}`)
        }
      } else {
        console.log(`  + ${result.extraFiles.length} files (use --verbose to see all)`)
      }
      console.log('')
    }

    if (result.modifiedFiles.length > 0) {
      console.log(pc.yellow(`Modified files (${result.modifiedFiles.length}):`))
      for (const file of result.modifiedFiles) {
        console.log(`  ~ ${file.path}`)
        if (verbose && file.difference) {
          console.log(pc.gray(`    Difference: ${file.difference.substring(0, 100)}...`))
        }
      }
      console.log('')
    }

    console.log(pc.blue('Recommendations:'))
    console.log('  1. Run ai-setup update to refresh managed files')
    console.log('  2. Review extra files and manually integrate if needed')
    console.log('  3. Run ai-setup doctor --migration-check after updates')

    p.outro('Done')
    throw Errors.unknown('Migration check found drift')

  } catch (error) {
    spinner.stop('Check failed')
    if (error && typeof error === 'object' && 'code' in error) {
      throw error
    }
    throw Errors.unknown(String(error))
  }
}
