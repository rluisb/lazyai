import path from 'node:path'
import * as p from '@clack/prompts'
import pc from 'picocolors'
import { detectAdapters, detectExistingSetup, formatPlan, importSetup } from '../migration/index.js'
import type { MergeStrategy, MigrationResult } from '../migration/types.js'
import { DETECTION_NAMES } from '../migration/detector.js'

export interface MigrationCommandOptions {
  path?: string
  preview?: boolean
  strategy?: string
  verbose?: boolean
  skipBackup?: boolean
  interactive?: boolean
  yes?: boolean
  canonical?: boolean
}

const ADAPTER_LABELS: Record<string, string> = {
  opencode: 'OpenCode',
  'claude-code': 'Claude Code',
  gemini: 'Gemini CLI',
  copilot: 'GitHub Copilot',
}

export const VALID_MERGE_STRATEGIES: MergeStrategy[] = ['smart', 'preserve', 'replace', 'append']
export const MIGRATION_MARKER_HINT = 'Expected markers include: AGENTS.md, CLAUDE.md, GEMINI.md, .opencode/, .claude/, .gemini/, or .github/copilot-instructions.md'

const STRATEGY_DESCRIPTIONS: Record<MergeStrategy, string> = {
  smart: 'attempt a 3-way merge and stop when manual review is required',
  preserve: 'keep existing files when there is overlap',
  replace: 'overwrite with ai-setup-managed files and create backups first',
  append: 'combine content where supported by the parser',
}

export async function runMigrationCommand(
  commandName: 'import' | 'migrate',
  sourcePath: string | undefined,
  options: MigrationCommandOptions,
): Promise<void> {
  const targetPath = process.cwd()
  const resolvedSource = sourcePath ? path.resolve(sourcePath) : targetPath

  p.intro(pc.blue(`ai-setup ${commandName}`))

  const strategy = normalizeStrategy(options.strategy)
  if (!strategy) {
    p.cancel(`Unknown merge strategy "${options.strategy}".`)
    printStrategyHelp(commandName)
    process.exit(1)
  }

  const spinner = p.spinner()
  spinner.start('Scanning for supported AI setup files...')

  const adapters = await detectAdapters(resolvedSource)

  // Also detect observed-only tools for informational message
  const context = { sourcePath: resolvedSource, targetPath: process.cwd(), options: { preview: false, mergeStrategy: 'smart' as const, verbose: false, skipBackup: false, interactive: false } }
  const allDetections = await detectExistingSetup(context)
  const observedTools = allDetections
    .filter(d => d.metadata?.observed && d.confidence > 0.3)
    .map(d => DETECTION_NAMES[d.adapterId] || d.adapterId)

  if (adapters.length === 0 && observedTools.length > 0) {
    spinner.stop(pc.yellow('No supported AI setup detected'))
    console.log(pc.gray(`👁️  Detected (not managed): ${observedTools.join(', ')}`))
    printNoSetupHelp(commandName, resolvedSource)
    process.exit(1)
  }

  if (adapters.length === 0) {
    spinner.stop(pc.yellow('No supported AI setup detected'))
    printNoSetupHelp(commandName, resolvedSource)
    process.exit(1)
  }

  if (observedTools.length > 0) {
    console.log(pc.gray(`👁️  Detected (not managed): ${observedTools.join(', ')}`))
  }

  spinner.stop(pc.green(`Detected ${adapters.length} setup(s): ${formatAdapterList(adapters)}`))

  spinner.start('Building migration plan...')

  const analysisResult = await importSetup({
    path: resolvedSource,
    preview: true,
    canonicalOutput: options.canonical !== false,
    mergeStrategy: strategy,
    ...(options.verbose !== undefined ? { verbose: options.verbose } : {}),
    ...(options.skipBackup !== undefined ? { skipBackup: options.skipBackup } : {}),
    ...(options.interactive !== undefined ? { interactive: options.interactive } : {}),
  })

  spinner.stop(pc.green('Migration plan ready'))

  if (analysisResult.plan) {
    console.log(`\n${formatPlan(analysisResult.plan)}`)
  }

  if (options.preview) {
    console.log(pc.blue('\nPreview only — no files were changed.'))
    console.log(pc.gray(`Run ai-setup ${commandName} without --preview to apply this plan.`))
    return
  }

  if (analysisResult.plan && !analysisResult.plan.canProceed && !options.interactive) {
    printBlockedHelp(commandName, analysisResult.plan.conflicts.filter((conflict) => !conflict.resolved).length)
    process.exit(1)
  }

  if (!options.yes) {
    const confirm = await p.confirm({
      message: `Proceed with ${commandName === 'import' ? 'importing' : 'migrating'} this setup?`,
      initialValue: false,
    })

    if (!confirm) {
      p.cancel('Migration cancelled')
      process.exit(0)
    }
  }

  spinner.start('Applying migration plan...')

  const executeResult = await importSetup({
    path: resolvedSource,
    preview: false,
    canonicalOutput: options.canonical !== false,
    mergeStrategy: strategy,
    ...(options.verbose !== undefined ? { verbose: options.verbose } : {}),
    ...(options.skipBackup !== undefined ? { skipBackup: options.skipBackup } : {}),
    ...(options.interactive !== undefined ? { interactive: options.interactive } : {}),
  })

  spinner.stop(executeResult.success ? pc.green('Migration complete') : pc.red('Migration finished with issues'))
  printMigrationResult(executeResult)

  if (!executeResult.success) {
    process.exit(1)
  }

  p.outro('Done!')
}

