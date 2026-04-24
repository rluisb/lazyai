import fs from 'node:fs'
import path from 'node:path'
import type { ChainState, ErrorJournalEntry, ExecutionPlan, HandoffDocument, TeamState, WorkflowState } from '../types.js'
import {
  getStateRoot,
  readJsonLines,
  saveChainState,
  saveErrorJournalEntry,
  saveExecutionPlan,
  saveHandoff,
  saveTeamState,
  saveWorkflowState,
} from '../persistence.js'

export interface ImportResult {
  chains: number
  teams: number
  workflows: number
  plans: number
  handoffs: number
  journalEntries: number
  errors: Array<{ file: string; error: string }>
}

function readJsonFile<T>(filePath: string): T | null {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf-8')) as T
  } catch {
    return null
  }
}

function importJsonDir<T>(
  dir: string,
  save: (item: T) => void,
  errs: ImportResult['errors'],
): number {
  if (!fs.existsSync(dir)) return 0
  let count = 0
  for (const entry of fs.readdirSync(dir)) {
    if (!entry.endsWith('.json')) continue
    const filePath = path.join(dir, entry)
    const item = readJsonFile<T>(filePath)
    if (!item) {
      errs.push({ file: filePath, error: 'JSON parse failed' })
      continue
    }
    try {
      save(item)
      count++
    } catch (err) {
      errs.push({ file: filePath, error: err instanceof Error ? err.message : String(err) })
    }
  }
  return count
}

export function importFromFiles(projectRoot: string): ImportResult {
  const stateRoot = getStateRoot(projectRoot)
  const errors: ImportResult['errors'] = []

  const chains = importJsonDir<ChainState>(
    path.join(stateRoot, 'chains'),
    (s) => saveChainState(projectRoot, s),
    errors,
  )

  const teams = importJsonDir<TeamState>(
    path.join(stateRoot, 'teams'),
    (s) => saveTeamState(projectRoot, s),
    errors,
  )

  const workflows = importJsonDir<WorkflowState>(
    path.join(stateRoot, 'workflows'),
    (s) => saveWorkflowState(projectRoot, s),
    errors,
  )

  const plans = importJsonDir<ExecutionPlan>(
    path.join(stateRoot, 'plans'),
    (p) => saveExecutionPlan(projectRoot, p),
    errors,
  )

  const handoffs = importJsonDir<HandoffDocument>(
    path.join(stateRoot, 'handoffs'),
    (h) => saveHandoff(projectRoot, h),
    errors,
  )

  const journalPath = path.join(stateRoot, 'error-journal.jsonl')
  const journalEntries = readJsonLines<ErrorJournalEntry>(journalPath)
  let journalCount = 0
  for (const entry of journalEntries) {
    try {
      saveErrorJournalEntry(entry)
      journalCount++
    } catch (err) {
      errors.push({ file: journalPath, error: err instanceof Error ? err.message : String(err) })
    }
  }

  return { chains, teams, workflows, plans, handoffs, journalEntries: journalCount, errors }
}
