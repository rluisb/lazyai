import { join } from 'node:path'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'
import { Errors } from '../errors/index.js'
import { readStoreReadonly } from '../store/index.js'
import { fileExists, fileHash } from '../utils/files.js'
import { showSummaryBox } from '../utils/ui.js'

function formatFeatures(features?: Record<string, boolean>): string[] {
  if (!features) return ['(all defaults)']

  const enabled = Object.entries(features)
    .filter(([, value]) => value)
    .map(([key]) => key)

  if (enabled.length === 0) return ['(none)']
  return enabled
}

function formatHealthBar(healthy: number, total: number): string {
  const percent = total > 0 ? Math.round((healthy / total) * 100) : 100
  const barWidth = 20
  const filled = Math.round((healthy / total) * barWidth) || 0
  const empty = barWidth - filled

  const bar = pc.green('━'.repeat(filled)) + pc.dim('━'.repeat(empty))
  return `${bar} ${percent}%`
}

export function registerStatus(program: Command): void {
  program
    .command('status')
    .description('Show current setup status')
    .option('--json', 'Output as JSON')
    .action(async (opts: { json?: boolean }) => {
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
      const tools = store.config.tools.length > 0 ? store.config.tools : ['(none)']
      const features = formatFeatures(store.selections.features as Record<string, boolean> | undefined)

      // JSON output
      if (opts.json) {
        console.log(
          JSON.stringify(
            {
              scope: store.config.setupScope,
              projectName,
              tools: store.config.tools,
              planningDir: store.config.planningDir ?? '.planning',
              features,
              gitConventions: store.selections.gitConventions,
              files: { total, healthy, missing, modified },
              lastInit: store.meta.installedAt,
              cliVersion: store.meta.cliVersion,
            },
            null,
            2
          )
        )
        return
      }

      p.intro(pc.bold('ai-setup status'))

      // Project info box
      showSummaryBox('📦 Project', [
        { label: 'Name', value: projectName },
        { label: 'Scope', value: store.config.setupScope },
        { label: 'Planning dir', value: store.config.planningDir ?? '.planning' },
      ])

      // Tools box
      console.log('')
      showSummaryBox('🔧 Tools', [{ label: 'Installed', value: tools.join(', ') }])

      // Features box
      console.log('')
      showSummaryBox('⚡ Features', features.map((f, i) => ({ label: i === 0 ? 'Enabled' : '', value: f })))

      // Git conventions
      if (store.selections.gitConventions) {
        const gc = store.selections.gitConventions
        console.log('')
        showSummaryBox('📝 Git Conventions', [
          { label: 'Branch', value: gc.branchPattern ?? '{type}/{ticket}-{description}' },
          { label: 'Commit', value: gc.commitPattern ?? '{type}({scope}): {description}' },
          { label: 'Require ticket', value: gc.requireTicket ? 'Yes' : 'No' },
          { label: 'Types', value: gc.types?.join(', ') ?? '(defaults)' },
        ])
      }

      // File health box
      console.log('')
      const healthStatus =
        missing === 0 && modified === 0 ? pc.green('✓ All files healthy') : pc.yellow(`⚠ ${missing + modified} issues`)
      showSummaryBox('📁 File Health', [
        { label: 'Status', value: healthStatus },
        { label: 'Health', value: formatHealthBar(healthy, total) },
        { label: 'Total', value: `${total} managed files` },
        { label: 'Healthy', value: pc.green(`${healthy}`) },
        { label: 'Missing', value: missing > 0 ? pc.red(`${missing}`) : pc.dim('0') },
        { label: 'Modified', value: modified > 0 ? pc.yellow(`${modified}`) : pc.dim('0') },
      ])

      // Version info
      console.log('')
      showSummaryBox('ℹ️  Info', [
        { label: 'CLI version', value: store.meta.cliVersion },
        { label: 'Last init', value: new Date(store.meta.installedAt).toLocaleDateString() },
      ])

      if (missing > 0 || modified > 0) {
        p.outro(`Run ${pc.cyan('ai-setup doctor')} for details`)
      } else {
        p.outro(pc.green('Setup is healthy'))
      }
    })
}
