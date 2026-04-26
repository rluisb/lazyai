import fs from 'node:fs'
import path from 'node:path'
import Database from 'better-sqlite3'

export type Db = InstanceType<typeof Database>

export interface OpenDatabaseOptions {
  readonly?: boolean
}

export function openDatabase(dbPath: string, options: OpenDatabaseOptions = {}): Db {
  if (dbPath !== ':memory:') {
    fs.mkdirSync(path.dirname(dbPath), { recursive: true })
  }
  const db = new Database(dbPath, { readonly: options.readonly ?? false })
  db.pragma('journal_mode = WAL')
  db.pragma('foreign_keys = ON')
  db.pragma('busy_timeout = 5000')
  return db
}
