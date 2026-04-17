import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import {
  loadTeamState,
  loadWorkflowState,
  readMaintenanceContracts,
  readSyncState,
  saveTeamState,
  saveWorkflowState,
  writeSyncState,
} from '../persistence.js'
import type { SyncStateSnapshot, TeamState, WorkflowState } from '../types.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

describe('persistence', () => {
  it('persists team and workflow states', () => {
    const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-persistence-'))
    tempDirs.push(projectRoot)

    const teamState: TeamState = {
      teamId: 'team-1',
      definitionName: 'review-team',
      definitionVersion: '1.0.0',
      state: 'running',
      task: 'Review auth changes',
      tasks: [],
      readyTaskIds: [],
      synthesisTaskId: 'review-team:synthesize',
      budgetPolicy: {
        id: 'team-budget',
        scope: 'team',
        defaultActionOnLimit: 'pause',
      },
      budget: {
        policyId: 'team-budget',
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
      workflowId: 'workflow-1',
      definitionName: 'deliver-feature',
      definitionVersion: '1.0.0',
      state: 'waiting_on_child',
      task: 'Deliver auth changes',
      entryPhaseId: 'implement',
      currentPhaseId: 'implement',
      phases: [],
      childRuns: [],
      budgetPolicy: {
        id: 'workflow-budget',
        scope: 'workflow',
        defaultActionOnLimit: 'pause',
      },
      budget: {
        policyId: 'workflow-budget',
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

    saveTeamState(projectRoot, teamState)
    saveWorkflowState(projectRoot, workflowState)

    expect(loadTeamState(projectRoot, 'team-1')).toEqual(teamState)
    expect(loadWorkflowState(projectRoot, 'workflow-1')).toEqual(workflowState)
  })

  it('persists sync state and filters expired maintenance contracts', () => {
    const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-persistence-housekeeping-'))
    tempDirs.push(projectRoot)

    const syncState: SyncStateSnapshot = {
      schemaVersion: 1,
      updatedAt: '2026-04-17T00:00:00Z',
      qmd: { enabled: true, driftStatus: 'fresh' },
      codegraph: { enabled: false, driftStatus: 'disabled' },
      staleAcked: { qmd: [], codegraph: [] },
      repairProposals: [],
    }

    writeSyncState(projectRoot, syncState)
    fs.mkdirSync(path.join(projectRoot, '.ai', 'housekeeping', 'contracts'), { recursive: true })
    fs.writeFileSync(
      path.join(projectRoot, '.ai', 'housekeeping', 'contracts', 'active.json'),
      JSON.stringify({
        id: 'active',
        approvalScope: 'task_scoped',
        permittedActions: ['memory_write'],
        approvalExpiresAt: '2026-05-01T00:00:00Z',
      }),
    )
    fs.writeFileSync(
      path.join(projectRoot, '.ai', 'housekeeping', 'contracts', 'expired.json'),
      JSON.stringify({
        id: 'expired',
        approvalScope: 'standing',
        permittedActions: ['memory_write'],
        approvalExpiresAt: '2026-04-01T00:00:00Z',
      }),
    )

    expect(readSyncState(projectRoot)).toEqual(syncState)
    expect(readMaintenanceContracts(projectRoot, '2026-04-17T00:00:00Z').map((contract) => contract.id)).toEqual(['active'])
  })
})
