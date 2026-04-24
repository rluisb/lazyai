import { EventEmitter } from 'node:events'
import type { Db } from '../db/index.js'
import type { RunKind } from '../types.js'

export type { RunKind }

export interface RunEvent {
  id?: number
  eventType: string
  runId: string
  runKind: RunKind
  payload: Record<string, unknown>
  emittedAt: string
}

export type RunEventListener = (event: RunEvent) => void

// One singleton bus per process — the HTTP server and all tool handlers share it.
let _bus: EventBus | null = null

export function getEventBus(): EventBus {
  if (!_bus) _bus = new EventBus()
  return _bus
}

export function resetEventBus(): void {
  _bus?.removeAllListeners()
  _bus = null
}

export class EventBus {
  private readonly emitter = new EventEmitter()

  constructor() {
    // Prevent Node warning when many SSE clients subscribe
    this.emitter.setMaxListeners(256)
  }

  emit(db: Db | null, event: Omit<RunEvent, 'emittedAt'>): void {
    const full: RunEvent = { ...event, emittedAt: new Date().toISOString() }

    if (db) {
      try {
        db.prepare(
          `INSERT INTO run_events (event_type, run_id, run_kind, payload_json, emitted_at)
           VALUES (?, ?, ?, ?, ?)`,
        ).run(full.eventType, full.runId, full.runKind, JSON.stringify(full.payload), full.emittedAt)
      } catch { /* non-fatal — bus still fires in-memory */ }
    }

    this.emitter.emit('run_event', full)
    this.emitter.emit(`run:${full.runId}`, full)
  }

  /** Subscribe to all run events. */
  onAny(listener: RunEventListener): () => void {
    this.emitter.on('run_event', listener)
    return () => this.emitter.off('run_event', listener)
  }

  /** Subscribe to events for a specific run ID. */
  onRun(runId: string, listener: RunEventListener): () => void {
    const key = `run:${runId}`
    this.emitter.on(key, listener)
    return () => this.emitter.off(key, listener)
  }

  removeRunListeners(runId: string): void {
    this.emitter.removeAllListeners(`run:${runId}`)
  }

  removeAllListeners(): void {
    this.emitter.removeAllListeners()
  }

  /** Replay persisted events for a run from DB (for late-joining subscribers).
   *  Pass sinceEventId to replay only events with id > sinceEventId. */
  replayFromDb(db: Db, runId: string, sinceEventId?: number): RunEvent[] {
    const sinceClause = sinceEventId !== undefined ? 'AND id > ?' : ''
    const params: unknown[] = sinceEventId !== undefined ? [runId, sinceEventId] : [runId]

    const rows = db.prepare<unknown[], {
      id: number; event_type: string; run_id: string; run_kind: string; payload_json: string; emitted_at: string
    }>(
      `SELECT id, event_type, run_id, run_kind, payload_json, emitted_at
       FROM run_events WHERE run_id = ? ${sinceClause} ORDER BY id ASC`,
    ).all(...params)

    return rows.map((r) => ({
      id: r.id,
      eventType: r.event_type,
      runId: r.run_id,
      runKind: r.run_kind as RunKind,
      payload: JSON.parse(r.payload_json) as Record<string, unknown>,
      emittedAt: r.emitted_at,
    }))
  }
}
