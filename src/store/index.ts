import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { Low, Memory } from 'lowdb'
import { JSONFile } from 'lowdb/node'
import { Errors } from '../errors/index.js'
import { migrate } from './migrations.js'
import { defaultStore, type Operation, type StoreData } from './schema.js'

const STORE_FILE_NAME = '.ai-setup.json'

export async function createStore(targetDir: string): Promise<Low<StoreData>> {
  const file = new JSONFile<StoreData>(join(targetDir, STORE_FILE_NAME))
  const db = new Low(file, defaultStore())

  await db.read()

  const wasNew = db.data === null
  if (wasNew) {
    db.data = defaultStore()
    await db.write()
  }

  const beforeMigrate = db.data
  const migrated = migrate(targetDir, db.data)
  const needsMigration = migrated !== beforeMigrate

  db.data = migrated

  if (wasNew || needsMigration) {
    db.data.meta.lastUpdatedAt = new Date().toISOString()
    db.data.sync.dirty = true
    await db.write()
  }

  return db
}

export async function createTestStore(data?: Partial<StoreData>): Promise<Low<StoreData>> {
  const adapter = new Memory<StoreData>()
  const db = new Low(adapter, {
    ...defaultStore(),
    ...data,
  })

  return db
}

export async function readStore(targetDir: string): Promise<StoreData> {
  const db = await createStore(targetDir)
  return db.data
}

export async function readStoreReadonly(targetDir: string): Promise<StoreData> {
  const storePath = join(targetDir, STORE_FILE_NAME)

  if (!existsSync(storePath)) {
    throw Errors.manifestNotFound(targetDir)
  }

  try {
    const raw = JSON.parse(readFileSync(storePath, 'utf-8')) as unknown
    return migrate(targetDir, raw)
  } catch (error) {
    throw Errors.manifestCorrupt(targetDir, error instanceof Error ? error : new Error(String(error)))
  }
}

export async function writeStore(targetDir: string, data: StoreData): Promise<void> {
  const db = await createStore(targetDir)
  db.data = data
  db.data.meta.lastUpdatedAt = new Date().toISOString()
  db.data.sync.dirty = false
  await db.write()
}

export async function appendOperation(targetDir: string, operation: Operation): Promise<void> {
  const db = await createStore(targetDir)
  db.data.operations.push(operation)

  if (db.data.operations.length > 50) {
    db.data.operations = db.data.operations.slice(-50)
  }

  db.data.meta.lastUpdatedAt = new Date().toISOString()
  await db.write()
}
