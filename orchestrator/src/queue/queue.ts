import crypto from 'node:crypto'
import type { Db } from '../db/index.js'

export type JobStatus = 'pending' | 'claimed' | 'completed' | 'failed'

export interface Job<P = Record<string, unknown>> {
  id: string
  jobType: string
  payload: P
  status: JobStatus
  priority: number
  attempts: number
  maxAttempts: number
  error?: Record<string, unknown>
  createdAt: string
  claimedAt?: string
  completedAt?: string
}

export interface EnqueueInput {
  jobType: string
  payload?: Record<string, unknown>
  priority?: number
  maxAttempts?: number
  id?: string
}

interface JobRow {
  id: string
  job_type: string
  payload_json: string
  status: string
  priority: number
  attempts: number
  max_attempts: number
  error_json: string | null
  created_at: string
  claimed_at: string | null
  completed_at: string | null
}

function rowToJob(row: JobRow): Job {
  return {
    id: row.id,
    jobType: row.job_type,
    payload: JSON.parse(row.payload_json) as Record<string, unknown>,
    status: row.status as JobStatus,
    priority: row.priority,
    attempts: row.attempts,
    maxAttempts: row.max_attempts,
    ...(row.error_json ? { error: JSON.parse(row.error_json) as Record<string, unknown> } : {}),
    createdAt: row.created_at,
    ...(row.claimed_at ? { claimedAt: row.claimed_at } : {}),
    ...(row.completed_at ? { completedAt: row.completed_at } : {}),
  }
}

export class JobQueue {
  constructor(private readonly db: Db) {}

  enqueue(input: EnqueueInput): Job {
    const id = input.id ?? crypto.randomUUID()
    const now = new Date().toISOString()
    this.db.prepare(`
      INSERT INTO queue_jobs (id, job_type, payload_json, priority, max_attempts, created_at)
      VALUES (?, ?, ?, ?, ?, ?)
    `).run(
      id,
      input.jobType,
      JSON.stringify(input.payload ?? {}),
      input.priority ?? 0,
      input.maxAttempts ?? 3,
      now,
    )
    return this.getJob(id)!
  }

  /** Atomically claim the next pending job. Returns null when queue is empty. */
  dequeue(jobType?: string): Job | null {
    const typeFilter = jobType ? `AND job_type = '${jobType.replace(/'/g, "''")}'` : ''
    const now = new Date().toISOString()

    // SQLite doesn't support RETURNING in older versions; use a transaction to claim atomically.
    const job = this.db.transaction(() => {
      const row = this.db.prepare<[], JobRow>(`
        SELECT * FROM queue_jobs
        WHERE status = 'pending' ${typeFilter}
        ORDER BY priority DESC, created_at ASC
        LIMIT 1
      `).get()

      if (!row) return null

      this.db.prepare(`
        UPDATE queue_jobs
        SET status = 'claimed', claimed_at = ?, attempts = attempts + 1
        WHERE id = ?
      `).run(now, row.id)

      return this.db.prepare<[string], JobRow>('SELECT * FROM queue_jobs WHERE id = ?').get(row.id) ?? null
    })()

    return job ? rowToJob(job) : null
  }

  complete(id: string): void {
    const now = new Date().toISOString()
    this.db.prepare(`
      UPDATE queue_jobs SET status = 'completed', completed_at = ? WHERE id = ?
    `).run(now, id)
  }

  fail(id: string, error: Record<string, unknown>): void {
    const now = new Date().toISOString()
    // If attempts < max_attempts, reset to pending for retry; otherwise mark failed.
    this.db.transaction(() => {
      const row = this.db.prepare<[string], { attempts: number; max_attempts: number }>(
        'SELECT attempts, max_attempts FROM queue_jobs WHERE id = ?',
      ).get(id)

      if (!row) return

      const nextStatus = row.attempts < row.max_attempts ? 'pending' : 'failed'
      this.db.prepare(`
        UPDATE queue_jobs
        SET status = ?, error_json = ?, completed_at = CASE WHEN ? = 'failed' THEN ? ELSE NULL END
        WHERE id = ?
      `).run(nextStatus, JSON.stringify(error), nextStatus, now, id)
    })()
  }

  /** Reclaim jobs that were claimed but never completed (process crash recovery). */
  reclaim(timeoutMs = 60_000): number {
    const cutoff = new Date(Date.now() - timeoutMs).toISOString()
    const result = this.db.prepare(`
      UPDATE queue_jobs
      SET status = 'pending', claimed_at = NULL
      WHERE status = 'claimed' AND claimed_at < ?
    `).run(cutoff)
    return result.changes
  }

  getJob(id: string): Job | null {
    const row = this.db.prepare<[string], JobRow>('SELECT * FROM queue_jobs WHERE id = ?').get(id)
    return row ? rowToJob(row) : null
  }

  listJobs(status?: JobStatus, limit = 50): Job[] {
    const where = status ? 'WHERE status = ?' : ''
    const params: unknown[] = status ? [status] : []
    const rows = this.db.prepare<unknown[], JobRow>(`
      SELECT * FROM queue_jobs ${where} ORDER BY created_at DESC LIMIT ${limit}
    `).all(...params)
    return rows.map(rowToJob)
  }

  pendingCount(jobType?: string): number {
    const where = jobType ? `AND job_type = ?` : ''
    const params: unknown[] = jobType ? [jobType] : []
    const row = this.db.prepare<unknown[], { n: number }>(
      `SELECT COUNT(*) AS n FROM queue_jobs WHERE status = 'pending' ${where}`,
    ).get(...params)
    return row?.n ?? 0
  }
}
