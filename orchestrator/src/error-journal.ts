import crypto from 'node:crypto'
import fs from 'node:fs'
import type { ErrorCategory, ErrorJournalEntry, RecoveryAction, StructuredError } from './types.js'
import { ensureStateDir, getErrorJournalPath, readJsonLines } from './persistence.js'

export interface CreateStructuredErrorInput {
  category: ErrorCategory
  code: string
  message: string
  stepId: string
  agent: string
  skills: string[]
  context: StructuredError['context']
  suggestedRecovery?: RecoveryAction
}

export function createStructuredError(input: CreateStructuredErrorInput): StructuredError {
  return {
    category: input.category,
    code: input.code,
    message: input.message,
    stepId: input.stepId,
    agent: input.agent,
    skills: input.skills,
    context: input.context,
    suggestedRecovery: input.suggestedRecovery ?? defaultRecoveryForCategory(input.category),
    timestamp: new Date().toISOString(),
  }
}

export function defaultRecoveryForCategory(category: ErrorCategory): RecoveryAction {
  switch (category) {
    case 'transient':
      return { type: 'retry', maxAttempts: 2, guidance: 'Retry the step after the transient failure clears.' }
    case 'logical':
      return { type: 'escalate', targetAgent: 'planner', reason: 'The current approach likely needs a different perspective.' }
    case 'budget':
      return { type: 'pause', reason: 'A configured budget threshold was reached.' }
    case 'permission':
      return { type: 'abort', reason: 'Required permissions or tools were denied by the host.' }
    case 'validation':
      return { type: 'retry', guidance: 'Return the required structured output fields and try again.' }
    case 'fatal':
      return { type: 'handoff', summary: 'Runtime encountered an unrecoverable condition.' }
  }
}

export function appendErrorJournalEntry(projectRoot: string, entry: ErrorJournalEntry): void {
  ensureStateDir(projectRoot)
  fs.appendFileSync(getErrorJournalPath(projectRoot), `${JSON.stringify(entry)}\n`, 'utf-8')
}

export function readErrorJournal(projectRoot: string): ErrorJournalEntry[] {
  return readJsonLines<ErrorJournalEntry>(getErrorJournalPath(projectRoot))
}

export function createErrorJournalEntry(input: Omit<ErrorJournalEntry, 'id'>): ErrorJournalEntry {
  return {
    ...input,
    id: crypto.randomUUID(),
  }
}
