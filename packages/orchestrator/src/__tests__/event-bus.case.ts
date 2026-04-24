import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { EventBus, resetEventBus, getEventBus } from '../events/bus.js'
import type { Db } from '../db/index.js'

let db: Db

beforeEach(() => {
  db = openDatabase(':memory:')
  runMigrations(db)
  resetEventBus()
})

afterEach(() => {
  resetEventBus()
})

describe('EventBus', () => {
  it('delivers event to onAny listener', () => {
    const bus = new EventBus()
    const received: string[] = []
    bus.onAny((e) => received.push(e.eventType))

    bus.emit(null, { eventType: 'chain.started', runId: 'r1', runKind: 'chain', payload: {} })
    expect(received).toEqual(['chain.started'])
  })

  it('delivers event to onRun listener for matching run', () => {
    const bus = new EventBus()
    const received: string[] = []
    bus.onRun('r1', (e) => received.push(e.runId))

    bus.emit(null, { eventType: 'chain.started', runId: 'r1', runKind: 'chain', payload: {} })
    bus.emit(null, { eventType: 'chain.started', runId: 'r2', runKind: 'chain', payload: {} })

    expect(received).toEqual(['r1'])
  })

  it('persists events to DB when db is provided', () => {
    const bus = new EventBus()
    bus.emit(db, { eventType: 'chain.completed', runId: 'r1', runKind: 'chain', payload: { steps: 3 } })

    const rows = db.prepare('SELECT * FROM run_events WHERE run_id = ?').all('r1') as { event_type: string }[]
    expect(rows).toHaveLength(1)
    expect(rows[0]?.event_type).toBe('chain.completed')
  })

  it('replayFromDb returns persisted events in order with id', () => {
    const bus = new EventBus()
    bus.emit(db, { eventType: 'chain.started', runId: 'r1', runKind: 'chain', payload: {} })
    bus.emit(db, { eventType: 'chain.step.done', runId: 'r1', runKind: 'chain', payload: {} })
    bus.emit(db, { eventType: 'chain.completed', runId: 'r1', runKind: 'chain', payload: {} })

    const history = bus.replayFromDb(db, 'r1')
    expect(history.map((e) => e.eventType)).toEqual([
      'chain.started',
      'chain.step.done',
      'chain.completed',
    ])
    expect(history[0]?.id).toBeTypeOf('number')
  })

  it('replayFromDb with sinceEventId returns only events after that id', () => {
    const bus = new EventBus()
    bus.emit(db, { eventType: 'chain.started', runId: 'r1', runKind: 'chain', payload: {} })
    bus.emit(db, { eventType: 'chain.step.done', runId: 'r1', runKind: 'chain', payload: {} })
    bus.emit(db, { eventType: 'chain.completed', runId: 'r1', runKind: 'chain', payload: {} })

    const all = bus.replayFromDb(db, 'r1')
    const firstId = all[0]!.id!

    const tail = bus.replayFromDb(db, 'r1', firstId)
    expect(tail.map((e) => e.eventType)).toEqual(['chain.step.done', 'chain.completed'])
  })

  it('removeRunListeners stops future delivery for that run', () => {
    const bus = new EventBus()
    const received: string[] = []
    bus.onRun('r1', (e) => received.push(e.eventType))

    bus.emit(null, { eventType: 'chain.started', runId: 'r1', runKind: 'chain', payload: {} })
    bus.removeRunListeners('r1')
    bus.emit(null, { eventType: 'chain.completed', runId: 'r1', runKind: 'chain', payload: {} })

    expect(received).toEqual(['chain.started'])
  })

  it('getEventBus returns the same singleton', () => {
    const a = getEventBus()
    const b = getEventBus()
    expect(a).toBe(b)
  })

  it('resetEventBus creates a fresh singleton', () => {
    const a = getEventBus()
    resetEventBus()
    const b = getEventBus()
    expect(a).not.toBe(b)
  })
})