function normalizeStrategy(value: string | undefined): MergeStrategy | null {
  const strategy = (value || 'smart').toLowerCase()
  return VALID_MERGE_STRATEGIES.includes(strategy as MergeStrategy)
    ? (strategy as MergeStrategy)
    : null
}

export function formatAdapterList(adapters: string[]): string {
  return adapters.map((adapter) => ADAPTER_LABELS[adapter] || adapter).join(', ')
}

function printStrategyHelp(commandName: 'import' | 'migrate'): void {
  console.log(pc.gray('\nSupported strategies:'))
  for (const strategy of VALID_MERGE_STRATEGIES) {
    console.log(`  • ${strategy} — ${STRATEGY_DESCRIPTIONS[strategy]}`)
  }
  console.log(pc.gray(`\nExample: ai-setup ${commandName} --strategy preserve`))
}

function printNoSetupHelp(commandName: 'import' | 'migrate', resolvedSource: string): void {
  console.log(pc.gray(`Scanned: ${resolvedSource}`))
  console.log(pc.gray(MIGRATION_MARKER_HINT))
  console.log(pc.gray(`Try again with an explicit path: ai-setup ${commandName} /path/to/project`))
}

function printBlockedHelp(commandName: 'import' | 'migrate', unresolvedConflicts: number): void {
  console.log(pc.red('\n❌ Migration needs attention before it can continue.'))
  console.log(pc.yellow(`Unresolved conflicts: ${unresolvedConflicts}`))
  console.log(pc.blue('\nTry one of these next steps:'))
  console.log(`  1. ai-setup ${commandName} --interactive`)
  console.log(`  2. ai-setup ${commandName} --strategy preserve`)
  console.log(`  3. ai-setup ${commandName} --strategy replace`)
}

function printMigrationResult(result: MigrationResult): void {
  if (result.success) {
    console.log(pc.green('\n✅ Migration successful!'))
    console.log(pc.gray('\nSummary:'))
    console.log(`  Files created: ${result.stats.filesCreated}`)
    console.log(`  Files modified: ${result.stats.filesModified}`)
    console.log(`  Files backed up: ${result.stats.filesBackedUp}`)
    console.log(`  Files skipped: ${result.stats.filesSkipped}`)

    if (result.stats.conflictsUnresolved > 0) {
      console.log(pc.yellow(`  Unresolved conflicts: ${result.stats.conflictsUnresolved}`))
    }

    if (result.warnings.length > 0) {
      console.log(pc.yellow('\nWarnings:'))
      for (const warning of result.warnings) {
        console.log(`  • ${warning}`)
      }
    }

    if (result.backupPath) {
      console.log(pc.gray(`\nBackup created at: ${result.backupPath}`))
    }

    console.log(pc.blue('\nRecommended next steps:'))
    console.log('  1. Review the generated and updated files')
    console.log('  2. Run ai-setup doctor --migration-check')
    console.log('  3. Commit the migration once everything looks correct')
    return
  }

  console.log(pc.red('\n❌ Migration failed.'))

  if (result.errors.length > 0) {
    console.log(pc.red('\nErrors:'))
    for (const error of result.errors) {
      console.log(`  • ${error}`)
    }
  }

  if (result.warnings.length > 0) {
    console.log(pc.yellow('\nWarnings:'))
    for (const warning of result.warnings) {
      console.log(`  • ${warning}`)
    }
  }

  if (result.backupPath) {
    console.log(pc.yellow(`\nBackup available at: ${result.backupPath}`))
  }

  console.log(pc.blue('\nSuggested recovery steps:'))
  console.log('  1. Re-run with --preview to inspect the plan')
  console.log('  2. Re-run with --interactive to resolve conflicts manually')
  console.log('  3. Choose --strategy preserve or --strategy replace for a simpler merge path')
}
