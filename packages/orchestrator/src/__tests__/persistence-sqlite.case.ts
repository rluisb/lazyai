import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { importFromFiles } from '../persistence/importer.js'
import {
  initPersistenceDb,
  loadChainState,
  loadErrorJournalEntries,
  loadTeamState,
  loadWorkflowState,
  saveChainState,
  saveErrorJournalEntry,
  saveTeamState,
  saveWorkflowState,
} from '../persistence.js'
import type { ChainState, ErrorJournalEntry, TeamState, WorkflowState } from '../types.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function makeTempDir(prefix: string): string {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), prefix))
  tempDirs.push(dir)
  return dir
}

function makeFreshDb(): void {
  const db = openDatabase(':memory:')
  runMigrations(db)
  initPersistenceDb(':memory:')
}

const ROOT = '/tmp/test-project'

const teamState: TeamState = {
  teamId: 'team-sqlite-1',
  definitionName: 'review-team',
  definitionVersion: '1.0.0',
  state: 'running',
  task: 'Review auth',
  tasks: [],
  readyTaskIds: [],
  synthesisTaskId: 'review-team:synthesize',
  budgetPolicy: { id: 'bp', scope: 'team', defaultActionOnLimit: 'pause' },
  budget: {
    policyId: 'bp',
    scope: 'team',
    tokens: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    costUsd: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    wallClockMs: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    retries: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    byStep: {},
    lastUpdatedAt: '2026-01-01T00:00:00.000Z',
  },
  createdAt: '2026-01-01T00:00:00.000Z',
  updatedAt: '2026-01-01T00:00:00.000Z',
}

const workflowState: WorkflowState = {
  workflowId: 'wf-sqlite-1',
  definitionName: 'deliver-feature',
  definitionVersion: '1.0.0',
  state: 'waiting_on_child',
  task: 'Deliver auth changes',
  entryPhaseId: 'implement',
  currentPhaseId: 'implement',
  phases: [],
  childRuns: [],
  budgetPolicy: { id: 'wfbp', scope: 'workflow', defaultActionOnLimit: 'pause' },
  budget: {
    policyId: 'wfbp',
    scope: 'workflow',
    tokens: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    costUsd: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    wallClockMs: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    retries: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    byStep: {},
    lastUpdatedAt: '2026-01-01T00:00:00.000Z',
  },
  createdAt: '2026-01-01T00:00:00.000Z',
  updatedAt: '2026-01-01T00:00:00.000Z',
}

const chainState: ChainState = {
  chainId: 'chain-sqlite-1',
  definitionName: 'implement-chain',
  definitionVersion: '1.0.0',
  executionPlanId: 'plan-1',
  state: 'running',
  task: 'Implement feature',
  entryStepId: 'plan',
  currentStepId: 'plan',
  steps: [],
  completedStepIds: [],
  budget: {
    policyId: 'cbp',
    scope: 'chain',
    tokens: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    costUsd: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    wallClockMs: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    retries: { consumed: 0, warningTriggered: false, pausedAtLimit: false },
    byStep: {},
    lastUpdatedAt: '2026-01-01T00:00:00.000Z',
  },
  createdAt: '2026-01-01T00:00:00.000Z',
  updatedAt: '2026-01-01T00:00:00.000Z',
}

describe('persistence (SQLite)', () => {
  beforeEach(() => makeFreshDb())

  it('round-trips ChainState through save/load', () => {
    saveChainState(ROOT, chainState)
    expect(loadChainState(ROOT, 'chain-sqlite-1')).toEqual(chainState)
  })

  it('upserts ChainState on second save', () => {
    saveChainState(ROOT, chainState)
    const updated = { ...chainState, state: 'completed' as const, updatedAt: '2026-06-01T00:00:00.000Z' }
    saveChainState(ROOT, updated)
    expect(loadChainState(ROOT, 'chain-sqlite-1').state).toBe('completed')
  })

  it('round-trips TeamState through save/load', () => {
    saveTeamState(ROOT, teamState)
    expect(loadTeamState(ROOT, 'team-sqlite-1')).toEqual(teamState)
  })

  it('round-trips WorkflowState through save/load', () => {
    saveWorkflowState(ROOT, workflowState)
    expect(loadWorkflowState(ROOT, 'wf-sqlite-1')).toEqual(workflowState)
  })

  it('throws when loading a nonexistent record', () => {
    expect(() => loadChainState(ROOT, 'ghost')).toThrow(/not found/)
  })

  it('round-trips ErrorJournalEntry through save/load', () => {
    const entry: ErrorJournalEntry = {
      id: 'ej-1',
      runId: 'chain-sqlite-1',
      runKind: 'chain',
      definitionName: 'implement-chain',
      error: {
        category: 'validation',
        code: 'BAD_OUTPUT',
        message: 'invalid',
        stepId: 'plan',
        agent: 'builder',
        skills: [],
        context: { runId: 'chain-sqlite-1', runKind: 'chain', task: 'x', attempt: 1, hostCli: 'opencode' },
        suggestedRecovery: { type: 'retry', guidance: 'Try again.' },
        timestamp: '2026-01-01T00:00:00.000Z',
      },
    }
    saveErrorJournalEntry(entry)
    const rows = loadErrorJournalEntries(ROOT)
    expect(rows).toHaveLength(1)
    expect(rows[0]?.error.code).toBe('BAD_OUTPUT')
  })

  it('importer reads on-disk JSON and inserts into DB', () => {
    const dir = makeTempDir('orchestrator-import-')
    const stateDir = path.join(dir, '.ai', 'orchestration', 'state')
    fs.mkdirSync(path.join(stateDir, 'teams'), { recursive: true })
    fs.mkdirSync(path.join(stateDir, 'workflows'), { recursive: true })
    fs.mkdirSync(path.join(stateDir, 'chains'), { recursive: true })
    fs.mkdirSync(path.join(stateDir, 'plans'), { recursive: true })
    fs.mkdirSync(path.join(stateDir, 'handoffs'), { recursive: true })

    const fileTeam = { ...teamState, teamId: 'import-team-1' }
    const fileWf = { ...workflowState, workflowId: 'import-wf-1' }
    fs.writeFileSync(path.join(stateDir, 'teams', 'import-team-1.json'), JSON.stringify(fileTeam), 'utf-8')
    fs.writeFileSync(path.join(stateDir, 'workflows', 'import-wf-1.json'), JSON.stringify(fileWf), 'utf-8')

    const result = importFromFiles(dir)
    expect(result.teams).toBe(1)
    expect(result.workflows).toBe(1)
    expect(result.errors).toEqual([])

    expect(loadTeamState(dir, 'import-team-1')).toEqual(fileTeam)
    expect(loadWorkflowState(dir, 'import-wf-1')).toEqual(fileWf)
  })
})
