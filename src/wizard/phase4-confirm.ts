import * as p from '@clack/prompts'
import { Errors } from '../errors/index.js'
import type { SetupScope, ToolId, WizardConfig } from '../types.js'
import { GO_BACK, type SummaryItem, showSummaryBox } from '../utils/ui.js'
import type { PlannedFile } from './planner.js'

function formatTools(tools: ToolId[]): string {
  const toolNames: Record<ToolId, string> = {
    opencode: 'OpenCode',
    'claude-code': 'Claude Code',
    gemini: 'Gemini CLI',
    copilot: 'GitHub Copilot',
    codex: 'Codex',
  }
  return tools.map((t) => toolNames[t] || t).join(', ')
}

function formatScope(scope: SetupScope): string {
  const scopeNames: Record<SetupScope, string> = {
    global: 'Global (~/.ai/)',
    workspace: 'Workspace (multi-repo)',
    project: 'Project (single repo)',
  }
  return scopeNames[scope] || scope
}

export async function runPhase4(opts: {
  interactive: boolean
  plan: PlannedFile[]
  config?: WizardConfig
}): Promise<boolean | typeof GO_BACK> {
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

  // Calculate totals
  const totalNew = opts.plan.filter((f) => f.isNew).length
  const totalUpdate = opts.plan.filter((f) => !f.isNew).length

  // Build summary items
  const summaryItems: SummaryItem[] = []

  if (opts.config) {
    summaryItems.push({ label: 'Scope', value: formatScope(opts.config.setupScope) })
    summaryItems.push({ label: 'Tools', value: formatTools(opts.config.tools) })
    summaryItems.push({ label: 'Project', value: opts.config.projectName })
  }

  summaryItems.push({ label: 'Files', value: `${totalNew} new, ${totalUpdate} updates` })

  // Show summary box
  showSummaryBox('📦 Setup Summary', summaryItems)

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
    mcp: 'MCP configuration',
  }

  // Build detail lines
  const lines: string[] = []
  for (const [category, counts] of groups) {
    const name = categoryNames[category] ?? category
    const parts: string[] = []
    if (counts.newCount > 0) parts.push(`${counts.newCount} new`)
    if (counts.updateCount > 0) parts.push(`${counts.updateCount} existing`)
    lines.push(`  ${name}: ${parts.join(', ')}`)
  }

  if (lines.length > 0) {
    p.note(lines.join('\n'), 'File Breakdown')
  }

  // Confirm — use select instead of confirm to support Back navigation
  const confirmedResult = await p.select({
    message: 'Proceed with installation?',
    options: [
      { value: 'yes', label: 'Yes, install', hint: 'Continue with the setup' },
      { value: 'no', label: 'No, cancel', hint: 'Cancel the setup' },
      { value: GO_BACK, label: '↩ Back', hint: 'Go back to conflict resolution' },
    ],
    initialValue: 'yes',
  })

  if (p.isCancel(confirmedResult)) {
    p.cancel('Setup cancelled.')
    throw Errors.userCancelled()
  }

  if (confirmedResult === GO_BACK) {
    return GO_BACK
  }

  return confirmedResult === 'yes'
}
