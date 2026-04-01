import { Low, Memory } from 'lowdb'
import { JSONFile } from 'lowdb/node'
import { join } from 'path'
import { defaultStore, type StoreData, type Operation } from './schema.js'
import { migrate } from './migrations.js'

const STORE_FILE_NAME = '.ai-setup.json'

export async function createStore(targetDir: string): Promise<Low<StoreData>> {
  const file = new JSONFile<StoreData>(join(targetDir, STORE_FILE_NAME))
  const adapter = file
  const db = new Low(adapter, defaultStore())

  await db.read()

  if (db.data === null) {
    db.data = defaultStore()
    await db.write()
  }

  db.data = migrate(targetDir, db.data)
  db.data.meta.lastUpdatedAt = new Date().toISOString()
  db.data.sync.dirty = true

  await db.write()
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
