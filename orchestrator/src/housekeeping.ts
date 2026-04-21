import type {
  DriftStatus,
  HousekeepingReport,
  MaintenanceContractRecord,
  SyncStateSnapshot,
} from './types.js'
import { selectRelevantContract } from './bootstrap.js'

export interface RunPreTaskHousekeepingInput {
  syncState?: SyncStateSnapshot | null
  contracts?: MaintenanceContractRecord[]
  now?: string
}

export interface RunInlineMemoryExtractionInput {
  stepId: string
  stepOutput?: Record<string, unknown>
  contracts?: MaintenanceContractRecord[]
  now?: string
}

export interface RunPostTaskHousekeepingInput {
  syncState?: SyncStateSnapshot | null
  contracts?: MaintenanceContractRecord[]
  stagedMemoryEntries?: string[]
  now?: string
}

export function runPreTaskHousekeeping(input: RunPreTaskHousekeepingInput): HousekeepingReport {
  const contract = selectRelevantContract(input.contracts ?? [], input.now)
  const qmdStatus = input.syncState?.qmd.driftStatus ?? 'unknown'
  const codegraphStatus = input.syncState?.codegraph.driftStatus ?? 'unknown'
  const approvalsRequested: string[] = []

  if (contract.status !== 'active') {
    approvalsRequested.push('memory_load')
    if (needsApproval(qmdStatus)) approvalsRequested.push('qmd_sync')
    if (needsApproval(codegraphStatus)) approvalsRequested.push('codegraph_sync')
  }

  return {
    phase: 'pre',
    contractStatus: contract.status,
    approvalsRequested,
    contextLoaded: [],
    stagedMemoryEntries: [],
    deferredMaintenance: deferredMaintenance(qmdStatus, codegraphStatus),
    qmdStatus,
    codegraphStatus,
    warnings: [],
  }
}

export function runInlineMemoryExtraction(input: RunInlineMemoryExtractionInput): HousekeepingReport {
  const contract = selectRelevantContract(input.contracts ?? [], input.now)
  const entry = buildInlineEntry(input.stepId, input.stepOutput)
  const canWrite = contract.status === 'active' && contract.contract?.permittedActions.includes('memory_write')

  return {
    phase: 'inline',
    contractStatus: contract.status,
    approvalsRequested: entry && !canWrite ? ['memory_write'] : [],
    contextLoaded: entry && canWrite ? [`memory:${entry}`] : [],
    stagedMemoryEntries: entry && !canWrite ? [entry] : [],
    deferredMaintenance: [],
    qmdStatus: 'unknown',
    codegraphStatus: 'unknown',
    warnings: [],
  }
}

export function runPostTaskHousekeeping(input: RunPostTaskHousekeepingInput): HousekeepingReport {
  const contract = selectRelevantContract(input.contracts ?? [], input.now)
  const qmdStatus = input.syncState?.qmd.driftStatus ?? 'unknown'
  const codegraphStatus = input.syncState?.codegraph.driftStatus ?? 'unknown'
  const stagedMemoryEntries = input.stagedMemoryEntries ?? []
  const approvalsRequested: string[] = []
  const deferred = deferredMaintenance(qmdStatus, codegraphStatus)
  const canWriteMemory = contract.status === 'active' && contract.contract?.permittedActions.includes('memory_write')

  if (stagedMemoryEntries.length > 0 && !canWriteMemory) {
    approvalsRequested.push('memory_write')
    deferred.push('memory_write_pending')
  }
  if (contract.status !== 'active') {
    if (needsApproval(qmdStatus)) approvalsRequested.push('qmd_sync')
    if (needsApproval(codegraphStatus)) approvalsRequested.push('codegraph_sync')
  }

  return {
    phase: 'post',
    contractStatus: contract.status,
    approvalsRequested: unique(approvalsRequested),
    contextLoaded: [],
    stagedMemoryEntries,
    deferredMaintenance: unique(deferred),
    qmdStatus,
    codegraphStatus,
    warnings: [],
  }
}

export function combineHousekeepingReports(...reports: HousekeepingReport[]): HousekeepingReport {
  return {
    phase: 'combined',
    contractStatus: reports.find((report) => report.contractStatus === 'active')?.contractStatus
      ?? reports.at(-1)?.contractStatus
      ?? 'missing',
    approvalsRequested: unique(reports.flatMap((report) => report.approvalsRequested)),
    contextLoaded: unique(reports.flatMap((report) => report.contextLoaded)),
    stagedMemoryEntries: unique(reports.flatMap((report) => report.stagedMemoryEntries)),
    deferredMaintenance: unique(reports.flatMap((report) => report.deferredMaintenance)),
    qmdStatus: reports.find((report) => report.qmdStatus !== 'unknown')?.qmdStatus ?? 'unknown',
    codegraphStatus: reports.find((report) => report.codegraphStatus !== 'unknown')?.codegraphStatus ?? 'unknown',
    warnings: unique(reports.flatMap((report) => report.warnings)),
  }
}

function buildInlineEntry(stepId: string, stepOutput?: Record<string, unknown>): string | null {
  if (!stepOutput) return null
  const summary = typeof stepOutput.summary === 'string' ? stepOutput.summary.trim() : ''
  if (!summary) return null
  return `${stepId}:${summary}`
}

function deferredMaintenance(qmdStatus: DriftStatus, codegraphStatus: DriftStatus): string[] {
  const deferred: string[] = []
  if (needsApproval(qmdStatus)) deferred.push(`qmd:${qmdStatus}`)
  if (needsApproval(codegraphStatus)) deferred.push(`codegraph:${codegraphStatus}`)
  return deferred
}

function needsApproval(status: DriftStatus): boolean {
  return status === 'stale' || status === 'stale_acked'
}

function unique(values: string[]): string[] {
  return [...new Set(values)]
}
