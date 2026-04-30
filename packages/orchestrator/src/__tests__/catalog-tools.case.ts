import { beforeEach, describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { CatalogToolHandlers } from '../catalog-tools.js'
import type { Db } from '../db/index.js'

let db: Db
let handlers: CatalogToolHandlers

beforeEach(() => {
  db = openDatabase(':memory:')
  runMigrations(db)
  handlers = new CatalogToolHandlers(db)
})

describe('CatalogToolHandlers catalog lifecycle', () => {
  it('deactivates a definition while pinned versions remain readable', () => {
    handlers.catalogCreateVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })

    expect(handlers.catalogDeactivate({ kind: 'agent', name: 'reviewer' })).toEqual({
      kind: 'agent',
      name: 'reviewer',
      activeVersion: null,
      deactivated: true,
    })

    expect(() => handlers.catalogGetVersion({ kind: 'agent', name: 'reviewer' })).toThrow(/No active version/)
    expect(handlers.catalogGetVersion({ kind: 'agent', name: 'reviewer', version: 1 }).body).toBe('v1')
    expect(handlers.catalogList({ kind: 'agent' }).definitions).toMatchObject([
      { kind: 'agent', name: 'reviewer', activeVersion: null, totalVersions: 1 },
    ])
  })

  it('removes a definition and reports removed versions', () => {
    handlers.catalogCreateVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })
    handlers.catalogCreateVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2' })

    expect(handlers.catalogRemove({ kind: 'agent', name: 'reviewer' })).toEqual({
      kind: 'agent',
      name: 'reviewer',
      removed: true,
      versionsRemoved: 2,
    })

    expect(handlers.catalogList({ kind: 'agent' }).definitions).toEqual([])
    expect(handlers.catalogListVersions({ kind: 'agent', name: 'reviewer' }).versions).toEqual([])
    expect(() => handlers.catalogGetVersion({ kind: 'agent', name: 'reviewer' })).toThrow(/No active version/)
    expect(() => handlers.catalogGetVersion({ kind: 'agent', name: 'reviewer', version: 1 })).toThrow(/not found/)
  })
})
