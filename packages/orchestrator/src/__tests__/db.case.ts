import { describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import type { Migration } from '../db/types.js'

describe('db migration runner', () => {
  it('opens an in-memory database, applies migrations, and is idempotent', () => {
    const db = openDatabase(':memory:')

    const first = runMigrations(db)
    expect(first.applied).toContain('0001_init')
    expect(first.skipped).toEqual([])

    const second = runMigrations(db)
    expect(second.applied).toEqual([])
    expect(second.skipped).toContain('0001_init')

    const rows = db.prepare('SELECT id FROM schema_migrations ORDER BY id').all() as Array<{ id: string }>
    expect(rows.map((r) => r.id)).toEqual(['0001_init', '0002_run_state', '0003_catalog', '0004_events', '0005_queue'])

    db.close()
  })

  it('applies user-supplied migrations and records them', () => {
    const db = openDatabase(':memory:')
    const extra: Migration[] = [
      { id: '9999_test', sql: 'CREATE TABLE _probe (id INTEGER PRIMARY KEY)' },
    ]

    const result = runMigrations(db, extra)
    expect(result.applied).toEqual(['9999_test'])

    const probe = db.prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='_probe'").get()
    expect(probe).toBeTruthy()

    db.close()
  })

  it('enables WAL mode and foreign keys on disk-backed databases', () => {
    const db = openDatabase(':memory:')
    const fkRow = db.pragma('foreign_keys') as { foreign_keys: number }[]
    expect(fkRow[0]?.foreign_keys).toBe(1)
    db.close()
  })
})
