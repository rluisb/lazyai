import { beforeEach, describe, expect, it, vi } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { JobQueue } from '../queue/queue.js'
import { QueueWorker } from '../queue/worker.js'
import type { Db } from '../db/index.js'

let db: Db
let queue: JobQueue

beforeEach(() => {
  db = openDatabase(':memory:')
  runMigrations(db)
  queue = new JobQueue(db)
})

describe('JobQueue', () => {
  it('enqueues and returns a pending job', () => {
    const job = queue.enqueue({ jobType: 'test_job', payload: { x: 1 } })
    expect(job.status).toBe('pending')
    expect(job.jobType).toBe('test_job')
    expect(job.payload).toEqual({ x: 1 })
    expect(job.attempts).toBe(0)
  })

  it('dequeue atomically claims a job', () => {
    queue.enqueue({ jobType: 'test_job' })
    const job = queue.dequeue()
    expect(job).not.toBeNull()
    expect(job?.status).toBe('claimed')
    expect(job?.attempts).toBe(1)
  })

  it('dequeue returns null when queue is empty', () => {
    expect(queue.dequeue()).toBeNull()
  })

  it('complete marks job as completed', () => {
    queue.enqueue({ jobType: 'test_job' })
    const claimed = queue.dequeue()!
    queue.complete(claimed.id)
    const done = queue.getJob(claimed.id)!
    expect(done.status).toBe('completed')
    expect(done.completedAt).toBeDefined()
  })

  it('fail resets to pending when attempts < maxAttempts', () => {
    queue.enqueue({ jobType: 'test_job', maxAttempts: 3 })
    const claimed = queue.dequeue()! // attempts = 1
    queue.fail(claimed.id, { message: 'transient error' })
    const reset = queue.getJob(claimed.id)!
    expect(reset.status).toBe('pending')
  })

  it('fail marks as failed when max_attempts exhausted', () => {
    queue.enqueue({ jobType: 'test_job', maxAttempts: 1 })
    const claimed = queue.dequeue()! // attempts = 1 = max
    queue.fail(claimed.id, { message: 'fatal' })
    const failed = queue.getJob(claimed.id)!
    expect(failed.status).toBe('failed')
  })

  it('reclaim resets stale claimed jobs to pending', () => {
    queue.enqueue({ jobType: 'test_job' })
    queue.dequeue() // claim it

    // Simulate stale claim by back-dating claimed_at
    db.prepare("UPDATE queue_jobs SET claimed_at = datetime('now', '-2 minutes')").run()

    const reclaimed = queue.reclaim(60_000) // 1 minute timeout
    expect(reclaimed).toBe(1)
    expect(queue.pendingCount()).toBe(1)
  })

  it('respects priority ordering', () => {
    queue.enqueue({ jobType: 'test_job', payload: { n: 1 }, priority: 0 })
    queue.enqueue({ jobType: 'test_job', payload: { n: 2 }, priority: 10 })
    queue.enqueue({ jobType: 'test_job', payload: { n: 3 }, priority: 5 })

    const first = queue.dequeue()
    expect(first?.payload).toEqual({ n: 2 }) // highest priority
  })

  it('filters dequeue by jobType', () => {
    queue.enqueue({ jobType: 'alpha' })
    queue.enqueue({ jobType: 'beta' })

    const job = queue.dequeue('beta')
    expect(job?.jobType).toBe('beta')
    expect(queue.pendingCount('alpha')).toBe(1)
  })

  it('listJobs returns jobs filtered by status', () => {
    queue.enqueue({ jobType: 'test_job' })
    queue.enqueue({ jobType: 'test_job' })
    const j = queue.dequeue()!
    queue.complete(j.id)

    const pending = queue.listJobs('pending')
    const completed = queue.listJobs('completed')
    expect(pending).toHaveLength(1)
    expect(completed).toHaveLength(1)
  })
})

describe('QueueWorker', () => {
  it('processes a job via registered handler', async () => {
    queue.enqueue({ jobType: 'greet', payload: { name: 'world' } })

    const processed: string[] = []
    const worker = new QueueWorker({ db, pollIntervalMs: 50 })
    worker.register('greet', async (job) => {
      processed.push((job.payload as { name: string }).name)
    })

    worker.start()
    await vi.waitFor(() => expect(processed).toHaveLength(1), { timeout: 1000 })
    worker.stop()

    expect(processed[0]).toBe('world')
    expect(queue.getJob(queue.listJobs('completed')[0]!.id)?.status).toBe('completed')
  })

  it('retries a job on handler error', async () => {
    queue.enqueue({ jobType: 'flaky', maxAttempts: 2 })

    let calls = 0
    const worker = new QueueWorker({ db, pollIntervalMs: 50 })
    worker.register('flaky', async () => {
      calls++
      if (calls < 2) throw new Error('temporary failure')
    })

    worker.start()
    await vi.waitFor(() => expect(calls).toBeGreaterThanOrEqual(2), { timeout: 2000 })
    worker.stop()

    // After 2 calls the job should be completed (second attempt succeeds)
    const jobs = queue.listJobs('completed')
    expect(jobs).toHaveLength(1)
  })
})
