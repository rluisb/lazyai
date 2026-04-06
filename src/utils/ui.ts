/**
 * UI Utilities
 *
 * Shared UI helpers for consistent CLI presentation using @clack/prompts.
 */

import * as p from '@clack/prompts'

/** Phase information for progress display */
export interface PhaseInfo {
  current: number
  total: number
  name: string
}

/**
 * Display a phase progress indicator
 *
 * Example output:
 * ◇  Phase 1 of 4: Setup Context
 * │  ━━━━━━━━━━○○○○○○○○○○  25%
 */
export function showPhaseProgress(phase: PhaseInfo): void {
  const percent = Math.round((phase.current / phase.total) * 100)
  const barWidth = 20
  const filled = Math.round((phase.current / phase.total) * barWidth)
  const empty = barWidth - filled

  const bar = '━'.repeat(filled) + '○'.repeat(empty)

  p.log.step(`Phase ${phase.current} of ${phase.total}: ${phase.name}`)
  p.log.message(`${bar}  ${percent}%`)
}

/** Summary item for display */
export interface SummaryItem {
  label: string
  value: string
}

/**
 * Display a summary box
 *
 * Example output:
 * ┌─────────────────────────────────┐
 * │  📦 Setup Summary               │
 * ├─────────────────────────────────┤
 * │  Scope:    Project              │
 * │  Tools:    OpenCode, Claude     │
 * │  Files:    23 new, 2 modified   │
 * └─────────────────────────────────┘
 */
export function showSummaryBox(title: string, items: SummaryItem[]): void {
  const maxLabelLen = Math.max(...items.map((i) => i.label.length))
  const lines = items.map((item) => `${item.label.padEnd(maxLabelLen)}  ${item.value}`)

  const content = lines.join('\n')
  p.note(content, title)
}

/** Task status for display */
export type TaskStatus = 'pending' | 'running' | 'success' | 'error' | 'skipped'

/** Task item for task list */
export interface TaskItem {
  name: string
  status: TaskStatus
  message?: string
}

/**
 * Get status symbol for a task
 */
function getStatusSymbol(status: TaskStatus): string {
  switch (status) {
    case 'pending':
      return '○'
    case 'running':
      return '◌'
    case 'success':
      return '✔'
    case 'error':
      return '✖'
    case 'skipped':
      return '◇'
  }
}

/**
 * Format a task list for display
 *
 * Example output:
 * ✔ Created .ai/agents/orchestrator.md
 * ✔ Created .ai/skills/tdd-loop.md
 * ◌ Compiling AGENTS.md...
 */
export function formatTaskList(tasks: TaskItem[]): string {
  return tasks
    .map((task) => {
      const symbol = getStatusSymbol(task.status)
      const msg = task.message ? ` — ${task.message}` : ''
      return `${symbol} ${task.name}${msg}`
    })
    .join('\n')
}

/**
 * File operation type for preview
 */
export type FileOperation = 'create' | 'modify' | 'replace' | 'skip'

/** File operation item */
export interface FileOpItem {
  path: string
  operation: FileOperation
  note?: string
}

/**
 * Get operation prefix with color hint
 */
function getOperationPrefix(op: FileOperation): string {
  switch (op) {
    case 'create':
      return 'CREATE '
    case 'modify':
      return 'MODIFY '
    case 'replace':
      return 'REPLACE'
    case 'skip':
      return 'SKIP   '
  }
}

/**
 * Format file operations preview
 *
 * Example output:
 * CREATE  .ai/agents/orchestrator.md
 * CREATE  .ai/skills/tdd-loop.md
 * MODIFY  .gitignore (add 4 entries)
 * REPLACE CLAUDE.md (will overwrite)
 */
export function formatFileOps(files: FileOpItem[]): string {
  return files
    .map((file) => {
      const prefix = getOperationPrefix(file.operation)
      const note = file.note ? ` (${file.note})` : ''
      return `${prefix} ${file.path}${note}`
    })
    .join('\n')
}

/**
 * Count file operations by type
 */
export function countFileOps(files: FileOpItem[]): { created: number; modified: number; replaced: number; skipped: number } {
  return files.reduce(
    (acc, file) => {
      switch (file.operation) {
        case 'create':
          acc.created++
          break
        case 'modify':
          acc.modified++
          break
        case 'replace':
          acc.replaced++
          break
        case 'skip':
          acc.skipped++
          break
      }
      return acc
    },
    { created: 0, modified: 0, replaced: 0, skipped: 0 }
  )
}

/**
 * Format file count summary
 *
 * Example: "23 new, 2 modified, 1 replaced"
 */
export function formatFileCountSummary(counts: { created: number; modified: number; replaced: number; skipped: number }): string {
  const parts: string[] = []
  if (counts.created > 0) parts.push(`${counts.created} new`)
  if (counts.modified > 0) parts.push(`${counts.modified} modified`)
  if (counts.replaced > 0) parts.push(`${counts.replaced} replaced`)
  if (counts.skipped > 0) parts.push(`${counts.skipped} skipped`)
  return parts.join(', ') || 'no changes'
}
