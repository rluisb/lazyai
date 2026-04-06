import * as p from '@clack/prompts'
import { Errors } from '../errors/index.js'
import type { PlannedFile } from './planner.js'

export async function runPhase4(opts: {
  interactive: boolean
  plan: PlannedFile[]
}): Promise<boolean> {
  // Non-interactive: always return true
  if (!opts.interactive) {
    return true
  }

  // Group files by category
  const groups = new Map<string, { newCount: number; updateCount: number }>()
  for (const file of opts.plan) {
    const existing = groups.get(file.category) ?? { newCount: 0, updateCount: 0 }
    if (file.isNew) {
      existing.newCount++
    } else {
      existing.updateCount++
    }
    groups.set(file.category, existing)
  }

  // Build summary lines
  const lines: string[] = []
  const totalNew = opts.plan.filter((f) => f.isNew).length
  const totalUpdate = opts.plan.filter((f) => !f.isNew).length

  // Category display names
  const categoryNames: Record<string, string> = {
    constitution: 'Constitution files',
    specs: 'Specs dirs',
    'specs-agents': 'Specs AGENTS.md files',
    templates: 'Templates',
    rules: 'Rules',
    infra: 'Infrastructure',
    root: 'Root config files',
    agent: 'Agent definitions',
    skill: 'Skills',
    prompt: 'Prompt templates',
  }

  for (const [category, counts] of groups) {
    const name = categoryNames[category] ?? category
    const parts: string[] = []
    if (counts.newCount > 0) parts.push(`${counts.newCount} new`)
    if (counts.updateCount > 0) parts.push(`${counts.updateCount} existing`)
    lines.push(`  ${name}: ${parts.join(', ')}`)
  }

  lines.unshift(`Total: ${totalNew} new files, ${totalUpdate} updates`)
  lines.push('')

  // Show summary
  p.note(lines.join('\n'), 'Installation Plan')

  // Confirm
  const confirmed = await p.confirm({
    message: 'Proceed with installation?',
    initialValue: true,
  })

  if (p.isCancel(confirmed)) {
    p.cancel('Setup cancelled.')
    throw Errors.userCancelled()
  }

  return confirmed as boolean
}
