import { join } from 'node:path'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'
import { Errors } from '../errors/index.js'
import { migrationCheck } from '../migration/index.js'
import type { DriftCheckResult } from '../migration/types.js'
import { readStore, writeStore } from '../store/index.js'
import { fileExists, fileHash } from '../utils/files.js'
import { compareLibrarySkills } from '../utils/library-compare.js'
import { showSummaryBox } from '../utils/ui.js'

function formatHealthBar(healthy: number, total: number): string {
  const percent = total > 0 ? Math.round((healthy / total) * 100) : 100
  const barWidth = 30
  const filled = Math.round((healthy / total) * barWidth) || 0
  const empty = barWidth - filled

  let color = pc.green
  if (percent < 90) color = pc.yellow
  if (percent < 70) color = pc.red

  const bar = color('━'.repeat(filled)) + pc.dim('━'.repeat(empty))
  return `${bar} ${percent}%`
}

export function registerDoctor(program: Command): void {
  program
    .command('doctor')
    .description('Verify setup integrity against .ai-setup.json')
    .option('--migration-check', 'Check for drift between current setup and clean ai-setup state')
    .option('--skills-check', 'Check installed skills against current library versions')
    .option('--verbose', 'Show detailed output')
    .option('--json', 'Output as JSON')
    .action(async (opts: { migrationCheck?: boolean; skillsCheck?: boolean; verbose?: boolean; json?: boolean }) => {
      if (opts.migrationCheck) {
        await runMigrationCheck(opts.verbose)
      } else if (opts.skillsCheck) {
        await runSkillsCheck(opts)
      } else {
        await runIntegrityCheck(opts)
      }
    })
}

async function runIntegrityCheck(opts: { verbose?: boolean; json?: boolean }): Promise<void> {
  const targetDir = process.cwd()
  const configPath = join(targetDir, '.ai-setup.json')

  if (!fileExists(configPath)) {
    throw Errors.manifestNotFound(targetDir)
  }

  const storeData = await readStore(targetDir)

  const missing: string[] = []
  const modified: string[] = []
  let healthy = 0

  if (!opts.json) {
    p.intro(pc.bold('ai-setup doctor'))
  }

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

  const total = storeData.files.length
  const issues = missing.length + modified.length
  const isHealthy = issues === 0

  // JSON output
  if (opts.json) {
    console.log(
      JSON.stringify(
        {
          healthy: isHealthy,
          files: { total, healthy, missing: missing.length, modified: modified.length },
          missingFiles: missing,
          modifiedFiles: modified,
          checkedAt,
        },
        null,
        2
      )
    )
    if (!isHealthy) {
      throw Errors.unknown(`Doctor found ${issues} integrity issue(s)`)
    }
    return
  }

  // Summary box
  const statusEmoji = isHealthy ? '✅' : issues < 5 ? '⚠️' : '❌'
  const statusText = isHealthy ? pc.green('All files healthy') : pc.yellow(`${issues} issue(s) found`)

  showSummaryBox(`${statusEmoji} Integrity Check`, [
    { label: 'Status', value: statusText },
    { label: 'Health', value: formatHealthBar(healthy, total) },
    { label: 'Total files', value: `${total}` },
    { label: 'Healthy', value: pc.green(`${healthy}`) },
    { label: 'Missing', value: missing.length > 0 ? pc.red(`${missing.length}`) : pc.dim('0') },
    { label: 'Modified', value: modified.length > 0 ? pc.yellow(`${modified.length}`) : pc.dim('0') },
  ])

  // Show missing files
  if (missing.length > 0) {
    console.log('')
    p.log.warn(pc.red(`Missing files (${missing.length}):`))
    const displayMissing = opts.verbose ? missing : missing.slice(0, 5)
    for (const file of displayMissing) {
      p.log.message(`  ${pc.red('✗')} ${file}`)
    }
    if (!opts.verbose && missing.length > 5) {
      p.log.message(pc.dim(`  ... and ${missing.length - 5} more (use --verbose to see all)`))
    }
  }

  // Show modified files
  if (modified.length > 0) {
    console.log('')
    p.log.warn(pc.yellow(`Modified files (${modified.length}):`))
    const displayModified = opts.verbose ? modified : modified.slice(0, 5)
    for (const file of displayModified) {
      p.log.message(`  ${pc.yellow('~')} ${file}`)
    }
    if (!opts.verbose && modified.length > 5) {
      p.log.message(pc.dim(`  ... and ${modified.length - 5} more (use --verbose to see all)`))
    }
  }

  // Recommendations
  if (!isHealthy) {
    console.log('')
    showSummaryBox('💡 Recommendations', [
      { label: '1', value: `Run ${pc.cyan('ai-setup update')} to restore missing files` },
      { label: '2', value: `Run ${pc.cyan('ai-setup update --force')} to reset modified files` },
      { label: '3', value: `Run ${pc.cyan('ai-setup compile')} to regenerate tool files` },
    ])
  }

  if (isHealthy) {
    p.outro(pc.green('✓ Setup integrity verified'))
  } else {
    p.outro(pc.yellow('⚠ Setup has integrity issues'))
    throw Errors.unknown(`Doctor found ${issues} integrity issue(s)`)
  }
}

