import type { Db } from './index.js'
import type { Migration } from './types.js'
import { migrations as defaultMigrations } from './migrations/index.js'

export interface MigrationResult {
  applied: string[]
  skipped: string[]
}

export function runMigrations(db: Db, list: Migration[] = defaultMigrations): MigrationResult {
  db.exec(`
    CREATE TABLE IF NOT EXISTS schema_migrations (
      id TEXT PRIMARY KEY,
      applied_at TEXT NOT NULL
    )
  `)

  const applied: string[] = []
  const skipped: string[] = []
  const isApplied = db.prepare('SELECT 1 AS present FROM schema_migrations WHERE id = ?')
  const recordApplied = db.prepare('INSERT INTO schema_migrations (id, applied_at) VALUES (?, ?)')

  const tx = db.transaction((items: Migration[]) => {
    for (const m of items) {
      if (isApplied.get(m.id)) {
        skipped.push(m.id)
        continue
      }
      const sql = m.sql.trim()
      if (sql.length > 0) db.exec(sql)
      recordApplied.run(m.id, new Date().toISOString())
      applied.push(m.id)
    }
  })

  tx(list)
  return { applied, skipped }
}
