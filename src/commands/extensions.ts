import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'
import { discoverExtensions } from '../extensions/discovery.js'
import { showSummaryBox } from '../utils/ui.js'

export function registerExtensions(program: Command): void {
  program
    .command('extensions')
    .alias('ext')
    .description('List discovered ai-setup extensions')
    .option('--json', 'Output as JSON')
    .action(async (opts: { json?: boolean }) => {
      const targetDir = process.cwd()
      const extensions = discoverExtensions(targetDir)

      if (opts.json) {
        console.log(JSON.stringify(extensions, null, 2))
        return
      }

      if (extensions.length === 0) {
        p.intro(pc.bold('ai-setup extensions'))
        p.log.info('No extensions discovered.')
        p.note(
          [
            'Add extensions via .ai-setup.toml:',
            '  [extensions.my-ext]',
            '  path = "./path/to/extension"',
            '',
            'Or place content in .ai/extensions/<name>/',
          ].join('\n'),
          'How to add extensions',
        )
        return
      }

      p.intro(pc.bold('ai-setup extensions'))

      const totalAgents = extensions.reduce((sum, e) => sum + e.content.agents.length, 0)
      const totalSkills = extensions.reduce((sum, e) => sum + e.content.skills.length, 0)

      for (const ext of extensions) {
        const items = []
        if (ext.content.agents.length) items.push(`${ext.content.agents.length} agents`)
        if (ext.content.skills.length) items.push(`${ext.content.skills.length} skills`)
        if (ext.content.prompts.length) items.push(`${ext.content.prompts.length} prompts`)
        if (ext.content.rules.length) items.push(`${ext.content.rules.length} rules`)

        showSummaryBox(`📦 ${ext.name}`, [
          { label: 'Kind', value: ext.kind === 'toml' ? 'TOML config' : 'Local directory' },
          { label: 'Path', value: ext.path },
          { label: 'Content', value: items.join(', ') || '(empty)' },
        ])
      }

      console.log('')
      p.log.info(`${extensions.length} extension(s), ${totalAgents} agent(s), ${totalSkills} skill(s)`)
    })
}
