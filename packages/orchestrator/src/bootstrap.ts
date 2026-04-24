import fs from 'node:fs'
import path from 'node:path'
import type {
  BootstrapReport,
  DriftStatus,
  MaintenanceContractRecord,
  MaintenanceApprovalScope,
  SyncStateSnapshot,
} from './types.js'

export interface RunBootstrapInput {
  projectRoot: string
  memoryPath?: string
  syncState?: SyncStateSnapshot | null
  contracts?: MaintenanceContractRecord[]
  now?: string
}

export function runBootstrap(input: RunBootstrapInput): BootstrapReport {
  const memoryPath = input.memoryPath ?? 'specs/memory'
  const specsAvailable = fs.existsSync(path.join(input.projectRoot, 'specs'))
  const memoryRoot = path.join(input.projectRoot, memoryPath)
  const memoryFiles = listMarkdownFiles(memoryRoot)
  const contract = selectRelevantContract(input.contracts ?? [], input.now)
  const qmdStatus = input.syncState?.qmd.driftStatus ?? 'unknown'
  const codegraphStatus = input.syncState?.codegraph.driftStatus ?? 'unknown'
  const deferredMaintenance = buildDeferredMaintenance(qmdStatus, codegraphStatus)
  const approvalsRequested: string[] = []
  const warnings: string[] = []

  if (!fs.existsSync(memoryRoot)) {
    warnings.push(`Memory path not found: ${memoryPath}`)
  }

  if (contract.status !== 'active') {
    approvalsRequested.push('memory_load')
    if (isMutatingDrift(qmdStatus)) approvalsRequested.push('qmd_sync')
    if (isMutatingDrift(codegraphStatus)) approvalsRequested.push('codegraph_sync')
  }

  return {
    memoryPath,
    specsAvailable,
    syncStatePresent: Boolean(input.syncState),
    qmdStatus,
    codegraphStatus,
    contractStatus: contract.status,
    ...(contract.scope ? { contractScope: contract.scope } : {}),
    approvalsRequested,
    contextLoaded: memoryFiles.map((file) => `${memoryPath}/${file}`),
    deferredMaintenance,
    warnings,
  }
}

export function selectRelevantContract(
  contracts: MaintenanceContractRecord[],
  now = new Date().toISOString(),
): { status: 'active' | 'expired' | 'missing'; scope?: MaintenanceApprovalScope; contract?: MaintenanceContractRecord } {
  if (contracts.length === 0) return { status: 'missing' }

  const active = contracts.find((contract) => isContractActive(contract, now))
  if (active) {
    return {
      status: 'active',
      scope: active.approvalScope,
      contract: active,
    }
  }

  return { status: 'expired' }
}

function listMarkdownFiles(dirPath: string): string[] {
  if (!fs.existsSync(dirPath)) return []
  return fs.readdirSync(dirPath).filter((entry) => entry.endsWith('.md')).sort()
}

function buildDeferredMaintenance(qmdStatus: DriftStatus, codegraphStatus: DriftStatus): string[] {
  const deferred: string[] = []
  if (isMutatingDrift(qmdStatus)) deferred.push(`qmd:${qmdStatus}`)
  if (isMutatingDrift(codegraphStatus)) deferred.push(`codegraph:${codegraphStatus}`)
  return deferred
}

function isMutatingDrift(status: DriftStatus): boolean {
  return status === 'stale' || status === 'stale_acked'
}

function isContractActive(contract: MaintenanceContractRecord, now: string): boolean {
  if (contract.status && contract.status !== 'active') return false
  if (!contract.approvalExpiresAt) return true
  return contract.approvalExpiresAt >= now
}
