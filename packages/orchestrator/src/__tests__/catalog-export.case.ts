import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { CatalogToolHandlers } from '../catalog-tools.js'
import type { Db } from '../db/index.js'

let db: Db
let handlers: CatalogToolHandlers
const tempDirs: string[] = []

beforeEach(() => {
  db = openDatabase(':memory:')
  runMigrations(db)
  handlers = new CatalogToolHandlers(db)
})

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function tmpDir(): string {
  const d = fs.mkdtempSync(path.join(os.tmpdir(), 'catalog-export-'))
  tempDirs.push(d)
  return d
}

describe('catalogExportVersion', () => {
  it('writes the active version body to the target path', () => {
    handlers.catalogCreateVersion({
      kind: 'agent',
      name: 'reviewer',
      frontmatter: { name: 'Reviewer' },
      body: '# Reviewer\n\nReview code.',
    })

    const targetPath = path.join(tmpDir(), 'reviewer.md')
    const result = handlers.catalogExportVersion({ kind: 'agent', name: 'reviewer', targetPath })

    expect(result.targetPath).toBe(targetPath)
    expect(result.version).toBe(1)
    expect(fs.readFileSync(targetPath, 'utf-8')).toBe('# Reviewer\n\nReview code.')
  })

  it('writes a pinned version body when version is specified', () => {
    handlers.catalogCreateVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1 body' })
    handlers.catalogCreateVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2 body' })

    const targetPath = path.join(tmpDir(), 'reviewer-v1.md')
    const result = handlers.catalogExportVersion({ kind: 'agent', name: 'reviewer', targetPath, version: 1 })

    expect(result.version).toBe(1)
    expect(fs.readFileSync(targetPath, 'utf-8')).toBe('v1 body')
  })

  it('creates intermediate directories', () => {
    handlers.catalogCreateVersion({ kind: 'skill', name: 'ts', frontmatter: { name: 'TS', description: 'TypeScript skill' }, body: 'TS skill body' })

    const targetPath = path.join(tmpDir(), 'nested', 'dir', 'ts.md')
    handlers.catalogExportVersion({ kind: 'skill', name: 'ts', targetPath })

    expect(fs.existsSync(targetPath)).toBe(true)
  })

  it('throws when the definition has no active version', () => {
    expect(() => handlers.catalogExportVersion({
      kind: 'agent',
      name: 'nobody',
      targetPath: path.join(tmpDir(), 'nobody.md'),
    })).toThrow(/No active version/)
  })
})
