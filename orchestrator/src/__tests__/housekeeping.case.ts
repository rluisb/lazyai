import { describe, expect, it } from 'vitest'
import {
  combineHousekeepingReports,
  runInlineMemoryExtraction,
  runPostTaskHousekeeping,
  runPreTaskHousekeeping,
} from '../housekeeping.js'
import type { MaintenanceContractRecord, SyncStateSnapshot } from '../types.js'

function syncState(status: 'fresh' | 'stale' | 'stale_acked'): SyncStateSnapshot {
  return {
    schemaVersion: 1,
    updatedAt: '2026-04-17T00:00:00Z',
    qmd: { enabled: true, driftStatus: status },
    codegraph: { enabled: true, driftStatus: 'fresh' },
    staleAcked: { qmd: [], codegraph: [] },
    repairProposals: [],
  }
}

describe('housekeeping', () => {
  it('skips approval when an active contract is present', () => {
    const contract: MaintenanceContractRecord = {
      id: 'contract-1',
      approvalScope: 'task_scoped',
      permittedActions: ['memory_write', 'qmd_sync'],
      approvalExpiresAt: '2026-05-01T00:00:00Z',
    }

    const report = runPreTaskHousekeeping({
      syncState: syncState('stale'),
      contracts: [contract],
      now: '2026-04-17T00:00:00Z',
    })

    expect(report.contractStatus).toBe('active')
    expect(report.approvalsRequested).toEqual([])
  })

  it('stages inline memory entries when no write contract exists', () => {
    const report = runInlineMemoryExtraction({
      stepId: 'implement-fix',
      stepOutput: { summary: 'Found a flaky timeout' },
    })

    expect(report.stagedMemoryEntries).toEqual(['implement-fix:Found a flaky timeout'])
    expect(report.approvalsRequested).toEqual(['memory_write'])
  })

  it('merges staged entries into post-task housekeeping and preserves stale acknowledgements', () => {
    const inline = runInlineMemoryExtraction({
      stepId: 'implement-fix',
      stepOutput: { summary: 'Document the timeout lesson' },
    })
    const post = runPostTaskHousekeeping({
      syncState: syncState('stale_acked'),
      stagedMemoryEntries: inline.stagedMemoryEntries,
    })
    const combined = combineHousekeepingReports(inline, post)

    expect(post.deferredMaintenance).toContain('qmd:stale_acked')
    expect(combined.stagedMemoryEntries).toEqual(['implement-fix:Document the timeout lesson'])
  })
})
