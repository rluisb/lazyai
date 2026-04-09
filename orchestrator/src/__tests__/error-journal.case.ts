import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import {
  appendErrorJournalEntry,
  createErrorJournalEntry,
  createStructuredError,
  readErrorJournal,
} from '../error-journal.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

describe('error-journal', () => {
  it('appends and reads journal entries', () => {
    const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-errors-'))
    tempDirs.push(projectRoot)

    const error = createStructuredError({
      category: 'validation',
      code: 'INVALID_OUTPUT',
      message: 'bad output',
      stepId: 'step-1',
      agent: 'builder',
      skills: ['typescript'],
      context: {
        runId: 'run-1',
        runKind: 'chain',
        task: 'repair',
        attempt: 1,
        hostCli: 'opencode',
      },
    })

    appendErrorJournalEntry(
      projectRoot,
      createErrorJournalEntry({
        runId: 'run-1',
        runKind: 'chain',
        definitionName: 'repair-chain',
        stepId: 'step-1',
        error,
      }),
    )

    const entries = readErrorJournal(projectRoot)
    expect(entries).toHaveLength(1)
    expect(entries[0]?.error.code).toBe('INVALID_OUTPUT')
    expect(entries[0]?.error.suggestedRecovery.type).toBe('retry')
  })
})
