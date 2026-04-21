import type { Db } from '../db/index.js'
import type { Logger } from '../logging/logger.js'
import { createNoopLogger } from '../logging/logger.js'
import { JobQueue } from './queue.js'
import type { Job } from './queue.js'

export type JobHandler = (job: Job) => Promise<void>

export interface WorkerOptions {
  db: Db
  logger?: Logger
  pollIntervalMs?: number
  claimTimeoutMs?: number
}

export class QueueWorker {
  private readonly queue: JobQueue
  private readonly log: Logger
  private readonly pollIntervalMs: number
  private readonly claimTimeoutMs: number
  private readonly handlers = new Map<string, JobHandler>()
  private timer: ReturnType<typeof setInterval> | null = null
  private running = false

  constructor(options: WorkerOptions) {
    this.queue = new JobQueue(options.db)
    this.log = (options.logger ?? createNoopLogger()).child({ component: 'queue-worker' })
    this.pollIntervalMs = options.pollIntervalMs ?? 2_000
    this.claimTimeoutMs = options.claimTimeoutMs ?? 60_000
  }

  register(jobType: string, handler: JobHandler): this {
    this.handlers.set(jobType, handler)
    return this
  }

  start(): void {
    if (this.running) return
    this.running = true
    const reclaimed = this.queue.reclaim(this.claimTimeoutMs)
    if (reclaimed > 0) {
      this.log.info('queue.reclaimed', { count: reclaimed })
    }
    this.timer = setInterval(() => { void this.tick() }, this.pollIntervalMs)
    // unref so the interval doesn't prevent process exit
    this.timer.unref()
  }

  stop(): void {
    if (this.timer) {
      clearInterval(this.timer)
      this.timer = null
    }
    this.running = false
  }

  private async tick(): Promise<void> {
    for (const jobType of this.handlers.keys()) {
      const job = this.queue.dequeue(jobType)
      if (!job) continue

      const handler = this.handlers.get(jobType)
      if (!handler) continue

      this.log.info('queue.job.start', { id: job.id, jobType: job.jobType, attempt: job.attempts })
      try {
        await handler(job)
        this.queue.complete(job.id)
        this.log.info('queue.job.done', { id: job.id })
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err)
        this.log.warn('queue.job.failed', { id: job.id, error: message })
        this.queue.fail(job.id, { message, ...(err instanceof Error && err.stack ? { stack: err.stack } : {}) })
      }
    }
  }
}

// Process-level singleton worker (started by serve / stdio boot)
let _worker: QueueWorker | null = null

export function getQueueWorker(): QueueWorker | null {
  return _worker
}

export function startQueueWorker(options: WorkerOptions): QueueWorker {
  if (_worker) _worker.stop()
  _worker = new QueueWorker(options)
  _worker.start()
  return _worker
}

export function stopQueueWorker(): void {
  _worker?.stop()
  _worker = null
}