async function runMigrationCheck(verbose?: boolean): Promise<void> {
  const targetDir = process.cwd()

  p.intro(pc.bold('ai-setup migration check'))

  const spinner = p.spinner()
  spinner.start('Analyzing current setup...')

  try {
    const result: DriftCheckResult = await migrationCheck({
      path: targetDir,
      ...(verbose !== undefined ? { verbose } : {}),
    })

    spinner.stop('Analysis complete')

    if (result.clean) {
      showSummaryBox('✅ No Drift Detected', [
        { label: 'Status', value: pc.green('Current setup matches clean ai-setup state') },
      ])
      p.outro(pc.green('Migration check passed'))
      return
    }

    // Show drift summary
    const totalDrift = result.missingFiles.length + result.extraFiles.length + result.modifiedFiles.length
    showSummaryBox('⚠️  Drift Detected', [
      { label: 'Total changes', value: pc.yellow(`${totalDrift}`) },
      { label: 'Missing', value: result.missingFiles.length > 0 ? pc.red(`${result.missingFiles.length}`) : pc.dim('0') },
      { label: 'Extra', value: result.extraFiles.length > 0 ? pc.blue(`${result.extraFiles.length}`) : pc.dim('0') },
      { label: 'Modified', value: result.modifiedFiles.length > 0 ? pc.yellow(`${result.modifiedFiles.length}`) : pc.dim('0') },
    ])

    if (result.missingFiles.length > 0) {
      console.log('')
      p.log.error(pc.red(`Missing files (${result.missingFiles.length}):`))
      const display = verbose ? result.missingFiles : result.missingFiles.slice(0, 5)
      for (const file of display) {
        p.log.message(`  ${pc.red('✗')} ${file}`)
      }
      if (!verbose && result.missingFiles.length > 5) {
        p.log.message(pc.dim(`  ... and ${result.missingFiles.length - 5} more`))
      }
    }

    if (result.extraFiles.length > 0) {
      console.log('')
      p.log.info(pc.blue(`Extra files (${result.extraFiles.length}):`))
      if (verbose) {
        for (const file of result.extraFiles) {
          p.log.message(`  ${pc.blue('+')} ${file}`)
        }
      } else {
        p.log.message(pc.dim(`  ${result.extraFiles.length} files (use --verbose to see all)`))
      }
    }

    if (result.modifiedFiles.length > 0) {
      console.log('')
      p.log.warn(pc.yellow(`Modified files (${result.modifiedFiles.length}):`))
      const display = verbose ? result.modifiedFiles : result.modifiedFiles.slice(0, 5)
      for (const file of display) {
        p.log.message(`  ${pc.yellow('~')} ${file.path}`)
        if (verbose && file.difference) {
          p.log.message(pc.dim(`    ${file.difference.substring(0, 100)}...`))
        }
      }
      if (!verbose && result.modifiedFiles.length > 5) {
        p.log.message(pc.dim(`  ... and ${result.modifiedFiles.length - 5} more`))
      }
    }

    console.log('')
    showSummaryBox('💡 Recommendations', [
      { label: '1', value: `Run ${pc.cyan('ai-setup update')} to refresh managed files` },
      { label: '2', value: 'Review extra files and manually integrate if needed' },
      { label: '3', value: `Run ${pc.cyan('ai-setup doctor --migration-check')} after updates` },
    ])

    p.outro(pc.yellow('Migration check found drift'))
    throw Errors.unknown('Migration check found drift')
  } catch (error) {
    if (error && typeof error === 'object' && 'code' in error) {
      throw error
    }
    spinner.stop('Check failed')
    throw Errors.unknown(String(error))
  }
}

