import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { runBootstrap } from '../bootstrap.js'
import type { MaintenanceContractRecord, SyncStateSnapshot } from '../types.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function makeProject(): string {
  const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-bootstrap-'))
  tempDirs.push(projectRoot)
  fs.mkdirSync(path.join(projectRoot, 'specs', 'memory'), { recursive: true })
  return projectRoot
}

function makeSyncState(overrides?: Partial<SyncStateSnapshot>): SyncStateSnapshot {
  return {
    schemaVersion: 1,
    updatedAt: '2026-04-17T00:00:00Z',
    qmd: { enabled: true, driftStatus: 'stale' },
    codegraph: { enabled: true, driftStatus: 'fresh' },
    staleAcked: { qmd: [], codegraph: [] },
    repairProposals: [],
    ...overrides,
  }
}

describe('bootstrap', () => {
  it('continues when sync state is absent', () => {
    const projectRoot = makeProject()
    const report = runBootstrap({ projectRoot })

    expect(report.syncStatePresent).toBe(false)
    expect(report.qmdStatus).toBe('unknown')
    expect(report.contractStatus).toBe('missing')
  })

  it('skips approval requests when an active contract is present', () => {
    const projectRoot = makeProject()
    fs.writeFileSync(path.join(projectRoot, 'specs', 'memory', 'note.md'), '# note\n')

    const contract: MaintenanceContractRecord = {
      id: 'contract-1',
      approvalScope: 'task_scoped',
      permittedActions: ['memory_write', 'qmd_sync'],
      approvalExpiresAt: '2026-05-01T00:00:00Z',
    }

    const report = runBootstrap({
      projectRoot,
      syncState: makeSyncState(),
      contracts: [contract],
      now: '2026-04-17T00:00:00Z',
    })

    expect(report.contractStatus).toBe('active')
    expect(report.approvalsRequested).toEqual([])
    expect(report.contextLoaded).toContain('specs/memory/note.md')
  })

  it('requests fresh approval when the only contract is expired', () => {
    const projectRoot = makeProject()
    const contract: MaintenanceContractRecord = {
      id: 'contract-1',
      approvalScope: 'standing',
      permittedActions: ['memory_write', 'qmd_sync'],
      approvalExpiresAt: '2026-04-01T00:00:00Z',
    }

    const report = runBootstrap({
      projectRoot,
      syncState: makeSyncState(),
      contracts: [contract],
      now: '2026-04-17T00:00:00Z',
    })

    expect(report.contractStatus).toBe('expired')
    expect(report.approvalsRequested).toContain('memory_load')
    expect(report.approvalsRequested).toContain('qmd_sync')
  })
})
