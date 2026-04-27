import { randomUUID } from 'node:crypto'
import type { Operation } from '../store/schema.js'

type OperationResult = 'success' | 'failure' | 'partial'

export class OperationTracker {
  private succeeded: string[] = []
  private failed: { path: string; error: string }[] = []
  private backups: Map<string, string> = new Map()
  private operationType: string
  private timestamp: string

  constructor(operationType: string) {
    this.operationType = operationType
    this.timestamp = new Date().toISOString()
  }

  trackSuccess(filePath: string): void {
    this.succeeded.push(filePath)
  }

  trackFailure(filePath: string, error: string): void {
    this.failed.push({ path: filePath, error })
  }

  registerBackup(sourcePath: string, backupPath: string): void {
    this.backups.set(sourcePath, backupPath)
  }

  get succeededCount(): number {
    return this.succeeded.length
  }

  get failedCount(): number {
    return this.failed.length
  }

  get result(): OperationResult {
    if (this.failed.length > 0 && this.succeeded.length > 0) return 'partial'
    if (this.failed.length > 0) return 'failure'
    return 'success'
  }

  toOperation(): Operation {
    const id = typeof randomUUID === 'function'
      ? randomUUID()
      : `op_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`

    return {
      id,
      type: this.operationType,
      timestamp: this.timestamp,
      filesAffected: [...this.succeeded, ...this.failed.map((entry) => entry.path)],
      result: this.result,
      backupPaths: Array.from(this.backups.values()),
      ...(this.failed.length > 0
        ? { error: this.failed.map((entry) => `${entry.path}: ${entry.error}`).join('; ') }
        : {}),
    }
  }
}