async function runSkillsCheck(opts: { verbose?: boolean; json?: boolean }): Promise<void> {
  const targetDir = process.cwd()
  const verbose = opts.verbose ?? false

  if (!opts.json) {
    p.intro(pc.bold('ai-setup skills check'))
  }

  const results = await compareLibrarySkills(targetDir)

  if (results.length === 0) {
    p.log.info('No library skills found.')
    p.outro(pc.dim('Run ai-setup init first to install skills.'))
    return
  }

  const current = results.filter((r) => r.status === 'current')
  const drifted = results.filter((r) => r.status === 'drifted')
  const modified = results.filter((r) => r.status === 'modified')
  const missing = results.filter((r) => r.status === 'missing')
  const issues = drifted.length + modified.length + missing.length
  const isHealthy = issues === 0

  // JSON output
  if (opts.json) {
    console.log(JSON.stringify({
      healthy: isHealthy,
      total: results.length,
      current: current.length,
      drifted: drifted.length,
      modified: modified.length,
      missing: missing.length,
      skills: results.map((r) => ({
        name: r.name,
        description: r.description,
        status: r.status,
        locations: r.installedLocations,
      })),
    }, null, 2))
    if (!isHealthy) {
      throw Errors.unknown(`Skills check found ${issues} issue(s)`)
    }
    return
  }

  // Summary
  const statusEmoji = isHealthy ? '✅' : '⚠️'
  const statusText = isHealthy ? pc.green('All skills current') : pc.yellow(`${issues} issue(s) found`)

  showSummaryBox(`${statusEmoji} Skills Check`, [
    { label: 'Status', value: statusText },
    { label: 'Total skills', value: `${results.length}` },
    { label: 'Current', value: pc.green(`${current.length}`) },
    { label: 'Drifted', value: drifted.length > 0 ? pc.yellow(`${drifted.length}`) : pc.dim('0') },
    { label: 'Modified', value: modified.length > 0 ? pc.yellow(`${modified.length}`) : pc.dim('0') },
    { label: 'Missing', value: missing.length > 0 ? pc.red(`${missing.length}`) : pc.dim('0') },
  ])

  // Show drifted skills
  if (drifted.length > 0) {
    console.log('')
    p.log.warn(pc.yellow(`Drifted skills (${drifted.length}): library updated, installed is stale`))
    for (const skill of drifted.slice(0, verbose ? undefined : 5)) {
      p.log.message(`  ${pc.yellow('↻')} ${skill.name}`)
      if (opts.verbose && skill.installedLocations.length > 0) {
        for (const loc of skill.installedLocations) {
          p.log.message(pc.dim(`    ${loc.path}`))
        }
      }
    }
    if (!opts.verbose && drifted.length > 5) {
      p.log.message(pc.dim(`  ... and ${drifted.length - 5} more (use --verbose for details)`))
    }
  }

  // Show modified skills
  if (modified.length > 0) {
    console.log('')
    p.log.warn(pc.yellow(`Modified skills (${modified.length}): user-changed content`))
    for (const skill of modified.slice(0, verbose ? undefined : 5)) {
      p.log.message(`  ${pc.yellow('~')} ${skill.name}`)
      if (opts.verbose && skill.installedLocations.length > 0) {
        for (const loc of skill.installedLocations) {
          p.log.message(pc.dim(`    ${loc.path}`))
        }
      }
    }
    if (!opts.verbose && modified.length > 5) {
      p.log.message(pc.dim(`  ... and ${modified.length - 5} more (use --verbose for details)`))
    }
  }

  // Show missing skills
  if (missing.length > 0) {
    console.log('')
    p.log.warn(pc.red(`Missing skills (${missing.length}): not installed`))
    for (const skill of missing.slice(0, verbose ? undefined : 5)) {
      p.log.message(`  ${pc.red('✗')} ${skill.name}`)
    }
    if (!opts.verbose && missing.length > 5) {
      p.log.message(pc.dim(`  ... and ${missing.length - 5} more (use --verbose for details)`))
    }
  }

  // Recommendations
  if (!isHealthy) {
    console.log('')
    showSummaryBox('💡 Recommendations', [
      { label: '1', value: `Run ${pc.cyan('ai-setup update --force')} to refresh drifted/modified skills` },
      { label: '2', value: `Run ${pc.cyan('ai-setup compile')} to reinstall missing skills` },
      { label: '3', value: `Run ${pc.cyan('ai-setup update --check')} to preview changes first` },
    ])
  }

  if (isHealthy) {
    p.outro(pc.green('✓ All skills up-to-date'))
  } else {
    p.outro(pc.yellow('⚠ Skills have issues'))
    throw Errors.unknown(`Skills check found ${issues} issue(s)`)
  }
}
